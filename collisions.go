package model3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/essentials"
)

// A Ray is a line originating at a point and extending
// infinitely in some (positive) direction.
type Ray struct {
	Origin    Coord3D
	Direction Coord3D
}

// RayCollision is a point where a ray intersects a
// surface.
type RayCollision struct {
	// The amount of the ray direction to add to the ray
	// origin to hit the point in question.
	//
	// The scale should be non-negative.
	Scale float64

	// The normal pointing outward from the surface at the
	// point of collision.
	Normal Coord3D

	// Extra contains additional, implementation-specific
	// information about the collision.
	//
	// For an example, see TriangleCollision.
	Extra interface{}
}

// TriangleCollision is triangle-specific collision
// information.
type TriangleCollision struct {
	// The triangle that reported the collision.
	Triangle *Triangle

	// Barycentric coordinates in the triangle,
	// corresponding to the corners.
	Barycentric [3]float64
}

// A Collider is a surface which can detect intersections
// with linear rays and spheres.
type Collider interface {
	Bounder

	// RayCollisions enumerates the collisions with a ray.
	// It returns the total number of collisions.
	//
	// f may be nil, in which case this is simply used for
	// counting.
	RayCollisions(r *Ray, f func(RayCollision)) int

	// FirstRayCollision gets the ray collision with the
	// lowest scale.
	//
	// The second return value is false if no collisions
	// were found.
	FirstRayCollision(r *Ray) (collision RayCollision, collides bool)

	// SphereCollision checks if the collider touches a
	// sphere with origin c and radius r.
	SphereCollision(c Coord3D, r float64) bool
}

// A TriangleCollider is like a Collider, but it can also
// check if a triangle intersects the surface.
type TriangleCollider interface {
	Collider

	// TriangleCollisions gets all of the segments on the
	// surface which intersect the triangle t.
	TriangleCollisions(t *Triangle) []Segment
}

// MeshToCollider creates an efficient TriangleCollider
// out of a mesh.
func MeshToCollider(m *Mesh) TriangleCollider {
	tris := m.TriangleSlice()
	GroupTriangles(tris)
	return GroupedTrianglesToCollider(tris)
}

// GroupTriangles sorts the triangle slice in a special
// way for GroupedTrianglesToCollider().
// This can be used to prepare models for being turned
// into a collider efficiently.
func GroupTriangles(tris []*Triangle) {
	groupTriangles(sortTriangles(tris), 0, tris)
}

func groupTriangles(sortedTris [3][]*flaggedTriangle, axis int, output []*Triangle) {
	numTris := len(sortedTris[axis])
	if numTris == 1 {
		output[0] = sortedTris[axis][0].T
		return
	} else if numTris == 0 {
		return
	}

	midIdx := numTris / 2
	for i, t := range sortedTris[axis][:] {
		t.Flag = i < midIdx
	}

	separated := [3][]*flaggedTriangle{}
	separated[axis] = sortedTris[axis]

	for newAxis := 0; newAxis < 3; newAxis++ {
		if newAxis == axis {
			continue
		}
		sep := make([]*flaggedTriangle, numTris)
		idx0 := 0
		idx1 := numTris / 2
		for _, t := range sortedTris[newAxis] {
			if t.Flag {
				sep[idx0] = t
				idx0++
			} else {
				sep[idx1] = t
				idx1++
			}
		}
		separated[newAxis] = sep
	}

	groupTriangles([3][]*flaggedTriangle{
		separated[0][:midIdx],
		separated[1][:midIdx],
		separated[2][:midIdx],
	}, (axis+1)%3, output[:midIdx])

	groupTriangles([3][]*flaggedTriangle{
		separated[0][midIdx:],
		separated[1][midIdx:],
		separated[2][midIdx:],
	}, (axis+1)%3, output[midIdx:])
}

