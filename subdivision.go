package model3d

// A Subdivider is used for sub-dividing triangles in a
// mesh to add levels of detail where it is needed in a
// model.
//
// To achieve this, it tracks line segments that are to be
// split, and then replaces triangles in the mesh such
// that the given lines are split and the midpoint is
// replaced with a more accurate value.
type Subdivider struct {
	lines map[Segment]bool
}

// NewSubdivider creates an empty Subdivider.
func NewSubdivider() *Subdivider {
	return &Subdivider{lines: map[Segment]bool{}}
}

// Add adds a line segment that needs to be split.
func (s *Subdivider) Add(p1, p2 Coord3D) {
	s.lines[NewSegment(p1, p2)] = true
}

// AddFiltered adds line segments for which f returns
// true.
func (s *Subdivider) AddFiltered(m *Mesh, f func(p1, p2 Coord3D) bool) {
	visited := map[Segment]bool{}
	m.Iterate(func(t *Triangle) {
		segs := [3]Segment{
			NewSegment(t[0], t[1]),
			NewSegment(t[1], t[2]),
			NewSegment(t[2], t[0]),
		}
		for _, seg := range segs {
			if !visited[seg] {
				visited[seg] = true
				if f(seg[0], seg[1]) {
					s.Add(seg[0], seg[1])
				}
			}
		}
	})
}

// Subdivide modifies the mesh by replacing triangles
// whose sides are affected by subdivision.
//
// The midpointFunc computes a 3D coordinate that should
// replace the midpoint of a given line segment from the
// original mesh.
func (s *Subdivider) Subdivide(mesh *Mesh, midpointFunc func(p1, p2 Coord3D) Coord3D) {
	midpoints := map[Segment]Coord3D{}
	for segment := range s.lines {
		midpoints[segment] = midpointFunc(segment[0], segment[1])
	}
	mesh.Iterate(func(t *Triangle) {
		segs := [3]Segment{
			NewSegment(t[0], t[1]),
			NewSegment(t[1], t[2]),
			NewSegment(t[2], t[0]),
		}
		var splits [3]bool
		var numSplits int
		for i, seg := range segs {
			splits[i] = s.lines[seg]
			if splits[i] {
				numSplits++
			}
		}
		if numSplits == 1 {
			for i, seg := range segs {
				if splits[i] {
					subdivideSingle(mesh, t, seg, midpoints[seg])
					break
				}
			}
		} else if numSplits == 2 {
			if !splits[0] {
				subdivideDouble(mesh, t, segs[1], segs[2], midpoints)
			} else if !splits[1] {
				subdivideDouble(mesh, t, segs[0], segs[2], midpoints)
			} else {
				subdivideDouble(mesh, t, segs[0], segs[1], midpoints)
			}
		} else if numSplits == 3 {
			subdivideTriple(mesh, t, midpoints)
		}
	})
}

func subdivideSingle(mesh *Mesh, t *Triangle, splitSeg Segment, midpoint Coord3D) {
	p3 := t[0]
	if p3 == splitSeg[0] || p3 == splitSeg[1] {
		p3 = t[1]
		if p3 == splitSeg[0] || p3 == splitSeg[1] {
			p3 = t[2]
		}
	}
	replaceTriangle(mesh, t, &Triangle{splitSeg[0], midpoint, p3},
		&Triangle{p3, midpoint, splitSeg[1]})
}

func subdivideDouble(mesh *Mesh, t *Triangle, seg1, seg2 Segment,
	midpoints map[Segment]Coord3D) {
	mp1 := midpoints[seg1]
	mp2 := midpoints[seg2]
	shared := seg1.union(seg2)
	unshared1, unshared2 := seg1.inverseUnion(seg2)
	replaceTriangle(mesh, t, &Triangle{shared, mp1, mp2},
		&Triangle{mp1, mp2, unshared1},
		&Triangle{unshared1, unshared2, mp2})
}

func subdivideTriple(mesh *Mesh, t *Triangle, midpoints map[Segment]Coord3D) {
	mp1 := midpoints[NewSegment(t[0], t[1])]
	mp2 := midpoints[NewSegment(t[1], t[2])]
	mp3 := midpoints[NewSegment(t[2], t[0])]
	replaceTriangle(mesh, t, &Triangle{mp1, t[1], mp2},
		&Triangle{mp2, t[2], mp3},
		&Triangle{mp3, t[0], mp1},
		&Triangle{mp1, mp2, mp3})
}

func replaceTriangle(mesh *Mesh, original *Triangle, ts ...*Triangle) {
	mesh.Remove(original)

	norm := original.Normal()
	for _, t := range ts {
		// Make sure the triangle is facing the same way.
		if t.Normal().Dot(norm) < 0 {
			t[1], t[2] = t[2], t[1]
		}
		mesh.Add(t)
	}
}
