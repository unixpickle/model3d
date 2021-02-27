// Generated from templates/solid.template

package model2d

import (
	"sort"
)

// A Solid is a boolean function where a value of true
// indicates that a point is part of the solid, and false
// indicates that it is not.
//
// All methods of a Solid are safe for concurrency.
type Solid interface {
	// Contains must always return false outside of the
	// boundaries of the solid.
	Bounder

	Contains(p Coord) bool
}

type funcSolid struct {
	min Coord
	max Coord
	f   func(c Coord) bool
}

// FuncSolid creates a Solid from a function.
//
// If the bounds are invalid, FuncSolid() will panic().
// In particular, max must be no less than min, and all
// floating-point values must be finite numbers.
func FuncSolid(min, max Coord, f func(Coord) bool) Solid {
	if !BoundsValid(&Rect{MinVal: min, MaxVal: max}) {
		panic("invalid bounds")
	}
	return &funcSolid{
		min: min,
		max: max,
		f:   f,
	}
}

// CheckedFuncSolid is like FuncSolid, but it does an
// automatic bounds check before calling f.
func CheckedFuncSolid(min, max Coord, f func(Coord) bool) Solid {
	return FuncSolid(min, max, func(c Coord) bool {
		return c.Min(min) == min && c.Max(max) == max && f(c)
	})
}

func (f *funcSolid) Min() Coord {
	return f.min
}

func (f *funcSolid) Max() Coord {
	return f.max
}

func (f *funcSolid) Contains(c Coord) bool {
	return f.f(c)
}

// A JoinedSolid is a Solid composed of other solids.
type JoinedSolid []Solid

func (j JoinedSolid) Min() Coord {
	min := j[0].Min()
	for _, s := range j[1:] {
		min = min.Min(s.Min())
	}
	return min
}

func (j JoinedSolid) Max() Coord {
	max := j[0].Max()
	for _, s := range j[1:] {
		max = max.Max(s.Max())
	}
	return max
}

func (j JoinedSolid) Contains(c Coord) bool {
	for _, s := range j {
		if s.Contains(c) {
			return true
		}
	}
	return false
}

// Optimize creates a version of the solid that is faster
// when joining a large number of smaller solids.
func (j JoinedSolid) Optimize() Solid {
	bounders := make([]Bounder, len(j))
	for i, s := range j {
		bounders[i] = s
	}
	GroupBounders(bounders)
	return groupedBoundersToSolid(bounders)
}

func groupedBoundersToSolid(bs []Bounder) Solid {
	if len(bs) == 1 {
		return CacheSolidBounds(bs[0].(Solid))
	}
	firstHalf := bs[:len(bs)/2]
	secondHalf := bs[len(bs)/2:]
	return CacheSolidBounds(JoinedSolid{
		groupedBoundersToSolid(firstHalf),
		groupedBoundersToSolid(secondHalf),
	})
}

// SubtractedSolid is a Solid consisting of all the points
// in Positive which are not in Negative.
type SubtractedSolid struct {
	Positive Solid
	Negative Solid
}

func (s *SubtractedSolid) Min() Coord {
	return s.Positive.Min()
}

func (s *SubtractedSolid) Max() Coord {
	return s.Positive.Max()
}

func (s *SubtractedSolid) Contains(c Coord) bool {
	return s.Positive.Contains(c) && !s.Negative.Contains(c)
}

// IntersectedSolid is a Solid containing the intersection
// of one or more Solids.
type IntersectedSolid []Solid

func (i IntersectedSolid) Min() Coord {
	bound := i[0].Min()
	for _, s := range i[1:] {
		bound = bound.Max(s.Min())
	}
	return bound
}

func (i IntersectedSolid) Max() Coord {
	bound := i[0].Max()
	for _, s := range i[1:] {
		bound = bound.Min(s.Max())
	}
	// Prevent negative area.
	return bound.Max(i.Min())
}

