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
