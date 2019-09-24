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