func (i IntersectedSolid) Contains(c Coord) bool {
	for _, s := range i {
		if !s.Contains(c) {
			return false
		}
	}
	return true
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
	collider Collider
	min      Coord
	max      Coord
	inset    float64
	radius   float64
}

// NewColliderSolid creates a basic ColliderSolid.
func NewColliderSolid(c Collider) *ColliderSolid {
	return &ColliderSolid{collider: c, min: c.Min(), max: c.Max()}
}

// NewColliderSolidInset creates a ColliderSolid that only
// reports containment at some distance from the surface.
//
// If inset is negative, then the solid is outset from the
// collider.
func NewColliderSolidInset(c Collider, inset float64) *ColliderSolid {
	insetVec := XY(inset, inset)
	min := c.Min().Add(insetVec)
	max := min.Max(c.Max().Sub(insetVec))
	return &ColliderSolid{collider: c, min: min, max: max, inset: inset}
}

// NewColliderSolidHollow creates a ColliderSolid that
// only reports containment around the edges.
func NewColliderSolidHollow(c Collider, r float64) *ColliderSolid {
	insetVec := XY(r, r)
	min := c.Min().Sub(insetVec)
	max := c.Max().Add(insetVec)
	return &ColliderSolid{collider: c, min: min, max: max, radius: r}
}

// Min gets the minimum of the bounding box.
func (c *ColliderSolid) Min() Coord {
	return c.min
}

// Max gets the maximum of the bounding box.
func (c *ColliderSolid) Max() Coord {
	return c.max
}

// Contains checks if coord is in the solid.
func (c *ColliderSolid) Contains(coord Coord) bool {
	if !InBounds(c, coord) {
		return false
	}
	if c.radius != 0 {
		return c.collider.CircleCollision(coord, c.radius)
	}
	return ColliderContains(c.collider, coord, c.inset)
}

// ForceSolidBounds creates a new solid that reports the
// exact bounds given by min and max.
//
// Points outside of these bounds will be removed from s,
// but otherwise s is preserved.
func ForceSolidBounds(s Solid, min, max Coord) Solid {
	return CheckedFuncSolid(min, max, s.Contains)
}

// CacheSolidBounds creates a Solid that has a cached
// version of the solid's boundary coordinates.
//
// The solid also explicitly checks that points are inside
// the boundary before passing them off to s.
func CacheSolidBounds(s Solid) Solid {
	return ForceSolidBounds(s, s.Min(), s.Max())
}

type smoothJoin struct {
	min    Coord
	max    Coord
	sdfs   []SDF
	radius float64
}

// SmoothJoin joins the SDFs into a union Solid and
// smooths the intersections using a given smoothing
// radius.
//
// If the radius is 0, it is equivalent to turning the
// SDFs directly into solids and then joining them.
func SmoothJoin(radius float64, sdfs ...SDF) Solid {
	min := sdfs[0].Min()
	max := sdfs[0].Max()
	for _, s := range sdfs[1:] {
		min = min.Min(s.Min())
		max = max.Max(s.Max())
	}
	return &smoothJoin{
		min:    min,
		max:    max,
		sdfs:   sdfs,
		radius: radius,
	}
}

func (s *smoothJoin) Min() Coord {
	return s.min
}

func (s *smoothJoin) Max() Coord {
	return s.max
}

func (s *smoothJoin) Contains(c Coord) bool {
	if !InBounds(s, c) {
		return false
	}

	var distances []float64
	for _, s := range s.sdfs {
		d := s.SDF(c)
		if d > 0 {
			return true
		}
		distances = append(distances, -d)
	}
	sort.Float64s(distances)

	if distances[1] > s.radius {
		return false
	}
	d1 := s.radius - distances[0]
	d2 := s.radius - distances[1]
	return d1*d1+d2*d2 > s.radius*s.radius
}

func BitmapToSolid(b *Bitmap) Solid {
	return CheckedFuncSolid(Coord{}, XY(float64(b.Width), float64(b.Height)), func(c Coord) bool {
		return b.Get(int(c.X), int(c.Y))
	})
}
