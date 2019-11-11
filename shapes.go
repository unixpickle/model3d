package model3d

// A Triangle is a triangle in 3D Euclidean space.
type Triangle [3]Coord3D

// Area computes the area of the triangle.
func (t *Triangle) Area() float64 {
	return t.crossProduct().Norm() / 2
}

// Normal computes a normal vector for the triangle using
// the right-hand rule.
func (t *Triangle) Normal() Coord3D {
	return t.crossProduct().Normalize()
}

func (t *Triangle) crossProduct() Coord3D {
	return t[1].Sub(t[0]).Cross(t[2].Sub(t[0]))
}

// Min gets the element-wise minimum of all the points.
func (t *Triangle) Min() Coord3D {
	return t[0].Min(t[1]).Min(t[2])
}

// Max gets the element-wise maximum of all the points.
func (t *Triangle) Max() Coord3D {
	return t[0].Max(t[1]).Max(t[2])
}

// Segments gets all three line segments in the triangle.
func (t *Triangle) Segments() [3]Segment {
	var res [3]Segment
	for i := 0; i < 3; i++ {
		res[i] = NewSegment(t[i], t[(i+1)%3])
	}
	return res
}

// SharesEdge checks if p shares exactly one edge with p1.
func (t *Triangle) SharesEdge(t1 *Triangle) bool {
	return t.inCommon(t1) == 2
}

func (t *Triangle) inCommon(t1 *Triangle) int {
	inCommon := 0
	for _, p := range t {
		for _, p1 := range t1 {
			if p == p1 {
				inCommon += 1
				break
			}
		}
	}
	return inCommon
}

// AreaGradient computes the gradient of the triangle's
// area with respect to every coordinate.
func (t *Triangle) AreaGradient() *Triangle {
	var grad Triangle
	for i, p := range t {
		p1 := t[(i+1)%3]
		p2 := t[(i+2)%3]
		base := p2.Sub(p1)
		baseNorm := base.Norm()
		if baseNorm == 0 {
			continue
		}
		// Project the base out of v to get the height
		// vector of the triangle.
		normed := base.Scale(1 / baseNorm)
		v := p.Sub(p1)
		v = v.Sub(normed.Scale(normed.Dot(v)))

		vNorm := v.Norm()
		if vNorm == 0 {
			continue
		}
		grad[i] = v.Scale(baseNorm / (2 * vNorm))
	}
	return &grad
}

// A Segment is a line segment in a canonical ordering,
// such that segments can be compared via the == operator
// even if they were created with their points in the
// opposite order.
type Segment [2]Coord3D

// NewSegment creates a segment with the canonical
// ordering.
func NewSegment(p1, p2 Coord3D) Segment {
	if p1.X < p2.X || (p1.X == p2.X && p1.Y < p2.Y) ||
		(p1.X == p2.X && p1.Y == p2.Y && p1.Z < p2.Z) {
		return Segment{p1, p2}
	} else {
		return Segment{p2, p1}
	}
}

// Mid gets the midpoint of the segment.
func (s Segment) Mid() Coord3D {
	return s[0].Mid(s[1])
}

// other gets the third point in a triangle for which s is
// a segment.
func (s Segment) other(t *Triangle) Coord3D {
	if t[0] != s[0] && t[0] != s[1] {
		return t[0]
	} else if t[1] != s[0] && t[1] != s[1] {
		return t[1]
	} else {
		return t[2]
	}
}

// union finds the point that s and s1 have in common,
// assuming that they have exactly one point in common.
func (s Segment) union(s1 Segment) Coord3D {
	if s1[0] == s[0] || s1[0] == s[1] {
		return s1[0]
	} else {
		return s1[1]
	}
}

// inverseUnion finds the two points that s and s1 do not
// have in common, assuming that they have exactly one
// point in common.
//
// The first point is from s, and the second is from s1.
func (s Segment) inverseUnion(s1 Segment) (Coord3D, Coord3D) {
	union := s.union(s1)
	if union == s1[0] {
		if union == s[0] {
			return s[1], s1[1]
		} else {
			return s[0], s1[1]
		}
	} else {
		if union == s[0] {
			return s[1], s1[0]
		} else {
			return s[0], s1[0]
		}
	}
}