func sortTriangles(tris []*Triangle) [3][]*flaggedTriangle {
	ts := make([]*flaggedTriangle, len(tris))
	for i, t := range tris {
		ts[i] = &flaggedTriangle{T: t}
	}

	var result [3][]*flaggedTriangle
	for axis := range result {
		tsCopy := append([]*flaggedTriangle{}, ts...)
		if axis == 0 {
			essentials.VoodooSort(tsCopy, func(i, j int) bool {
				return tsCopy[i].T[0].X < tsCopy[j].T[0].X
			})
		} else if axis == 1 {
			essentials.VoodooSort(tsCopy, func(i, j int) bool {
				return tsCopy[i].T[0].Y < tsCopy[j].T[0].Y
			})
		} else {
			essentials.VoodooSort(tsCopy, func(i, j int) bool {
				return tsCopy[i].T[0].Z < tsCopy[j].T[0].Z
			})
		}
		result[axis] = tsCopy
	}
	return result
}

// GroupedTrianglesToCollider converts a mesh of triangles
// into a TriangleCollider.
//
// The triangles should be sorted by GroupTriangles.
// Otherwise, the resulting Collider may not be efficient.
func GroupedTrianglesToCollider(tris []*Triangle) TriangleCollider {
	if len(tris) == 1 {
		return tris[0]
	} else if len(tris) == 0 {
		return nullCollider{}
	}
	midIdx := len(tris) / 2
	c1 := GroupedTrianglesToCollider(tris[:midIdx])
	c2 := GroupedTrianglesToCollider(tris[midIdx:])
	return joinedTriangleCollider{NewJoinedCollider([]Collider{c1, c2})}
}

// FirstRayCollision gets the ray collision if there is
// one.
//
// The Extra field is a *TriangleCollision.
func (t *Triangle) FirstRayCollision(r *Ray) (RayCollision, bool) {
	info, scale := t.rayCollision(r)
	if info != nil && scale >= 0 {
		return RayCollision{Scale: scale, Normal: t.Normal(), Extra: info}, true
	} else {
		return RayCollision{}, false
	}
}

// RayCollisions calls f (if non-nil) with a collision (if
// applicable) and returns the collisions count (0 or 1).
//
// The Extra field is a *TriangleCollision.
func (t *Triangle) RayCollisions(r *Ray, f func(RayCollision)) int {
	info, scale := t.rayCollision(r)
	if info == nil || scale < 0 {
		return 0
	}
	if f != nil {
		f(RayCollision{Scale: scale, Normal: t.Normal(), Extra: info})
	}
	return 1
}

func (t *Triangle) rayCollision(r *Ray) (tc *TriangleCollision, scale float64) {
	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])
	matrix := Matrix3{
		v1.X, v2.X, r.Direction.X,
		v1.Y, v2.Y, r.Direction.Y,
		v1.Z, v2.Z, r.Direction.Z,
	}
	if math.Abs(matrix.Det()) < 1e-8*t.Area()*r.Direction.Norm() {
		return nil, 0
	}
	matrix.InvertInPlace()
	result := matrix.MulColumn(r.Origin.Sub(t[0]))
	barycentric := [3]float64{1 - (result.X + result.Y), result.X, result.Y}
	if barycentric[0] >= 0 && barycentric[1] >= 0 && barycentric[2] >= 0 {
		tc = &TriangleCollision{
			Triangle:    t,
			Barycentric: barycentric,
		}
	}
	return tc, -result.Z
}

// SphereCollision checks if any part of the triangle is
// within the sphere.
func (t *Triangle) SphereCollision(c Coord3D, r float64) bool {
	for _, p := range t {
		if p.Dist(c) < r {
			return true
		}
	}
	for i := 0; i < 3; i++ {
		p1 := t[i]
		p2 := t[(i+1)%3]
		if segmentEntersSphere(p1, p2, c, r) {
			return true
		}
	}
	ray := &Ray{
		Origin:    c,
		Direction: t.Normal(),
	}
	info, frac := t.rayCollision(ray)
	return info != nil && math.Abs(frac) < r
}

func segmentEntersSphere(p1, p2, c Coord3D, r float64) bool {
	v := p2.Sub(p1)
	frac := (c.Dot(v) - p1.Dot(v)) / v.Dot(v)
	closest := p1.Add(v.Scale(frac))
	return frac >= 0 && frac <= 1 && closest.Dist(c) < r
}

