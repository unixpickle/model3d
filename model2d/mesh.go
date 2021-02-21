// Generated from templates/mesh.template

package model2d

import (
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
)

// A Mesh is a collection of segments.
//
// The segments are uniquely identified as pointers, not
// as values. This is important for methods which
// reference existing segments, such as Remove and
// Neighbors.
//
// Segments in a mesh are "connected" when they contain
// exactly identical points. Thus, small rounding errors
// can cause segments to incorrectly be disassociated
// with each other.
//
// A Mesh can be read safely from concurrent Goroutines,
// but modifications must not be performed concurrently
// with any mesh operations.
type Mesh struct {
	faces map[*Segment]bool

	// Stores a map[Coord][]*Segment
	vertexToFace  atomic.Value
	v2fCreateLock sync.Mutex
}

// NewMesh creates an empty mesh.
func NewMesh() *Mesh {
	return &Mesh{
		faces: map[*Segment]bool{},
	}
}

// NewMeshSegments creates a mesh with the given
// collection of faces.
func NewMeshSegments(faces []*Segment) *Mesh {
	m := NewMesh()
	for _, f := range faces {
		m.Add(f)
	}
	return m
}

// NewMeshPolar creates a closed polar mesh.
//
// The mesh will have correct normals if the radius
// function returns positive values when theta is in the
// range [0, 2*pi].
//
// Even if the polar function does not reach its original
// value at 2*pi radians, the mesh will be closed by
// connecting the first point to the last.
func NewMeshPolar(radius func(theta float64) float64, stops int) *Mesh {
	getPoint := func(t int) Coord {
		theta := float64(t) * math.Pi * 2 / float64(stops)
		return NewCoordPolar(theta, radius(theta))
	}

	firstPoint := getPoint(0)
	lastPoint := firstPoint

	res := NewMesh()
	for i := 1; i < stops; i++ {
		p := getPoint(i)
		res.Add(&Segment{p, lastPoint})
		lastPoint = p
	}
	res.Add(&Segment{firstPoint, lastPoint})
	return res
}

// Add adds the face f to the mesh.
func (m *Mesh) Add(f *Segment) {
	v2f := m.getVertexToFaceOrNil()
	if v2f == nil {
		m.faces[f] = true
		return
	} else if m.faces[f] {
		return
	}

	for _, p := range f {
		v2f[p] = append(v2f[p], f)
	}
	m.faces[f] = true
}

// AddMesh adds all the faces from m1 to m.
func (m *Mesh) AddMesh(m1 *Mesh) {
	m1.Iterate(m.Add)
}

// Remove removes the face f from the mesh.
//
// It looks at f as a pointer, so the pointer must be
// exactly the same as a face passed to Add.
func (m *Mesh) Remove(f *Segment) {
	if !m.faces[f] {
		return
	}
	delete(m.faces, f)
	v2f := m.getVertexToFaceOrNil()
	if v2f != nil {
		for _, p := range f {
			m.removeFaceFromVertex(v2f, f, p)
		}
	}
}

func (m *Mesh) removeFaceFromVertex(v2f map[Coord][]*Segment, f *Segment, p Coord) {
	s := v2f[p]
	for i, f1 := range s {
		if f1 == f {
			essentials.UnorderedDelete(&s, i)
			break
		}
	}
	if len(s) == 0 {
		delete(v2f, p)
	} else {
		v2f[p] = s
	}
}

// Contains checks if f has been added to the mesh.
func (m *Mesh) Contains(f *Segment) bool {
	_, ok := m.faces[f]
	return ok
}

// Iterate calls f for every face in m in an arbitrary
// order.
//
// If f adds or removes faces, they will not be visited.
func (m *Mesh) Iterate(f func(*Segment)) {
	m.IterateSorted(f, nil)
}

// IterateSorted is like Iterate, but it first sorts all
// the faces according to a less than function, cmp.
func (m *Mesh) IterateSorted(f func(*Segment), cmp func(f1, f2 *Segment) bool) {
	all := m.SegmentSlice()
	if cmp != nil {
		sort.Slice(all, func(i, j int) bool {
			return cmp(all[i], all[j])
		})
	}
	for _, face := range all {
		if m.faces[face] {
			f(face)
		}
	}
}

// IterateVertices calls f for every vertex in m in an
// arbitrary order.
//
// If f adds or removes vertices, they will not be
// visited.
func (m *Mesh) IterateVertices(f func(c Coord)) {
	v2f := m.getVertexToFace()
	for _, c := range m.VertexSlice() {
		if _, ok := v2f[c]; ok {
			f(c)
		}
	}
}

// Neighbors gets all the faces with a side touching a
// given face f.
//
// The face f itself is not included in the results.
//
// The face f needn't be in the mesh. However, if it is
// not in the mesh, but an equivalent face is, then said
// equivalent face will be in the results.
func (m *Mesh) Neighbors(f *Segment) []*Segment {
	neighbors := map[*Segment]bool{}
	for _, p := range f {
		for _, n := range m.Find(p) {
			if n != f {
				neighbors[n] = true
			}
		}
	}
	res := make([]*Segment, 0, len(neighbors))
	for s1 := range neighbors {
		res = append(res, s1)
	}
	return res
}

