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

// A Rect is a 2D primitive that fills an axis-aligned
// rectangular space.
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

// Min yields r.MinVal.
func (r *Rect) Min() Coord {
	return r.MinVal
}

// Max yields r.MaxVal.
func (r *Rect) Max() Coord {
	return r.MaxVal
}

// Contains checks if c is inside of r.
func (r *Rect) Contains(c Coord) bool {
	return c.Min(r.MinVal) == r.MinVal && c.Max(r.MaxVal) == r.MaxVal
}

// FirstRayCollision gets the first ray collision with the
// rectangular surface.
func (r *Rect) FirstRayCollision(ray *Ray) (RayCollision, bool) {
	tMin, tMax := rayCollisionWithBounds(ray, r.MinVal, r.MaxVal)
	if tMax < tMin || tMax < 0 {
		return RayCollision{}, false
	}

	t := tMin
	if t < 0 {
		t = tMax
	}

	return RayCollision{
		Scale:  t,
		Normal: r.normalAt(ray.Origin.Add(ray.Direction.Scale(t))),
	}, true
}

// RayCollisions calls f (if non-nil) with each ray
// collision with the rectangular surface.
// It returns the number of collisions.
func (r *Rect) RayCollisions(ray *Ray, f func(RayCollision)) int {
	tMin, tMax := rayCollisionWithBounds(ray, r.MinVal, r.MaxVal)
	if tMax < tMin || tMax < 0 {
		return 0
	}

	var count int
	for _, t := range []float64{tMin, tMax} {
		if t < 0 {
			continue
		}
		count++
		if f != nil {
			f(RayCollision{
				Scale:  t,
				Normal: r.normalAt(ray.Origin.Add(ray.Direction.Scale(t))),
			})
		}
	}
	return count
}

func (r *Rect) normalAt(c Coord) Coord {
	var axis int
	var sign float64
	minDist := math.Inf(1)

	minArr := r.MinVal.Array()
	maxArr := r.MaxVal.Array()
	cArr := c.Array()
	for i, cVal := range cArr {
		if d := math.Abs(cVal - minArr[i]); d < minDist {
			minDist = d
			sign = -1
			axis = i
		}
		if d := math.Abs(cVal - maxArr[i]); d < minDist {
			minDist = d
			sign = 1
			axis = i
		}
	}

	var resArr [2]float64
	resArr[axis] = sign
	return NewCoordArray(resArr)
}

// SphereCollision checks if a solid sphere touches any
// part of the rectangular surface.
func (r *Rect) SphereCollision(c Coord, radius float64) bool {
	return math.Abs(r.SDF(c)) <= radius
}

// SDF gets the signed distance to the surface of the
// rectangular volume.
func (r *Rect) SDF(c Coord) float64 {
	return r.genericSDF(c, nil, nil)
}

// PointSDF gets the nearest point on the surface of the
// rect and the corresponding SDF.
func (r *Rect) PointSDF(c Coord) (Coord, float64) {
	var p Coord
	res := r.genericSDF(c, nil, &p)
	return p, res
}

// NormalSDF gets the nearest point on the surface of the
// rect and the corresponding SDF.
func (r *Rect) NormalSDF(c Coord) (Coord, float64) {
	var n Coord
	res := r.genericSDF(c, &n, nil)
	return n, res
}