// TriangleCollisions finds the segment where t intersects
// t1. If no segment exists, an empty slice is returned.
//
// If t and t1 are (nearly) co-planar, no collisions are
// reported, since small numerical differences can have a
// major impact.
func (t *Triangle) TriangleCollisions(t1 *Triangle) []Segment {
	if t.inCommon(t1) > 1 {
		// No way to be intersecting unless we are co-planar.
		return nil
	}

	// Check if the triangles are (nearly) co-planar.
	n1 := t.Normal()
	n2 := t1.Normal()
	if math.Abs(n1.Dot(n2)) > 1-1e-8 {
		return nil
	}

	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])
	v3 := t1[1].Sub(t1[0])
	v4 := t1[2].Sub(t1[0])

	// Intersections happen at solutions to this system:
	//
	//     a*v1+b*v2+t[0] = c*v3+d*v4+t1[0]
	//     a, b, c, d >= 0
	//     a+b <= 1
	//     c+d <= 1
	//
	// We can rewrite the first equation as follows:
	//
	//     Ax = t1[0] - t[0]
	//     where A = [v1 v2 -v3 -v4] (a matrix of columns)
	//     and x = [a; b; c; d] (a column matrix)
	//
	// The solutions to this equation are of the form:
	//
	//     o + t*d
	//
	// Where o is any solution, t is a scalar, and d is a
	// vector in the direction of the null-space of A.
	//
	// To compute the final intersection, we find the
	// intervals of t for which the other constraints are
	// satisfied.

	// Find the first three components of o, a combination
	// of v1, v2, v3, v4 where the planes intersect.
	m1 := NewMatrix3Columns(v1, v2, v3.Scale(-1))
	m2 := NewMatrix3Columns(v1, v2, v4.Scale(-1))
	matA := m1
	if math.Abs(m2.Det()) > math.Abs(m1.Det()) {
		// Using m2 may be more numerically stable.
		// Helps in the case that either v3 or v4 is on
		// the plane of triangle t.
		// Equivalent to a column swap during gaussian
		// elimination to find any solution.
		matA = m2
		v3, v4 = v4, v3
	}
	invA := matA.Inverse()
	o := invA.MulColumn(t1[0].Sub(t[0]))

	// Find the first three components of d, a combination
	// of v1, v2, v3, v4 that goes along the intersection
	// of the two planes (i.e. is the null-space of A).
	// The final component is 1.
	d := invA.MulColumn(v4)

	// A function which solves for a range of t values
	// such that o+t*d >= 0 and (o+t*d)*[1; 1] <= 1.
	// Returns min, max.
	// Used for finding the interval of t for each of the
	// two triangles.
	findContainedRange := func(o1, o2, d1, d2 float64) (float64, float64) {
		tMin := math.Inf(-1)
		tMax := math.Inf(1)

		// Rewriting the second constraint, we get
		// o*[1; 1] + t*d*[1; 1] <= 1.
		// Let sumO = o*[1; 1] and sumD = d*[1; 1].
		// Thus,
		//     t <= (1 - sumO)/sumD  where sumD > 0
		//     t >= (1 - sumO)/sumD  where sumD < 0
		sumO := o1 + o2
		sumD := d1 + d2
		if sumD != 0 {
			bound := (1 - sumO) / sumD
			if sumD > 0 {
				tMax = bound
			} else {
				tMin = bound
			}
		} else if sumO > 1 {
			// There is no way to satisfy the second constraint.
			return 0, 0
		}

		updateFirstConstraint := func(o, d float64) {
			// Given that o+t*d >= 0,
			//     t >= -o/d  where d > 0
			//     t <= -o/d  where d < 0
			if d == 0 {
				if o < 0 {
					// Impossible to satisfy.
					tMin, tMax = 0, 0
				}
			} else {
				bound := -o / d
				if d < 0 {
					tMax = math.Min(tMax, bound)
				} else {
					tMin = math.Max(tMin, bound)
				}
			}
		}
		updateFirstConstraint(o1, d1)
		updateFirstConstraint(o2, d2)

		return tMin, tMax
	}

	min1, max1 := findContainedRange(o.X, o.Y, d.X, d.Y)
	if min1 >= max1 {
		return nil
	}

	min2, max2 := findContainedRange(o.Z, 0, d.Z, 1)
	if min2 >= max2 {
		return nil
	}

	min := math.Max(min1, min2)
	max := math.Min(max1, max2)
	if min >= max {
		return nil
	}

	// Get a Euclidean coordinate for a given value of t
	// in the collision equations.
	collisionPoint := func(time float64) Coord3D {
		a := o.X + d.X*time
		b := o.Y + d.Y*time
		return t[0].Add(v1.Scale(a)).Add(v2.Scale(b))
	}

	p1, p2 := collisionPoint(min), collisionPoint(max)
	dist := p1.Dist(p2)
	if dist < v1.Norm()*1e-8 && dist < v2.Norm()*1e-8 {
		// Don't report collisions at a vertex.
		// This can happen due to rounding error.
		return nil
	}
	return []Segment{NewSegment(p1, p2)}
}