func (m *Mesh) neighborsWithCounts(t *Segment) map[*Segment]int {
	counts := map[*Segment]int{}
	for _, p := range t {
		for _, t1 := range m.getVertexToFace()[p] {
			if t1 != t {
				counts[t1]++
			}
		}
	}
	return counts
}

// Find gets all the faces that contain all of the passed
// points.
//
// For example, to find all faces containing a line from
// from p1 to p2, you could do m.Find(p1, p2).
func (m *Mesh) Find(ps ...Coord) []*Segment {
	if len(ps) == 1 {
		return append([]*Segment{}, m.getVertexToFace()[ps[0]]...)
	}

	faces := m.getVertexToFace()[ps[0]]
	res := make([]*Segment, 0, len(faces))

FaceLoop:
	for _, t := range faces {
		for _, p := range ps[1:] {
			if p != t[0] && p != t[1] {
				continue FaceLoop
			}
		}
		res = append(res, t)
	}

	return res
}

// Scale creates a new mesh by scaling the coordinates by
// a factor s.
func (m *Mesh) Scale(s float64) *Mesh {
	return m.MapCoords(XY(s, s).Mul)
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord) Coord) *Mesh {
	mapping := map[Coord]Coord{}
	if v2f := m.getVertexToFaceOrNil(); v2f != nil {
		for c := range v2f {
			mapping[c] = f(c)
		}
	} else {
		for t := range m.faces {
			for _, c := range t {
				if _, ok := mapping[c]; !ok {
					mapping[c] = f(c)
				}
			}
		}
	}
	m1 := NewMesh()
	m.Iterate(func(t *Segment) {
		t1 := *t
		for i, p := range t {
			t1[i] = mapping[p]
		}
		m1.Add(&t1)
	})
	return m1
}

// Transform applies t to the coordinates.
func (m *Mesh) Transform(t Transform) *Mesh {
	return m.MapCoords(t.Apply)
}

// SaveSVG encodes the mesh to an SVG file.
func (m *Mesh) SaveSVG(path string) error {
	data := EncodeSVG(m)
	w, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "save SVG")
	}
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return errors.Wrap(err, "save SVG")
	}
	return nil
}

// SegmentSlice gets a snapshot of all the triangles
// currently in the mesh. The resulting slice is a copy,
// and will not change as the mesh is updated.
func (m *Mesh) SegmentSlice() []*Segment {
	ts := make([]*Segment, 0, len(m.faces))
	for t := range m.faces {
		ts = append(ts, t)
	}
	return ts
}

// SegmentsSlice is exactly like SegmentSlice(), and is
// only implemented for backwards-compatibility.
func (m *Mesh) SegmentsSlice() []*Segment {
	return m.SegmentSlice()
}

// VertexSlice gets a snapshot of all the vertices
// currently in the mesh.
//
// The result is a copy and is in no way connected to the
// mesh in memory.
func (m *Mesh) VertexSlice() []Coord {
	v2f := m.getVertexToFace()
	vertices := make([]Coord, 0, len(v2f))
	for v := range v2f {
		vertices = append(vertices, v)
	}
	return vertices
}

// Min gets the component-wise minimum across all the
// vertices in the mesh.
func (m *Mesh) Min() Coord {
	if len(m.faces) == 0 {
		return Coord{}
	}
	var result Coord
	var firstFlag bool
	for t := range m.faces {
		for _, c := range t {
			if !firstFlag {
				result = c
				firstFlag = true
			} else {
				result = result.Min(c)
			}
		}
	}
	return result
}

// Max gets the component-wise maximum across all the
// vertices in the mesh.
func (m *Mesh) Max() Coord {
	if len(m.faces) == 0 {
		return Coord{}
	}
	var result Coord
	var firstFlag bool
	for t := range m.faces {
		for _, c := range t {
			if !firstFlag {
				result = c
				firstFlag = true
			} else {
				result = result.Max(c)
			}
		}
	}
	return result
}

func (m *Mesh) getVertexToFace() map[Coord][]*Segment {
	v2f := m.getVertexToFaceOrNil()
	if v2f != nil {
		return v2f
	}

	// Use a lock to ensure two different maps aren't
	// created and returned on different Goroutines.
	m.v2fCreateLock.Lock()
	defer m.v2fCreateLock.Unlock()

	// Another goroutine could have created a map while we
	// waited on the lock.
	v2f = m.getVertexToFaceOrNil()
	if v2f != nil {
		return v2f
	}

	v2f = map[Coord][]*Segment{}
	for t := range m.faces {
		for _, p := range t {
			v2f[p] = append(v2f[p], t)
		}
	}
	m.vertexToFace.Store(v2f)

	return v2f
}

func (m *Mesh) getVertexToFaceOrNil() map[Coord][]*Segment {
	res := m.vertexToFace.Load()
	if res == nil {
		return nil
	}
	return res.(map[Coord][]*Segment)
}
