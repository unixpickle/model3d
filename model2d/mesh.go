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

	// Stores a *CoordToFaces
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
// collection of segments.
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

// NewMeshRect creates a rectangle mesh.
func NewMeshRect(min, max Coord) *Mesh {
	m := NewMesh()
	p1 := XY(min.X, max.Y)
	p2 := XY(max.X, min.Y)
	m.Add(&Segment{min, p1})
	m.Add(&Segment{p1, max})
	m.Add(&Segment{max, p2})
	m.Add(&Segment{p2, min})
	return m
}

// Add adds the segment f to the mesh.
func (m *Mesh) Add(f *Segment) {
	v2f := m.getVertexToFaceOrNil()
	if v2f == nil {
		m.faces[f] = true
		return
	} else if m.faces[f] {
		return
	}

	uniqueVertices(f, func(p Coord) {
		v2f.Append(p, f)
	})
	m.faces[f] = true
}

// AddMesh adds all the segments from m1 to m.
func (m *Mesh) AddMesh(m1 *Mesh) {
	m1.Iterate(m.Add)
}

// Copy returns a shallow copy of m, where all of the
// segments are the same exact pointers.
func (m *Mesh) Copy() *Mesh {
	m1 := NewMesh()
	m1.AddMesh(m)
	return m1
}

// DeepCopy returns a deep copy of m, where all of the
// segments are copied individually.
func (m *Mesh) DeepCopy() *Mesh {
	m1 := NewMesh()
	m.Iterate(func(f *Segment) {
		f1 := new(Segment)
		*f1 = *f
		m1.Add(f1)
	})
	return m1
}

// Remove removes the segment f from the mesh.
//
// It looks at f as a pointer, so the pointer must be
// exactly the same as one passed to Add.
func (m *Mesh) Remove(f *Segment) {
	if !m.faces[f] {
		return
	}
	delete(m.faces, f)
	v2f := m.getVertexToFaceOrNil()
	if v2f != nil {
		uniqueVertices(f, func(p Coord) {
			m.removeFaceFromVertex(v2f, f, p)
		})
	}
}

func (m *Mesh) removeFaceFromVertex(v2f *CoordToFaces, f *Segment, p Coord) {
	s := v2f.Value(p)
	for i, f1 := range s {
		if f1 == f {
			essentials.UnorderedDelete(&s, i)
			break
		}
	}
	if len(s) == 0 {
		v2f.Delete(p)
	} else {
		v2f.Store(p, s)
	}
}

// Contains checks if f has been added to the mesh.
func (m *Mesh) Contains(f *Segment) bool {
	_, ok := m.faces[f]
	return ok
}

// Iterate calls f for every segment in m in an arbitrary
// order.
//
// If f adds or removes segments, they will not be visited.
func (m *Mesh) Iterate(f func(*Segment)) {
	m.IterateSorted(f, nil)
}

// IterateSorted is like Iterate, but it first sorts all
// the segments according to a less than function, cmp.
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
		if _, ok := v2f.Load(c); ok {
			f(c)
		}
	}
}

// Neighbors gets all the segments with a side touching a
// given segment f.
//
// The segment f itself is not included in the results.
//
// The segment f needn't be in the mesh. However, if it is
// not in the mesh, but an equivalent segment is, then said
// equivalent segment will be in the results.
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
	v2f := m.getVertexToFace()
	for _, p := range t {
		for _, t1 := range v2f.Value(p) {
			if t1 != t {
				counts[t1]++
			}
		}
	}
	return counts
}

// Find gets all the segments that contain all of the passed
// points.
//
// For example, to find all segments containing a line from
// from p1 to p2, you could do m.Find(p1, p2).
func (m *Mesh) Find(ps ...Coord) []*Segment {
	if len(ps) == 1 {
		return append([]*Segment{}, m.getVertexToFace().Value(ps[0])...)
	}

	faces := m.getVertexToFace().Value(ps[0])
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

// Translate returns a mesh with all coordinates added to
// a given vector.
func (m *Mesh) Translate(v Coord) *Mesh {
	return m.MapCoords(v.Add)
}

// Rotate returns a mesh with all coordinates rotated
// around the origin by a given angle (in radians).
func (m *Mesh) Rotate(angle float64) *Mesh {
	return m.Transform(Rotation(angle))
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord) Coord) *Mesh {
	mapping := NewCoordToCoord()
	if v2f := m.getVertexToFaceOrNil(); v2f != nil {
		v2f.KeyRange(func(c Coord) bool {
			mapping.Store(c, f(c))
			return true
		})
	} else {
		for t := range m.faces {
			for _, c := range t {
				if _, ok := mapping.Load(c); !ok {
					mapping.Store(c, f(c))
				}
			}
		}
	}
	m1 := NewMesh()
	m.Iterate(func(t *Segment) {
		t1 := *t
		for i, p := range t {
			t1[i] = mapping.Value(p)
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

// SegmentSlice gets a snapshot of all the segments
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
	vertices := make([]Coord, 0, v2f.Len())
	v2f.KeyRange(func(v Coord) bool {
		vertices = append(vertices, v)
		return true
	})
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

func (m *Mesh) getVertexToFace() *CoordToFaces {
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

	v2f = NewCoordToFaces()
	for f := range m.faces {
		uniqueVertices(f, func(p Coord) {
			v2f.Append(p, f)
		})
	}
	m.vertexToFace.Store(v2f)

	return v2f
}

func (m *Mesh) getVertexToFaceOrNil() *CoordToFaces {
	res := m.vertexToFace.Load()
	if res == nil {
		return nil
	}
	return res.(*CoordToFaces)
}

func (m *Mesh) clearVertexToFace() {
	m.vertexToFace = atomic.Value{}
}

func uniqueVertices(face *Segment, f func(Coord)) {
	f(face[0])
	if face[1] != face[0] {
		f(face[1])
	}

}
