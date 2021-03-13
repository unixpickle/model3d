package model2d

import "math"

// A Segment is a 2-dimensional line segment.
//
// The order determines the normal direction.
//
// In particular, if the segments in a polygon go in the
// clockwise direction, assuming the y-axis faces up, then
// the normals face outwards from the polygon.
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

// RayCollisions calls f (if non-nil) with a collision (if
// applicable) and returns the collisions count (0 or 1).
func (s *Segment) RayCollisions(r *Ray, f func(RayCollision)) int {
	if collides, scale := s.rayCollision(r); collides && scale >= 0 {
		if f != nil {
			f(RayCollision{Scale: scale, Normal: s.Normal()})
		}
		return 1
	}
	return 0
}

// FirstRayCollision gets the ray collision if there is
// one.
func (s *Segment) FirstRayCollision(r *Ray) (RayCollision, bool) {
	if collides, scale := s.rayCollision(r); collides && scale >= 0 {
		return RayCollision{Scale: scale, Normal: s.Normal()}, true
	}
	return RayCollision{}, false
}

// rayCollision computes where (and if) the ray intersects
// the segment.
//
// If it returns true as the first value, then the ray or
// its reverse hits the segment.
//
// The second return value is how much of the direction
// must be added to the origin to hit the line containing
// the segment.
// If it is negative, it means the segment is behind the
// ray.
func (s *Segment) rayCollision(r *Ray) (bool, float64) {
	v := s[1].Sub(s[0])
	matrix := Matrix2{
		v.X, r.Direction.X,
		v.Y, r.Direction.Y,
	}
	if math.Abs(matrix.Det()) < 1e-8*s.Length()*r.Direction.Norm() {
		return false, 0
	}
	matrix.InvertInPlace()
	result := matrix.MulColumn(r.Origin.Sub(s[0]))
	return result.X >= 0 && result.X <= 1, -result.Y
}

// CircleCollision checks if the circle intersects the
// segment s.
func (s *Segment) CircleCollision(c Coord, r float64) bool {
	if s[0].Dist(c) < r || s[1].Dist(c) < r {
		return true
	}

	// The segment may pass through the circle without
	// either endpoint being contained.
	v := s[1].Sub(s[0])
	frac := (c.Dot(v) - s[0].Dot(v)) / v.Dot(v)
	closest := s[0].Add(v.Scale(frac))
	return frac >= 0 && frac <= 1 && closest.Dist(c) < r
}

// SegmentCollision returns true if s intersects s1.
func (s *Segment) SegmentCollision(s1 *Segment) bool {
	collides, t := s.rayCollision(&Ray{Origin: s1[0], Direction: s1[1].Sub(s1[0])})
	return collides && t >= 0 && t <= 1
}

// RectCollision returns true if any part of the segment
// is inside of the rectangle.
func (s *Segment) RectCollision(r *Rect) bool {
	minPoint := s.Min()
	maxPoint := s.Max()
	if minPoint.X > r.MaxVal.X || minPoint.Y > r.MaxVal.Y {
		return false
	}
	if maxPoint.X < r.MinVal.X || maxPoint.Y < r.MinVal.Y {
		return false
	}
	if r.Contains(s[0]) || r.Contains(s[1]) {
		return true
	}
	outline := [4]Segment{
		{r.MinVal, XY(r.MaxVal.X, r.MinVal.Y)},
		{r.MinVal, XY(r.MinVal.X, r.MaxVal.Y)},
		{r.MaxVal, XY(r.MaxVal.X, r.MinVal.Y)},
		{r.MaxVal, XY(r.MinVal.X, r.MaxVal.Y)},
	}
	for _, seg := range outline {
		if s.SegmentCollision(&seg) {
			return true
		}
	}
	return false
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

// NewRect creates a Rect with a min and a max value.
func NewRect(min, max Coord) *Rect {
	return NewRect(min, max)
}

// BoundsRect creates a Rect from a Bounder's bounds.
func BoundsRect(b Bounder) *Rect {
	return NewRect(b.Min(), b.Max())
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
