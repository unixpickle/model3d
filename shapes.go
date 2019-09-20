package model3d

// A Triangle is a triangle in 3D Euclidean space.
type Triangle [3]Coord3D

// Normal computes a normal vector for the triangle using
// the right-hand rule.
func (t *Triangle) Normal() Coord3D {
	x1, y1, z1 := t[1].X-t[0].X, t[1].Y-t[0].Y, t[1].Z-t[0].Z
	x2, y2, z2 := t[2].X-t[0].X, t[2].Y-t[0].Y, t[2].Z-t[0].Z

	// The standard cross product formula.
	result := Coord3D{
		X: y1*z2 - z1*y2,
		Y: z1*x2 - x1*z2,
		Z: x1*y2 - y1*x2,
	}

	return result.Scale(1 / result.Norm())
}

// A Segment is a line segment in a canonical ordering,
// such that segments can be compared via the == operator
// even if they were created with their points in the
// opposite order.
type Segment [2]Coord3D

// NewSegment creates a segment with the canonical
// ordering.
func NewSegment(p1, p2 Coord3D) Segment {
	if p1.X < p2.X || (p1.X == p2.X && p1.Y < p2.Y) {
		return Segment{p1, p2}
	} else {
		return Segment{p2, p1}
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