// A JoinedCollider wraps multiple other Colliders and
// only passes along rays and spheres that enter their
// combined bounding box.
type JoinedCollider struct {
	min       Coord3D
	max       Coord3D
	colliders []Collider
}

// NewJoinedCollider creates a JoinedCollider which
// combines one or more other colliders.
func NewJoinedCollider(other []Collider) *JoinedCollider {
	res := &JoinedCollider{
		colliders: other,
		min:       other[0].Min(),
		max:       other[0].Max(),
	}
	for _, c := range other[1:] {
		res.min = res.min.Min(c.Min())
		res.max = res.max.Max(c.Max())
	}
	return res
}

func (j *JoinedCollider) Min() Coord3D {
	return j.min
}

func (j *JoinedCollider) Max() Coord3D {
	return j.max
}

func (j *JoinedCollider) RayCollisions(r *Ray, f func(RayCollision)) int {
	if !j.rayCollidesWithBounds(r) {
		return 0
	}

	var count int
	for _, c := range j.colliders {
		count += c.RayCollisions(r, f)
	}
	return count
}

func (j *JoinedCollider) FirstRayCollision(r *Ray) (RayCollision, bool) {
	if !j.rayCollidesWithBounds(r) {
		return RayCollision{}, false
	}
	var closest RayCollision
	var anyCollides bool
	for _, c := range j.colliders {
		if collision, collides := c.FirstRayCollision(r); collides {
			if collision.Scale < closest.Scale || !anyCollides {
				closest = collision
				anyCollides = true
			}
		}
	}
	return closest, anyCollides
}

func (j *JoinedCollider) SphereCollision(center Coord3D, r float64) bool {
	if !sphereTouchesBounds(center, r, j.min, j.max) {
		return false
	}

	for _, c := range j.colliders {
		if c.SphereCollision(center, r) {
			return true
		}
	}
	return false
}

func (j *JoinedCollider) rayCollidesWithBounds(r *Ray) bool {
	minFrac, maxFrac := rayCollisionWithBounds(r, j.min, j.max)
	return maxFrac >= minFrac && maxFrac >= 0
}

type joinedTriangleCollider struct {
	*JoinedCollider
}

func (j joinedTriangleCollider) TriangleCollisions(t *Triangle) []Segment {
	min := t.Min().Max(j.min)
	max := t.Max().Min(j.max)
	if min.X > max.X || min.Y > max.Y || min.Z > max.Z {
		return nil
	}

	var res []Segment
	for _, c := range j.colliders {
		res = append(res, c.(TriangleCollider).TriangleCollisions(t)...)
	}
	return res
}

type flaggedTriangle struct {
	T    *Triangle
	Flag bool
}

type nullCollider struct{}

func (n nullCollider) Min() Coord3D {
	return Coord3D{}
}

func (n nullCollider) Max() Coord3D {
	return Coord3D{}
}

func (n nullCollider) RayCollisions(r *Ray, float32 func(RayCollision)) int {
	return 0
}

func (n nullCollider) FirstRayCollision(r *Ray) (RayCollision, bool) {
	return RayCollision{}, false
}

func (n nullCollider) SphereCollision(c Coord3D, r float64) bool {
	return false
}

func (n nullCollider) TriangleCollisions(t *Triangle) []Segment {
	return nil
}

