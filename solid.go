package model3d

import (
	"math"
	"math/rand"
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

// A RectSolid is a Solid that fills an axis-aligned
// rectangular volume.
type RectSolid struct {
	MinVal Coord3D
	MaxVal Coord3D
}

func (r *RectSolid) Min() Coord3D {
	return r.MinVal
}

func (r *RectSolid) Max() Coord3D {
	return r.MaxVal
}

func (r *RectSolid) Contains(c Coord3D) bool {
	return c.Min(r.MinVal) == r.MinVal && c.Max(r.MaxVal) == r.MaxVal
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
	minCenter := c.P1.Min(c.P2)
	axis := c.P2.Sub(c.P1)
	minOffsets := (Coord3D{
		circleAxisBound(0, axis, -1),
		circleAxisBound(1, axis, -1),
		circleAxisBound(2, axis, -1),
	}).Scale(c.Radius)
	return minCenter.Add(minOffsets)
}

func (c *CylinderSolid) Max() Coord3D {
	maxCenter := c.P1.Max(c.P2)
	axis := c.P2.Sub(c.P1)
	maxOffsets := (Coord3D{
		circleAxisBound(0, axis, 1),
		circleAxisBound(1, axis, 1),
		circleAxisBound(2, axis, 1),
	}).Scale(c.Radius)
	return maxCenter.Add(maxOffsets)
}

func (c *CylinderSolid) Contains(p Coord3D) bool {
	diff := c.P1.Sub(c.P2)
	direction := diff.Normalize()
	frac := p.Sub(c.P2).Dot(direction)
	if frac < 0 || frac > diff.Norm() {
		return false
	}
	projection := c.P2.Add(direction.Scale(frac))
	return projection.Dist(p) <= c.Radius
}

// circleAxisBound gets the furthest along an axis
// (x, y, or z) you can move while remaining inside a unit
// circle which is normal to a given vector.
// The sign argument indicates if we are moving in the
// negative or positive direction.
func circleAxisBound(axis int, normal Coord3D, sign float64) float64 {
	var arr [3]float64
	arr[axis] = sign
	proj := newCoord3DArray(arr).ProjectOut(normal)

	// Care taken to deal with numerical issues.
	proj = proj.Scale(1 / (proj.Norm() + 1e-8))
	return sign * (math.Abs(proj.array()[axis]) + 1e-8)
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
	b1, b2 := t.Axis.OrthoBasis()
	centered := c.Sub(t.Center)

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
	})%2 == 1
}

// SolidNormal approximates the shortest vector that
// goes from point c to the boundary of a solid s.
// It is meant for cases when c is close to the boundary
// of s; otherwise, this algorithm is vastly inefficient.
//
// The epsilon argument provides an initial guess as to
// how far the point is from the surface boundary.
// The smaller the epsilon, the higher the accuracy may be
// when the surface has small crevices. Larger epsilons
// may miss small crevices, but may result in faster
// computation when the surface is smooth and doesn't have
// crevices.
//
// The maxIters argument specifies the maximum amount of
// searching that should be done before the algorithm
// gives up looking for a boundary.
// Each iteration scales and halves the epsilon in order
// to attempt to find the boundary.
// If zero, then epsilon must be exactly correct or else
// the boundary search will fail.
//
// The samplePoints argument specifies how many points to
// sample around c while estimating the normal.
// If zero, a default value is used.
// Higher values result in more accurate results, at the
// expense of more computation.
//
// If the normal could not be computed because no boundary
// could be found, false is returned.
func SolidNormal(s Solid, c Coord3D, epsilon float64, maxIters, samplePoints int) (Coord3D, bool) {
	if samplePoints == 0 {
		// Should be enough to get a rough estimate.
		samplePoints = 40
	}

	isContained := s.Contains(c)
	deltas := make([]Coord3D, samplePoints)
	for i := range deltas {
		deltas[i] = (Coord3D{
			X: rand.NormFloat64(),
			Y: rand.NormFloat64(),
			Z: rand.NormFloat64(),
		}).Normalize()
	}

	computeContainments := func(eps float64) []bool {
		containments := make([]bool, len(deltas))
		for i, delta := range deltas {
			containments[i] = s.Contains(c.Add(delta.Scale(eps)))
		}
		return containments
	}

	foundBoundary := func(containments []bool) bool {
		for _, c := range containments {
			if c != isContained {
				return true
			}
		}
		return false
	}

	unscaledDirection := func(eps float64, containments []bool) Coord3D {
		var sum Coord3D
		var count int
		for i, delta := range deltas {
			if containments[i] != isContained {
				sum = sum.Add(delta)
				count++
			}
		}

		// Normally, we should be able to get a good
		// direction by averaging all of the directions
		// that hit the boundary.
		// However, this is not guaranteed to work.
		avg := sum.Scale(eps / float64(count))
		if s.Contains(c.Add(avg)) != isContained {
			return avg
		}

		// Fall-back on nearest neighbor.
		var nearest Coord3D
		nearestDistance := math.Inf(1)
		for i, delta := range deltas {
			if containments[i] != isContained {
				step := delta.Scale(eps)
				dist := step.Dist(avg)
				if dist < nearestDistance {
					nearestDistance = dist
					nearest = step
				}
			}
		}
		return nearest
	}

	lineSearch := func(dir Coord3D) Coord3D {
		p1 := Coord3D{}
		p2 := dir
		for i := 0; i < samplePoints; i++ {
			m := p1.Mid(p2)
			if s.Contains(c.Add(m)) != isContained {
				p2 = m
			} else {
				p1 = m
			}
		}
		return p1.Mid(p2)
	}

	tryEpsilon := func(eps float64) (Coord3D, bool) {
		containments := computeContainments(eps)
		if foundBoundary(containments) {
			return lineSearch(unscaledDirection(eps, containments)), true
		}
		return Coord3D{}, false
	}

	if normal, ok := tryEpsilon(epsilon); ok {
		return normal, ok
	}

	for i := 0; i < maxIters; i++ {
		x := float64(uint(1) << uint(i+1))
		if normal, ok := tryEpsilon(epsilon / x); ok {
			return normal, ok
		}
		if normal, ok := tryEpsilon(epsilon * float64(i+2)); ok {
			return normal, ok
		}
	}

	return Coord3D{}, false
}
