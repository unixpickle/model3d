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

// Dist gets the minimum distance from c to a point on the
// line segment.
func (s Segment) Dist(c Coord) float64 {
	return c.Dist(s.Closest(c))
}

// Closest gets the point on the segment closest to c.
func (s Segment) Closest(c Coord) Coord {
	v1 := s[1].Sub(s[0])
	norm := v1.Norm()
	v := v1.Scale(1 / norm)

	v2 := c.Sub(s[0])
	mag := v.Dot(v2)
	if mag > norm {
		return s[1]
	} else if mag < 0 {
		return s[0]
	}

	return v.Scale(mag).Add(s[0])
}

// A Circle is a 2D perfect circle.
type Circle struct {
	Center Coord
	Radius float64
}

func (c *Circle) Min() Coord {
	return c.Center.Sub(Coord{X: c.Radius, Y: c.Radius})
}

func (c *Circle) Max() Coord {
	return c.Center.Add(Coord{X: c.Radius, Y: c.Radius})
}

func (c *Circle) Contains(coord Coord) bool {
	return InBounds(c, coord) && coord.Dist(c.Center) <= c.Radius
}

// A Rect is a 2D axis-aligned rectangle.
type Rect struct {
	MinVal Coord
	MaxVal Coord
}

func (r *Rect) Min() Coord {
	return r.MinVal
}

func (r *Rect) Max() Coord {
	return r.MaxVal
}

func (r *Rect) Contains(c Coord) bool {
	return InBounds(r, c)
}