// A SolidCollider approximates the behavior of a Collider
// based on nothing but a Solid.
type SolidCollider struct {
	Solid Solid

	// Epsilon is a distance considered "small" in the
	// context of the solid.
	// It is used to walk along rays to find
	// intersections.
	Epsilon float64

	// BisectCount, if non-zero, specifies the number of
	// bisections to use to narrow down collisions.
	// If it is zero, a reasonable default is used.
	BisectCount int

	// NormalSamples, if non-zero, specifies how many
	// samples to use to approximate normals.
	// If not specified, a default is used.
	NormalSamples int

	// NormalBisectEpsilon, if non-zero, specifies a small
	// distance to use in a bisection-based method to
	// compute approximate normals.
	//
	// If set, this should typically be smaller than
	// Epsilon, since smaller values don't affect runtime
	// but do improve accuracy (up to a point).
	//
	// If this is 0, bisection is not used to approximate
	// normals, but rather a more noisy but less brittle
	// algorithm.
	NormalBisectEpsilon float64
}

// Min gets the minimum boundary of the Solid.
func (s *SolidCollider) Min() Coord3D {
	return s.Solid.Min()
}

// Max gets the maximum boundary of the Solid.
func (s *SolidCollider) Max() Coord3D {
	return s.Solid.Max()
}

// RayCollisions counts the approximate number of times
// the ray collides with the solid's border.
//
// The result may be inaccurate for parts of the solid
// smaller than epsilon.
func (s *SolidCollider) RayCollisions(r *Ray, f func(RayCollision)) int {
	if s.Epsilon <= 0 {
		panic("invalid epsilon")
	}
	minFrac, maxFrac := rayCollisionWithBounds(r, s.Min(), s.Max())
	if maxFrac < minFrac || maxFrac < 0 {
		return 0
	}
	fracStep := s.Epsilon / r.Direction.Norm()
	intersections := 0
	contained := s.Solid.Contains(r.Origin)
	for t := minFrac; t <= maxFrac+fracStep; t += fracStep {
		c := r.Origin.Add(r.Direction.Scale(t))
		newContained := s.Solid.Contains(c)
		if newContained != contained {
			intersections++
			if f != nil {
				f(s.collision(r, t-fracStep, t, contained))
			}
		}
		contained = newContained
	}
	return intersections
}

// FirstRayCollision approximately finds the first time
// the ray collides with the solid.
//
// The result may be inaccurate for parts of the solid
// smaller than epsilon.
func (s *SolidCollider) FirstRayCollision(r *Ray) (RayCollision, bool) {
	if s.Epsilon <= 0 {
		panic("invalid epsilon")
	}
	minFrac, maxFrac := rayCollisionWithBounds(r, s.Min(), s.Max())
	if maxFrac < minFrac || maxFrac < 0 {
		return RayCollision{}, false
	}
	fracStep := s.Epsilon / r.Direction.Norm()
	startInside := s.Solid.Contains(r.Origin)
	for t := minFrac; t <= maxFrac+fracStep; t += fracStep {
		c := r.Origin.Add(r.Direction.Scale(t))
		if s.Solid.Contains(c) != startInside {
			return s.collision(r, t-fracStep, t, startInside), true
		}
	}
	return RayCollision{}, false
}

func (s *SolidCollider) collision(r *Ray, min, max float64, startInside bool) RayCollision {
	if startInside {
		min, max = max, min
	}
	scale := s.bisectCollision(r, min, max)
	normal := s.approximateNormal(r.Origin.Add(r.Direction.Scale(scale)))
	return RayCollision{Scale: scale, Normal: normal}
}

func (s *SolidCollider) bisectCollision(r *Ray, min, max float64) float64 {
	count := s.BisectCount
	if count == 0 {
		count = 32
	}
	for i := 0; i < count; i++ {
		f := (min + max) / 2
		if s.Solid.Contains(r.Origin.Add(r.Direction.Scale(f))) {
			max = f
		} else {
			min = f
		}
	}
	return (min + max) / 2
}

