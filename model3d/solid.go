// Generated from templates/solid.template

package model3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
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

	Contains(p Coord3D) bool
}

type funcSolid struct {
	min Coord3D
	max Coord3D
	f   func(c Coord3D) bool
}

// FuncSolid creates a Solid from a function.
//
// If the bounds are invalid, FuncSolid() will panic().
// In particular, max must be no less than min, and all
// floating-point values must be finite numbers.
func FuncSolid(min, max Coord3D, f func(Coord3D) bool) Solid {
	if !BoundsValid(NewRect(min, max)) {
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
func CheckedFuncSolid(min, max Coord3D, f func(Coord3D) bool) Solid {
	return FuncSolid(min, max, func(c Coord3D) bool {
		return c.Min(min) == min && c.Max(max) == max && f(c)
	})
}

func (f *funcSolid) Min() Coord3D {
	return f.min
}

func (f *funcSolid) Max() Coord3D {
	return f.max
}

func (f *funcSolid) Contains(c Coord3D) bool {
	return f.f(c)
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
		min = min.Min(s.Min())
	}
	return min
}

func (j JoinedSolid) Max() Coord3D {
	max := j[0].Max()
	for _, s := range j[1:] {
		max = max.Max(s.Max())
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

// Optimize creates a version of the solid that is faster
// when joining a large number of smaller solids.
func (j JoinedSolid) Optimize() Solid {
	grouped := append([]Solid{}, j...)
	GroupBounders(grouped)
	return groupedSolidsToSolid(grouped)
}

func groupedSolidsToSolid(s []Solid) Solid {
	if len(s) == 1 {
		return CacheSolidBounds(s[0].(Solid))
	}
	firstHalf := s[:len(s)/2]
	secondHalf := s[len(s)/2:]
	return CacheSolidBounds(JoinedSolid{
		groupedSolidsToSolid(firstHalf),
		groupedSolidsToSolid(secondHalf),
	})
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

// StackSolids joins solids together and moves each solid
// after the first so that the lowest Z value of its
// bounding box collides with the highest Z value of the
// previous solid's bounding box.
// In other words, the solids are stacked on top of each
// other along the Z axis.
func StackSolids(s ...Solid) Solid {
	result := make(JoinedSolid, len(s))
	result[0] = s[0]
	lastMax := s[0].Max().Z
	for i := 1; i < len(s); i++ {
		delta := lastMax - s[i].Min().Z
		result[i] = TransformSolid(&Translate{Offset: Z(delta)}, s[i])
		lastMax = result[i].Max().Z
	}
	return result
}

// A StackedSolid is like a JoinedSolid, but the solids
// after the first are moved so that the lowest Z value of
// their bounding box collides with the highest Z value of
// the previous solid.
// In other words, the solids are stacked on top of each
// other along the Z axis.
//
// This API is deprecated in favor of the StackSolids()
// function.
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
		if solid.Contains(c.Sub(Z(delta))) {
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
	collider Collider
	min      Coord3D
	max      Coord3D
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
	insetVec := XYZ(inset, inset, inset)
	min := c.Min().Add(insetVec)
	max := min.Max(c.Max().Sub(insetVec))
	return &ColliderSolid{collider: c, min: min, max: max, inset: inset}
}

// NewColliderSolidHollow creates a ColliderSolid that
// only reports containment around the edges.
func NewColliderSolidHollow(c Collider, r float64) *ColliderSolid {
	insetVec := XYZ(r, r, r)
	min := c.Min().Sub(insetVec)
	max := c.Max().Add(insetVec)
	return &ColliderSolid{collider: c, min: min, max: max, radius: r}
}

// Min gets the minimum of the bounding box.
func (c *ColliderSolid) Min() Coord3D {
	return c.min
}

// Max gets the maximum of the bounding box.
func (c *ColliderSolid) Max() Coord3D {
	return c.max
}

// Contains checks if coord is in the solid.
func (c *ColliderSolid) Contains(coord Coord3D) bool {
	if !InBounds(c, coord) {
		return false
	}
	if c.radius != 0 {
		return c.collider.SphereCollision(coord, c.radius)
	}
	return ColliderContains(c.collider, coord, c.inset)
}

// ForceSolidBounds creates a new solid that reports the
// exact bounds given by min and max.
//
// Points outside of these bounds will be removed from s,
// but otherwise s is preserved.
func ForceSolidBounds(s Solid, min, max Coord3D) Solid {
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
	return CheckedFuncSolid(
		min.AddScalar(-radius),
		max.AddScalar(radius),
		func(c Coord3D) bool {
			var closestDists [2]float64
			for i, s := range sdfs {
				d := s.SDF(c)
				if d > 0 {
					return true
				}
				if i < 2 {
					closestDists[i] = d
					if i == 2 {
						if closestDists[1] > closestDists[0] {
							closestDists[0], closestDists[1] = closestDists[1], closestDists[0]
						}
					}
				} else {
					if d >= closestDists[0] {
						closestDists[1] = closestDists[0]
						closestDists[0] = d
					} else if d > closestDists[1] {
						closestDists[1] = d
					}
				}
			}

			d1 := math.Max(0, closestDists[0]+radius)
			d2 := math.Max(0, closestDists[1]+radius)
			return d1*d1+d2*d2 > radius*radius
		},
	)
}

// SmoothJoinV2 is like SmoothJoin, but uses surface
// normals to improve results for SDFs that intersect at
// obtuse angles.
func SmoothJoinV2(radius float64, sdfs ...NormalSDF) Solid {
	min := sdfs[0].Min()
	max := sdfs[0].Max()
	for _, s := range sdfs[1:] {
		min = min.Min(s.Min())
		max = max.Max(s.Max())
	}
	return CheckedFuncSolid(
		min.AddScalar(-radius),
		max.AddScalar(radius),
		func(c Coord3D) bool {
			var closestDists [2]float64
			var closestNormals [2]Coord3D
			for i, s := range sdfs {
				p, d := s.NormalSDF(c)
				if d > 0 {
					return true
				}
				if i < 2 {
					closestNormals[i] = p
					closestDists[i] = d
					if i == 2 {
						if closestDists[1] > closestDists[0] {
							closestDists[0], closestDists[1] = closestDists[1], closestDists[0]
							closestNormals[0], closestNormals[1] = closestNormals[1], closestNormals[0]
						}
					}
				} else {
					if d >= closestDists[0] {
						closestDists[1] = closestDists[0]
						closestNormals[1] = closestNormals[0]
						closestDists[0] = d
						closestNormals[0] = p
					} else if d > closestDists[1] {
						closestDists[1] = d
						closestNormals[1] = p
					}
				}
			}

			cosTheta := math.Abs(closestNormals[0].Dot(closestNormals[1]))
			r := radius * math.Sqrt(1-cosTheta*cosTheta)
			d1 := math.Max(0, closestDists[0]+r)
			d2 := math.Max(0, closestDists[1]+r)
			return d1*d1+d2*d2 > r*r
		},
	)
}

// SDFToSolid creates a Solid which is true inside the SDF.
//
// If the outset argument is non-zero, it is the extra
// distance outside the SDF that is considered inside the
// solid. It can also be negative to inset the solid.
func SDFToSolid(s SDF, outset float64) Solid {
	return CheckedFuncSolid(
		s.Min().AddScalar(-outset),
		s.Max().AddScalar(outset),
		func(c Coord3D) bool {
			return s.SDF(c) > -outset
		},
	)
}

// ProfileSolid turns a 2D solid into a 3D solid by
// elongating the 2D solid along the Z axis.
func ProfileSolid(solid2d model2d.Solid, minZ, maxZ float64) Solid {
	min, max := solid2d.Min(), solid2d.Max()
	min3d := XYZ(min.X, min.Y, minZ)
	max3d := XYZ(max.X, max.Y, maxZ)
	return CheckedFuncSolid(min3d, max3d, func(c Coord3D) bool {
		return solid2d.Contains(c.XY())
	})
}

// CrossSectionSolid creates a 2D cross-section of a 3D
// solid as a 2D solid.
//
// The axis is 0, 1, or 2 for X, Y, or Z respectively.
// The axisValue is the value for the axis at which a
// plane is constructed.
func CrossSectionSolid(solid Solid, axis int, axisValue float64) model2d.Solid {
	var to2D func(Coord3D) Coord2D
	var to3D func(Coord2D) Coord3D
	if axis == 0 {
		to2D = func(c Coord3D) Coord2D {
			return c.YZ()
		}
		to3D = func(c Coord2D) Coord3D {
			return XYZ(axisValue, c.X, c.Y)
		}
	} else if axis == 1 {
		to2D = func(c Coord3D) Coord2D {
			return c.XZ()
		}
		to3D = func(c Coord2D) Coord3D {
			return XYZ(c.X, axisValue, c.Y)
		}
	} else {
		to2D = func(c Coord3D) Coord2D {
			return c.XY()
		}
		to3D = func(c Coord2D) Coord3D {
			return XYZ(c.X, c.Y, axisValue)
		}
	}
	return model2d.CheckedFuncSolid(
		to2D(solid.Min()),
		to2D(solid.Max()),
		func(c Coord2D) bool {
			return solid.Contains(to3D(c))
		},
	)
}

// RevolveSolid rotates a 2D solid around an axis to
// create a 3D solid.
// The y-axis of the 2D solid is extended along the axis
// of revolution, while the x-axis is used as a radius.
//
// The 2D solid should either be symmetrical around the
// axis, or be empty on one side of the axis.
// Either way, the union of both reflections of the 2D
// solid is used.
func RevolveSolid(solid model2d.Solid, axis Coord3D) Solid {
	axis = axis.Normalize()
	min, max := solid.Min(), solid.Max()
	maxRadius := math.Max(math.Abs(min.X), math.Abs(max.X))
	cylinder := &Cylinder{
		P1:     axis.Scale(min.Y),
		P2:     axis.Scale(max.Y),
		Radius: maxRadius,
	}
	return CheckedFuncSolid(
		cylinder.Min(),
		cylinder.Max(),
		func(c Coord3D) bool {
			x := c.ProjectOut(axis).Norm()
			y := axis.Dot(c)
			return solid.Contains(model2d.XY(x, y))
		},
	)
}

// A SolidMux computes many solid values in parallel and
// returns a bitmap of containment for each solid.
//
// This uses a BVH to efficiently check the containment of
// many solids without explicitly having to check every
// single solid's Contains() methods.
type SolidMux struct {
	bbox        Rect
	totalSolids int
	leaf        Solid
	leafIndex   int
	children    [2]*SolidMux
}

// NewSolidMux creates a SolidMux using the ordered list of
// solids provided as arguments.
func NewSolidMux(solids []Solid) *SolidMux {
	if len(solids) == 0 {
		return &SolidMux{}
	}
	// Group Rects instead of Solids so that we know
	// we can use the bounder as a key in a map to
	// track the index.
	bounders := make([]*Rect, len(solids))
	bounderToIndex := map[*Rect]int{}
	for i, s := range solids {
		bounders[i] = BoundsRect(s)
		bounderToIndex[bounders[i]] = i
	}
	GroupBounders(bounders)
	groupedSolids := make([]Solid, len(solids))
	indices := make([]int, len(solids))
	for i, b := range bounders {
		idx := bounderToIndex[b]
		groupedSolids[i] = solids[idx]
		indices[i] = idx
	}
	return groupedSolidsToSolidMux(groupedSolids, indices)
}

func groupedSolidsToSolidMux(solids []Solid, indices []int) *SolidMux {
	if len(solids) == 1 {
		return &SolidMux{
			bbox:        *BoundsRect(solids[0]),
			totalSolids: 1,
			leaf:        solids[0],
			leafIndex:   indices[0],
		}
	}
	splitIdx := len(solids) / 2
	return &SolidMux{
		bbox:        *BoundsRect(JoinedSolid(solids)),
		totalSolids: len(solids),
		children: [2]*SolidMux{
			groupedSolidsToSolidMux(solids[:splitIdx], indices[:splitIdx]),
			groupedSolidsToSolidMux(solids[splitIdx:], indices[splitIdx:]),
		},
	}
}

func (s *SolidMux) Min() Coord3D {
	return s.bbox.MinVal
}

func (s *SolidMux) Max() Coord3D {
	return s.bbox.MaxVal
}

func (s *SolidMux) Contains(c Coord3D) bool {
	if !s.bbox.Contains(c) || s.totalSolids == 0 {
		return false
	}
	if s.totalSolids == 1 {
		return s.leaf.Contains(c)
	} else {
		for _, ch := range s.children {
			if ch.Contains(c) {
				return true
			}
		}
		return false
	}
}

func (s *SolidMux) AllContains(c Coord3D) []bool {
	res := make([]bool, s.totalSolids)
	s.allContains(c, res)
	return res
}

func (s *SolidMux) allContains(c Coord3D, out []bool) {
	if !s.bbox.Contains(c) || s.totalSolids == 0 {
		return
	}
	if s.totalSolids == 1 {
		out[s.leafIndex] = s.leaf.Contains(c)
	} else {
		for _, ch := range s.children {
			ch.allContains(c, out)
		}
	}
}
