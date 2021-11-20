// Generated from templates/shapes.template

package model2d

import (
	"math"
)

// A Circle is a 2D perfect circle.
type Circle struct {
	Center Coord
	Radius float64
}

// Min gets the minimum point of the bounding box.
func (c *Circle) Min() Coord {
	return c.Center.AddScalar(-c.Radius)
}

// Max gets the maximum point of the bounding box.
func (c *Circle) Max() Coord {
	return c.Center.AddScalar(c.Radius)
}

// Contains checks if a point c is inside the circle.
func (c *Circle) Contains(coord Coord) bool {
	return coord.Dist(c.Center) <= c.Radius
}

// FirstRayCollision gets the first ray collision with the
// circle, if one occurs.
func (c *Circle) FirstRayCollision(r *Ray) (RayCollision, bool) {
	var res RayCollision
	var ok bool
	c.RayCollisions(r, func(rc RayCollision) {
		// Collisions are sorted from first to last.
		if !ok {
			res = rc
			ok = true
		}
	})
	return res, ok
}

// RayCollisions calls f (if non-nil) with every ray
// collision.
//
// It returns the total number of collisions.
func (c *Circle) RayCollisions(r *Ray, f func(RayCollision)) int {
	// Want to find where ||(o+a*d)-c||^2 = r^2
	// Let's call o = (o-c) for the rest of this.
	// ||a*d+o||^2 = r^2
	// a^2*d^2 + 2*a*d*o + o^2 = r^2
	// a^2*(d^2) + a*(2*d*o) + (o^2 - r^2) = 0
	// quadratic equation: a=(d^2), b=(2*d*o), c = (o^2 - r^2)
	o := r.Origin.Sub(c.Center)
	d := r.Direction
	a := d.Dot(d)
	b := 2 * d.Dot(o)
	c_ := o.Dot(o) - c.Radius*c.Radius

	discriminant := b*b - 4*a*c_
	if discriminant <= 0 {
		return 0
	}

	sqrtDisc := math.Sqrt(discriminant)
	t1 := (-b + sqrtDisc) / (2 * a)
	t2 := (-b - sqrtDisc) / (2 * a)
	if t1 > t2 {
		t1, t2 = t2, t1
	}

	var count int
	for _, t := range []float64{t1, t2} {
		if t < 0 {
			continue
		}
		count++
		if f != nil {
			point := r.Origin.Add(r.Direction.Scale(t))
			normal := point.Sub(c.Center).Normalize()
			f(RayCollision{Normal: normal, Scale: t})
		}
	}

	return count
}

// CircleCollision checks if the surface of c collides
// with a solid circle centered at c with radius r.
func (c *Circle) CircleCollision(center Coord, r float64) bool {
	return math.Abs(c.SDF(center)) <= r
}

// SDF gets the signed distance relative to the circle.
func (c *Circle) SDF(coord Coord) float64 {
	return c.Radius - coord.Dist(c.Center)
}

// PointSDF gets the signed distance function at coord
// and also returns the nearest point to it on the circle.
func (c *Circle) PointSDF(coord Coord) (Coord, float64) {
	direction := coord.Sub(c.Center)
	if norm := direction.Norm(); norm == 0 {
		// Pick an arbitrary point
		return c.Center.Add(X(c.Radius)), c.Radius
	} else {
		return c.Center.Add(direction.Scale(c.Radius / norm)), c.SDF(coord)
	}
}

// NormalSDF gets the signed distance function at coord
// and also returns the normal at the nearest point to it
// on the circle.
func (c *Circle) NormalSDF(coord Coord) (Coord, float64) {
	direction := coord.Sub(c.Center)
	if norm := direction.Norm(); norm == 0 {
		// Pick an arbitrary normal
		return X(1), c.Radius
	} else {
		return direction.Scale(1 / norm), c.SDF(coord)
	}
}

// A Rect is a 2D axis-aligned rectangle.
type Rect struct {
	MinVal Coord
	MaxVal Coord
}

// NewRect creates a Rect with a min and a max value.
func NewRect(min, max Coord) *Rect {
	return &Rect{MinVal: min, MaxVal: max}
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

// Expand returns a new Rect that is delta units further
// along in every direction, making it a total of 2*delta
// units longer along each axis.
func (r *Rect) Expand(delta float64) *Rect {
	return &Rect{
		MinVal: r.MinVal.AddScalar(-delta),
		MaxVal: r.MaxVal.AddScalar(delta),
	}
}