func (s *SolidCollider) approximateNormal(c Coord3D) Coord3D {
	count := s.NormalSamples
	if count == 0 {
		count = 40
	}
	if s.NormalBisectEpsilon == 0 || count < 5 {
		return s.approximateNormalAverage(c, count)
	} else {
		return s.approximateNormalBisection(c, count)
	}
}

func (s *SolidCollider) approximateNormalAverage(c Coord3D, count int) Coord3D {
	normalSum := Coord3D{}
	for i := 0; i < count; i++ {
		delta := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(),
			Z: rand.NormFloat64()}.Normalize()
		c1 := c.Add(delta.Scale(s.Epsilon))
		if s.Solid.Contains(c1) {
			normalSum = normalSum.Sub(delta)
		} else {
			normalSum = normalSum.Add(delta)
		}
	}
	return normalSum.Normalize()
}

func (s *SolidCollider) approximateNormalBisection(c Coord3D, count int) Coord3D {
	eps := s.NormalBisectEpsilon
	var planeAxes [2]Coord3D
	for i := 0; i < 2; i++ {
		v1 := NewCoord3DRandUnit().Scale(eps)
		v2 := NewCoord3DRandUnit().Scale(eps)
		if !s.Solid.Contains(c.Add(v1)) {
			v1 = v1.Scale(-1)
		}
		if s.Solid.Contains(c.Add(v2)) {
			v2 = v2.Scale(-1)
		}
		for j := 2; j < (count-1)/2; j++ {
			mp := v1.Add(v2).Normalize().Scale(eps)
			if s.Solid.Contains(c.Add(mp)) {
				v1 = mp
			} else {
				v2 = mp
			}
		}
		planeAxes[i] = v1.Add(v2).Normalize()
	}
	res := planeAxes[0].Cross(planeAxes[1]).Normalize()
	if s.Solid.Contains(c.Add(res.Scale(eps))) {
		return res.Scale(-1)
	} else {
		return res
	}
}

// SphereCollision checks if the solid touches a
// sphere with origin c and radius r.
//
// The result may be inaccurate for parts of the solid
// smaller than epsilon.
//
// This grows slower with r as O(r^3).
func (s *SolidCollider) SphereCollision(c Coord3D, r float64) bool {
	if s.Epsilon <= 0 {
		panic("invalid epsilon")
	}
	if !sphereTouchesBounds(c, r, s.Min(), s.Max()) {
		return false
	}
	for z := c.Z - r; z <= c.Z+r; z += s.Epsilon {
		for y := c.Y - r; y <= c.Y+r; y += s.Epsilon {
			for x := c.X - r; x <= c.X+r; x += s.Epsilon {
				coord := Coord3D{X: x, Y: y, Z: z}
				if c.Dist(coord) > r {
					continue
				}
				if s.Solid.Contains(coord) {
					return true
				}
			}
		}
	}
	return false
}

func sphereTouchesBounds(center Coord3D, r float64, min, max Coord3D) bool {
	return pointToBoundsDistSquared(center, min, max) <= r*r
}

func pointToBoundsDistSquared(center Coord3D, min, max Coord3D) float64 {
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 3; axis++ {
		min := min.Array()[axis]
		max := max.Array()[axis]
		value := center.Array()[axis]
		if value < min {
			distSquared += (min - value) * (min - value)
		} else if value > max {
			distSquared += (max - value) * (max - value)
		}
	}
	return distSquared
}

func rayCollisionWithBounds(r *Ray, min, max Coord3D) (minFrac, maxFrac float64) {
	minFrac = 0
	maxFrac = math.Inf(1)
	for axis := 0; axis < 3; axis++ {
		origin := r.Origin.Array()[axis]
		rate := r.Direction.Array()[axis]
		if rate == 0 {
			if origin < min.Array()[axis] || origin > max.Array()[axis] {
				return 0, -1
			}
			continue
		}
		t1 := (min.Array()[axis] - origin) / rate
		t2 := (max.Array()[axis] - origin) / rate
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t2 < 0 {
			// Short-circuit optimization.
			return 0, -1
		}
		if t1 > minFrac {
			minFrac = t1
		}
		if t2 < maxFrac {
			maxFrac = t2
		}
	}
	return
}
