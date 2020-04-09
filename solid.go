package model3d

import (
	"math"
)

// A Solid is a boolean function in 3D where a value of
// true indicates that a point is part of the solid, and
// false indicates that it is not.
type Solid interface {
	// Contains must always return false outside of the
	// boundaries of the solid.
	Bounder

	Contains(p Coord3D) bool
}

// Backwards compatibility type aliases.
type RectSolid = Rect
type SphereSolid = Sphere
type CylinderSolid = Cylinder
type TorusSolid = Torus

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

// SubtractedSolid is a Solid consisting of all the points
// in Positive which are not in Negative.
type SubtractedSolid struct {
	Positive Solid
	Negative Solid
}

func (s *SubtractedSolid) Min() Coord3D {
	return s.Positive.Min()
}

func (s *SubtractedSolid) Max() Coord3D {
	return s.Positive.Max()
}

func (s *SubtractedSolid) Contains(c Coord3D) bool {
	return s.Positive.Contains(c) && !s.Negative.Contains(c)
}

// IntersectedSolid is a Solid containing the intersection
// of one or more Solids.
type IntersectedSolid []Solid

func (i IntersectedSolid) Min() Coord3D {
	bound := i[0].Min()
	for _, s := range i[1:] {
		bound = bound.Max(s.Min())
	}
	return bound
}

func (i IntersectedSolid) Max() Coord3D {
	bound := i[0].Max()
	for _, s := range i[1:] {
		bound = bound.Min(s.Max())
	}
	// Prevent negative area.
	return bound.Max(i.Min())
}

func (i IntersectedSolid) Contains(c Coord3D) bool {
	for _, s := range i {
		if !s.Contains(c) {
			return false
		}
	}
	return true
}

// A StackedSolid is like a JoinedSolid, but the solids
// after the first are moved so that the lowest Z value of
// their bounding box collides with the highest Z value of
// the previous solid.
// In other words, the solids are stacked on top of each
// other along the Z axis.
type StackedSolid []Solid

func (s StackedSolid) Min() Coord3D {
	return JoinedSolid(s).Min()
}

func (s StackedSolid) Max() Coord3D {
	lastMax := s[0].Max()
	for i := 1; i < len(s); i++ {
		newMax := s[i].Max()
		newMax.Z += lastMax.Z - s[i].Min().Z
		lastMax = lastMax.Max(newMax)
	}
	return lastMax
}

func (s StackedSolid) Contains(c Coord3D) bool {
	if !InBounds(s, c) {
		return false
	}
	currentZ := s[0].Min().Z
	for _, solid := range s {
		delta := currentZ - solid.Min().Z
		if solid.Contains(c.Sub(Coord3D{Z: delta})) {
			return true
		}
		currentZ = solid.Max().Z + delta
	}
	return false
}

// A ColliderSolid is a Solid that uses a Collider to
// check if points are in the solid.
//
// There are two modes for a ColliderSolid. In the first,
// points are inside the solid if a ray passes through the
// surface of the Collider an odd number of times.
// In the second, points are inside the solid if a sphere
// of a pre-determined radius touches the surface of the
// Collider from the point.
// The second modality is equivalent to creating a thick
// but hollow solid.
type ColliderSolid struct {
	min       Coord3D
	max       Coord3D
	collider  Collider
	direction Coord3D

	hollowRadius float64
}

// NewColliderSolid creates a ColliderSolid.
func NewColliderSolid(collider Collider) *ColliderSolid {
	return &ColliderSolid{
		min:      collider.Min(),
		max:      collider.Max(),
		collider: collider,

		// Random direction; any direction should work, but we
		// want to avoid edge cases and rounding errors.
		direction: Coord3D{0.5224892708603626, 0.10494477243214506, 0.43558938446126527},
	}
}

// NewColliderSolidHollow creates a ColliderSolid which
// includes all points within r distance from the surface
// of the Collider.
//
// The radius r must be greater than zero.
func NewColliderSolidHollow(collider Collider, r float64) *ColliderSolid {
	res := NewColliderSolid(collider)
	p := Coord3D{r, r, r}
	res.min = res.min.Sub(p)
	res.max = res.max.Add(p)
	res.hollowRadius = r
	return res
}

func (c *ColliderSolid) Min() Coord3D {
	return c.min
}

func (c *ColliderSolid) Max() Coord3D {
	return c.max
}

func (c *ColliderSolid) Contains(p Coord3D) bool {
	if c.hollowRadius > 0 {
		return c.collider.SphereCollision(p, c.hollowRadius)
	}
	return c.collider.RayCollisions(&Ray{
		Origin:    p,
		Direction: c.direction,
	}, nil)%2 == 1
}

type boundCacheSolid struct {
	min Coord3D
	max Coord3D
	s   Solid
}

// CacheSolidBounds creates a Solid that has a cached
// version of the solid's boundary coordinates.
//
// The solid also explicitly checks that points are inside
// the boundary before passing them off to s.
func CacheSolidBounds(s Solid) Solid {
	return boundCacheSolid{
		min: s.Min(),
		max: s.Max(),
		s:   s,
	}
}

func (b boundCacheSolid) Min() Coord3D {
	return b.min
}

func (b boundCacheSolid) Max() Coord3D {
	return b.max
}

func (b boundCacheSolid) Contains(c Coord3D) bool {
	return InBounds(b, c) && b.s.Contains(c)
}
