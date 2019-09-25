package model3d

import (
	"math"
)

// A Solid is a boolean function in 3D where a value of
// true indicates that a point is part of the solid, and
// false indicates that it is not.
type Solid interface {
	// Get the corners of a bounding rectangular volume.
	// Outside of this volume, Contains() must always
	// return false.
	Min() Coord3D
	Max() Coord3D

	Contains(p Coord3D) bool
}

// A SphereSolid is a Solid that yields values for a
// sphere.
type SphereSolid struct {
	Center Coord3D
	Radius float64
}

func (s *SphereSolid) Min() Coord3D {
	return Coord3D{X: s.Center.X - s.Radius, Y: s.Center.Y - s.Radius, Z: s.Center.Z - s.Radius}
}

func (s *SphereSolid) Max() Coord3D {
	return Coord3D{X: s.Center.X + s.Radius, Y: s.Center.Y + s.Radius, Z: s.Center.Z + s.Radius}
}

func (s *SphereSolid) Contains(p Coord3D) bool {
	return p.Dist(s.Center) <= s.Radius
}

// A CylinderSolid is a Solid that yields values for a
// cylinder. The cylinder is defined as all the positions
// within a Radius distance from the line segment between
// P1 and P2.
type CylinderSolid struct {
	P1     Coord3D
	P2     Coord3D
	Radius float64
}

func (c *CylinderSolid) Min() Coord3D {
	return Coord3D{
		X: math.Min(c.P1.X, c.P2.X) - c.Radius,
		Y: math.Min(c.P1.Y, c.P2.Y) - c.Radius,
		Z: math.Min(c.P1.Z, c.P2.Z) - c.Radius,
	}
}

func (c *CylinderSolid) Max() Coord3D {
	return Coord3D{
		X: math.Max(c.P1.X, c.P2.X) + c.Radius,
		Y: math.Max(c.P1.Y, c.P2.Y) + c.Radius,
		Z: math.Max(c.P1.Z, c.P2.Z) + c.Radius,
	}
}

func (c *CylinderSolid) Contains(p Coord3D) bool {
	diff := c.P1.Add(c.P2.Scale(-1))
	direction := diff.Scale(1 / diff.Norm())
	frac := p.Add(c.P2.Scale(-1)).Dot(direction)
	if frac < 0 || frac > diff.Norm() {
		return false
	}
	projection := c.P2.Add(direction.Scale(frac))
	return projection.Dist(p) <= c.Radius
}

// A TorusSolid is a Solid that yields values for a torus.
// The torus is defined by revolving a circle with a
// radius InnerRadius around the point Center and around
// the axis Axis, at a distance of OuterRadius from
// Center.
type TorusSolid struct {
	Center      Coord3D
	Axis        Coord3D
	OuterRadius float64
	InnerRadius float64
}

func (t *TorusSolid) Min() Coord3D {
	maxSize := t.InnerRadius + t.OuterRadius
	return t.Center.Add(Coord3D{X: -maxSize, Y: -maxSize, Z: -maxSize})
}

func (t *TorusSolid) Max() Coord3D {
	maxSize := t.InnerRadius + t.OuterRadius
	return t.Center.Add(Coord3D{X: maxSize, Y: maxSize, Z: maxSize})
}

func (t *TorusSolid) Contains(c Coord3D) bool {
	b1, b2 := t.planarBasis()
	centered := c.Add(t.Center.Scale(-1))

	// Compute the closest point on the ring around
	// the center of the torus.
	x := b1.Dot(centered)
	y := b2.Dot(centered)
	scale := t.OuterRadius / math.Sqrt(x*x+y*y)
	x *= scale
	y *= scale
	ringPoint := b1.Scale(x).Add(b2.Scale(y))

	return ringPoint.Dist(centered) <= t.InnerRadius
}

func (t *TorusSolid) planarBasis() (Coord3D, Coord3D) {
	// Create the first basis vector by swapping two
	// coordinates and negating one of them.
	// For numerical stability, we involve the component
	// with the largest absolute value.
	var basis1 Coord3D
	if math.Abs(t.Axis.X) > math.Abs(t.Axis.Y) && math.Abs(t.Axis.X) > math.Abs(t.Axis.Z) {
		basis1.X = t.Axis.Y
		basis1.Y = -t.Axis.X
	} else {
		basis1.Y = t.Axis.Z
		basis1.Z = -t.Axis.Y
	}

	// Create the second basis vector using a cross product.
	basis2 := Coord3D{
		basis1.Y*t.Axis.Z - basis1.Z*t.Axis.Y,
		basis1.Z*t.Axis.X - basis1.X*t.Axis.Z,
		basis1.X*t.Axis.Y - basis1.Y*t.Axis.X,
	}

	basis1 = basis1.Scale(1 / basis1.Norm())
	basis2 = basis2.Scale(1 / basis2.Norm())
	return basis1, basis2
}

// A JoinedSolid is a Solid composed of other solids.
type JoinedSolid []Solid

func (j JoinedSolid) Min() Coord3D {
	min := j[0].Min()
	for _, s := range j[1:] {
		min1 := s.Min()
		min.X = math.Min(min.X, min1.X)
		min.Y = math.Min(min.Y, min1.Y)
		min.Z = math.Min(min.Z, min1.Z)
	}
	return min
}

func (j JoinedSolid) Max() Coord3D {
	max := j[0].Max()
	for _, s := range j[1:] {
		max1 := s.Max()
		max.X = math.Max(max.X, max1.X)
		max.Y = math.Max(max.Y, max1.Y)
		max.Z = math.Max(max.Z, max1.Z)
	}
	return max
}

func (j JoinedSolid) Contains(c Coord3D) bool {
	for _, s := range j {
		if s.Contains(c) {
			return true
		}
	}
	return false
}

// A ColliderSolid is a Solid that uses a Collider to
// check if points are in the solid.
type ColliderSolid struct {
	min       Coord3D
	max       Coord3D
	collider  Collider
	direction Coord3D
}

// NewColliderSolid creates a ColliderSolid.
func NewColliderSolid(min, max Coord3D, collider Collider) *ColliderSolid {
	return &ColliderSolid{
		min:       min,
		max:       max,
		collider:  collider,
		direction: Coord3D{0.5224892708603626, 0.10494477243214506, 0.43558938446126527},
	}
}

func (c *ColliderSolid) Min() Coord3D {
	return c.min
}

func (c *ColliderSolid) Max() Coord3D {
	return c.max
}

func (c *ColliderSolid) Contains(p Coord3D) bool {
	return c.collider.RayCollisions(&Ray{
		Origin:    p,
		Direction: c.direction,
	})%2 == 1
}