func (r *Rect) genericSDF(c Coord, normalOut, pointOut *Coord) float64 {
	if !r.Contains(c) {
		// We can project directly to the rect to hit the surface.
		nearest := c.Min(r.MaxVal).Max(r.MinVal)
		if pointOut != nil {
			*pointOut = nearest
		}
		if normalOut != nil {
			*normalOut = r.normalAt(nearest)
		}
		return -c.Dist(nearest)
	}

	// Find the closest side of the rect.
	minArr := r.MinVal.Array()
	maxArr := r.MaxVal.Array()
	cArr := c.Array()
	dist := math.Inf(1)
	for i := 0; i < 2; i++ {
		minD := cArr[i] - minArr[i]
		maxD := maxArr[i] - cArr[i]
		axisD := math.Min(minD, maxD)
		if axisD < dist {
			dist = axisD
			if normalOut != nil {
				var arr [2]float64
				if minD < maxD {
					arr[i] = -1.0
				} else {
					arr[i] = 1.0
				}
				*normalOut = NewCoordArray(arr)
			}
			if pointOut != nil {
				arr := cArr
				if minD < maxD {
					arr[i] = minArr[i]
				} else {
					arr[i] = maxArr[i]
				}
				*pointOut = NewCoordArray(arr)
			}
		}
	}
	return dist
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

// A Capsule is a shape which contains all of the points
// within a given distance of a line segment.
type Capsule struct {
	P1     Coord
	P2     Coord
	Radius float64
}

// Min gets the minimum point of the bounding box.
func (c *Capsule) Min() Coord {
	return c.P1.Min(c.P2).AddScalar(-c.Radius)
}

// Max gets the maximum point of the bounding box.
func (c *Capsule) Max() Coord {
	return c.P1.Max(c.P2).AddScalar(c.Radius)
}

// Contains checks if c is within the capsule.
func (c *Capsule) Contains(coord Coord) bool {
	segment := Segment{c.P1, c.P2}
	return segment.Dist(coord) <= c.Radius
}

// SDF gets the signed distance to the surface of the capsule.
func (c *Capsule) SDF(coord Coord) float64 {
	return c.genericSDF(coord, nil, nil)
}

// PointSDF gets the nearest point on the surface of the
// capsule and the corresponding SDF.
func (c *Capsule) PointSDF(coord Coord) (Coord, float64) {
	var p Coord
	res := c.genericSDF(coord, nil, &p)
	return p, res
}

// NormalSDF gets the nearest point on the surface of the
// capsule and the corresponding SDF.
func (c *Capsule) NormalSDF(coord Coord) (Coord, float64) {
	var n Coord
	res := c.genericSDF(coord, &n, nil)
	return n, res
}

func (c *Capsule) genericSDF(coord Coord, normalOut, pointOut *Coord) float64 {
	v := c.P2.Sub(c.P1)
	norm := v.Norm()
	axis := v.Scale(1 / norm)
	dot := coord.Sub(c.P1).Dot(axis)
	if dot < 0 || dot > norm {
		proxy := &Circle{Radius: c.Radius}
		if dot < 0 {
			proxy.Center = c.P1
		} else {
			proxy.Center = c.P2
		}
		if normalOut != nil {
			*normalOut, _ = proxy.NormalSDF(coord)
		}
		if pointOut != nil {
			*pointOut, _ = proxy.PointSDF(coord)
		}
		return proxy.SDF(coord)
	}

	sdf := c.Radius - Segment{c.P1, c.P2}.Dist(coord)
	if normalOut != nil || pointOut != nil {
		projPoint := c.P1.Add(axis.Scale(dot))
		delta := coord.Sub(projPoint)

		b1 := XY(-axis.Y, axis.X)
		normal := safeNormal(delta, b1, axis)
		if normalOut != nil {
			*normalOut = normal
		}
		if pointOut != nil {
			*pointOut = projPoint.Add(normal.Scale(c.Radius))
		}
	}
	return sdf
}

func safeNormal(direction, fallbackDirection, invalidDirection Coord) Coord {
	if norm := direction.Norm(); norm == 0 {
		// Fully degenerate case.
		direction = fallbackDirection
	} else {
		direction = direction.Scale(1 / norm)

		// When direction was very small, it might be pointing in
		// an invalid direction once we normalize it.
		direction = direction.ProjectOut(invalidDirection)
		if norm := direction.Norm(); norm < 1e-5 {
			// Most of the direction was due to rounding error.
			direction = fallbackDirection
		} else {
			direction = direction.Scale(1 / norm)
		}
	}
	return direction
}
