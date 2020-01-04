package model2d

// A Segment is a 2-dimensional line segment.
//
// The order determines the normal direction.
//
// In particular, if the segments in a polygon go in the
// clockwise direction (assuming the y-axis faces down),
// then the normals face outwards from the polygon.
type Segment [2]Coord

// Normal computes the normal vector to the segment,
// facing outwards from the surface.
func (s *Segment) Normal() Coord {
	delta := s[1].Sub(s[0])
	return (Coord{X: -delta.Y, Y: delta.X}).Normalize()
}

// Min gets the element-wise minimum of the endpoints.
func (s Segment) Min() Coord {
	return s[0].Min(s[1])
}

// Max gets the element-wise maximum of the endpoints.
func (s Segment) Max() Coord {
	return s[0].Max(s[1])
}

// Mid gets the midpoint of the segment.
func (s Segment) Mid() Coord {
	return s[0].Mid(s[1])
}

// Length gets the length of the segment.
func (s Segment) Length() float64 {
	return s[1].Sub(s[0]).Norm()
}
