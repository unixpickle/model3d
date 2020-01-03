package model2d

import (
	"sort"

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
// It is not safe to access a Mesh from multiple
// Goroutines at once, even for reading.
type Mesh struct {
	segments        map[*Segment]bool
	vertexToSegment map[Coord][]*Segment
}

// NewMesh creates an empty mesh.
func NewMesh() *Mesh {
	return &Mesh{
		segments: map[*Segment]bool{},
	}
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
	if m.vertexToSegment == nil {
		m.segments[s] = true
		return
	} else if m.segments[s] {
		return
	}
	for _, p := range s {
		m.vertexToSegment[p] = append(m.vertexToSegment[p], s)
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
	if m.vertexToSegment == nil {
		return
	}
	for _, p := range s {
		m.removeSegmentFromVertex(s, p)
	}
}

func (m *Mesh) removeSegmentFromVertex(s *Segment, p Coord) {
	v2s := m.vertexToSegment[p]
	for i, s1 := range v2s {
		if s1 == s {
			essentials.UnorderedDelete(&v2s, i)
			break
		}
	}
	m.vertexToSegment[p] = v2s
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
	resSet := map[*Segment]int{}
	for _, p := range ps {
		for _, t1 := range m.getVertexToSegment()[p] {
			resSet[t1]++
		}
	}
	res := make([]*Segment, 0, len(resSet))
	for t1, count := range resSet {
		if count == len(ps) {
			res = append(res, t1)
		}
	}
	return res
}

// MapCoords creates a new mesh by transforming all of the
// coordinates according to the function f.
func (m *Mesh) MapCoords(f func(Coord) Coord) *Mesh {
	mapping := map[Coord]Coord{}
	if m.vertexToSegment != nil {
		for c := range m.vertexToSegment {
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

func (m *Mesh) getVertexToSegment() map[Coord][]*Segment {
	if m.vertexToSegment != nil {
		return m.vertexToSegment
	}
	m.vertexToSegment = map[Coord][]*Segment{}
	for s := range m.segments {
		for _, p := range s {
			m.vertexToSegment[p] = append(m.vertexToSegment[p], s)
		}
	}
	return m.vertexToSegment
}
