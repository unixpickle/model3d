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
	segments map[*Segment]bool

	// Stores a map[Coord][]*Segment
	vertexToSegment atomic.Value
	v2sCreateLock   sync.Mutex
}

// NewMesh creates an empty mesh.
func NewMesh() *Mesh {
	return &Mesh{
		segments: map[*Segment]bool{},
	}
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

// NewMeshSegments creates a mesh with the given
// collection of Segments.
func NewMeshSegments(segs []*Segment) *Mesh {
	m := NewMesh()
	for _, s := range segs {
		m.Add(s)
	}
	return m
}

// Add adds the Segment s to the mesh.
func (m *Mesh) Add(s *Segment) {
	v2s := m.getVertexToSegmentOrNil()
	if v2s == nil {
		m.segments[s] = true
		return
	} else if m.segments[s] {
		return
	}
	for _, p := range s {
		v2s[p] = append(v2s[p], s)
	}
	m.segments[s] = true
}

// AddMesh adds all the Segments from m1 to m.
func (m *Mesh) AddMesh(m1 *Mesh) {
	m1.Iterate(m.Add)
}

// Remove removes the Segment t from the mesh.
//
// It looks at t as a pointer, so the pointer must be
// exactly the same as a Segment passed to Add.
func (m *Mesh) Remove(s *Segment) {
	if !m.segments[s] {
		return
	}
	delete(m.segments, s)
	v2s := m.getVertexToSegmentOrNil()
	if v2s != nil {
		for _, p := range s {
			m.removeSegmentFromVertex(v2s, s, p)
		}
	}
}

func (m *Mesh) removeSegmentFromVertex(v2s map[Coord][]*Segment, s *Segment, p Coord) {
	segs := v2s[p]
	for i, s1 := range segs {
		if s1 == s {
			essentials.UnorderedDelete(&segs, i)
			break
		}
	}
	if len(segs) == 0 {
		delete(v2s, p)
	} else {
		v2s[p] = segs
	}
}

// Contains checks if s has been added to the mesh.
func (m *Mesh) Contains(s *Segment) bool {
	_, ok := m.segments[s]
	return ok
}

// Iterate calls f for every Segment in m in an arbitrary
// order.
//
// If f adds or removes Segments, they will not be
// visited.
func (m *Mesh) Iterate(f func(s *Segment)) {
	m.IterateSorted(f, nil)
}

// IterateSorted is like Iterate, but it first sorts all
// the Segments according to a less than function, cmp.
func (m *Mesh) IterateSorted(f func(s *Segment), cmp func(s1, s2 *Segment) bool) {
	all := m.SegmentsSlice()
	if cmp != nil {
		sort.Slice(all, func(i, j int) bool {
			return cmp(all[i], all[j])
		})
	}
	for _, s := range all {
		if m.segments[s] {
			f(s)
		}
	}
}

// IterateVertices calls f for every vertex in m in an
// arbitrary order.
//
// If f adds or removes vertices, they will not be
// visited.
func (m *Mesh) IterateVertices(f func(c Coord)) {
	v2s := m.getVertexToSegment()
	for _, c := range m.VertexSlice() {
		if _, ok := v2s[c]; ok {
			f(c)
		}
	}
}

// Neighbors gets all the Segments with a vertex touching
// a given Segment s.
//
// The Segment s itself is not included in the results.
//
// The Segment s needn't be in the mesh. However, if it
// is not in the mesh, but an equivalent Segment is, then
// said equivalent Segment will be in the results.
func (m *Mesh) Neighbors(s *Segment) []*Segment {
	neighbors := map[*Segment]bool{}
	for _, p := range s {
		for _, n := range m.Find(p) {
			if n != s {
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

// Find gets all the Segments that contain all of the
// passed points.
//
// This is only useful with one or two coordinates.
func (m *Mesh) Find(ps ...Coord) []*Segment {
	if len(ps) == 1 {
		return append([]*Segment{}, m.getVertexToSegment()[ps[0]]...)
	}

	segs := m.getVertexToSegment()[ps[0]]
	res := make([]*Segment, 0, len(segs))

SegLoop:
	for _, s := range segs {
		for _, p := range ps[1:] {
			if p != s[0] && p != s[1] {
				continue SegLoop
			}
		}
		res = append(res, s)
	}

	return res
}

// Scale creates a new mesh by scaling the coordinates by
// a factor s.
func (m *Mesh) Scale(s float64) *Mesh {
	return m.MapCoords(Coord{X: s, Y: s}.Mul)
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord) Coord) *Mesh {
	mapping := map[Coord]Coord{}
	if v2s := m.getVertexToSegmentOrNil(); v2s != nil {
		for c := range v2s {
			mapping[c] = f(c)
		}
	} else {
		for t := range m.segments {
			for _, c := range t {
				if _, ok := mapping[c]; !ok {
					mapping[c] = f(c)
				}
			}
		}
	}
	m1 := NewMesh()
	m.Iterate(func(s *Segment) {
		s1 := *s
		for i, p := range s {
			s1[i] = mapping[p]
		}
		m1.Add(&s1)
	})
	return m1
}

// Transform applies t to the coordinates.
func (m *Mesh) Transform(t Transform) *Mesh {
	return m.MapCoords(t.Apply)
}

// SegmentsSlice gets a snapshot of all the Segments
// currently in the mesh. The resulting slice is a copy,
// and will not change as the mesh is updated.
func (m *Mesh) SegmentsSlice() []*Segment {
	segs := make([]*Segment, 0, len(m.segments))
	for s := range m.segments {
		segs = append(segs, s)
	}
	return segs
}

// VertexSlice gets a snapshot of all the vertices
// currently in the mesh.
//
// The result is a copy and is in no way connected to the
// mesh in memory.
func (m *Mesh) VertexSlice() []Coord {
	v2s := m.getVertexToSegment()
	vertices := make([]Coord, 0, len(v2s))
	for v := range v2s {
		vertices = append(vertices, v)
	}
	return vertices
}

// Min gets the component-wise minimum across all the
// vertices in the mesh.
func (m *Mesh) Min() Coord {
	if len(m.segments) == 0 {
		return Coord{}
	}
	var result Coord
	var firstFlag bool
	for s := range m.segments {
		for _, c := range s {
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
	if len(m.segments) == 0 {
		return Coord{}
	}
	var result Coord
	var firstFlag bool
	for s := range m.segments {
		for _, c := range s {
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

func (m *Mesh) getVertexToSegment() map[Coord][]*Segment {
	v2s := m.getVertexToSegmentOrNil()
	if v2s != nil {
		return v2s
	}

	// Use a lock to ensure two different maps aren't
	// created and returned on different Goroutines.
	m.v2sCreateLock.Lock()
	defer m.v2sCreateLock.Unlock()

	// Another goroutine could have created a map while we
	// waited on the lock.
	v2s = m.getVertexToSegmentOrNil()
	if v2s != nil {
		return v2s
	}

	v2s = map[Coord][]*Segment{}
	for s := range m.segments {
		for _, p := range s {
			v2s[p] = append(v2s[p], s)
		}
	}
	m.vertexToSegment.Store(v2s)

	return v2s
}

func (m *Mesh) getVertexToSegmentOrNil() map[Coord][]*Segment {
	res := m.vertexToSegment.Load()
	if res == nil {
		return nil
	}
	return res.(map[Coord][]*Segment)
}
