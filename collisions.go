package model3d

import (
	"math"

	"github.com/unixpickle/essentials"
)

// A Ray is a line originating at a point and extending
// infinitely in some direction.
type Ray struct {
	Origin    Coord3D
	Direction Coord3D
}

// Collision computes where (and if) the ray intersects
// the triangle.
//
// If it returns true as the first value, then the ray or
// its reverse hits the triangle.
//
// The second return value is how much of the direction
// must be added to the origin to hit the plane containing
// the triangle.
// If it is negative, it means the triangle is behind the
// ray.
func (r *Ray) Collision(t *Triangle) (bool, float64) {
	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])
	matrix := Matrix3{
		v1.X, v2.X, r.Direction.X,
		v1.Y, v2.Y, r.Direction.Y,
		v1.Z, v2.Z, r.Direction.Z,
	}
	if math.Abs(matrix.Det()) < 1e-8*t.Area()*r.Direction.Norm() {
		return false, 0
	}
	matrix.InvertInPlace()
	result := matrix.MulColumn(r.Origin.Sub(t[0]))
	return result.X >= 0 && result.Y >= 0 && result.X+result.Y <= 1, -result.Z
}

// A Collider is a surface which can count the number of
// times it intersects a ray, and check if any part of it
// is inside of a sphere.
type Collider interface {
	// Bounding box for the surface.
	Min() Coord3D
	Max() Coord3D

	// RayCollisions counts the number of collisions with
	// a ray.
	RayCollisions(r *Ray) int

	// FirstRayCollision gets the ray collision with the
	// lowest non-negative distance.
	// It also yields the normal from the surface where
	// the collision took place.
	FirstRayCollision(r *Ray) (collides bool, distance float64, normal Coord3D)

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

// RayCollisions returns 1 if the ray intersects the
// triangle, and 0 otherwise.
func (t *Triangle) RayCollisions(r *Ray) int {
	collides, frac := r.Collision(t)
	if collides && frac >= 0 {
		return 1
	} else {
		return 0
	}
}

// FirstRayCollision returns information about the
// triangle collision.
func (t *Triangle) FirstRayCollision(r *Ray) (collides bool, distance float64, normal Coord3D) {
	collides, frac := r.Collision(t)
	return collides && frac >= 0, frac, t.Normal()
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
	inside, frac := ray.Collision(t)
	return inside && math.Abs(frac) < r
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

func (j *JoinedCollider) RayCollisions(r *Ray) int {
	if !j.rayCollidesWithBounds(r) {
		return 0
	}

	var count int
	for _, c := range j.colliders {
		count += c.RayCollisions(r)
	}
	return count
}

func (j *JoinedCollider) FirstRayCollision(r *Ray) (bool, float64, Coord3D) {
	if !j.rayCollidesWithBounds(r) {
		return false, 0, Coord3D{}
	}
	var anyCollides bool
	var closestDistance float64
	var closestNormal Coord3D
	for _, c := range j.colliders {
		if collides, dist, normal := c.FirstRayCollision(r); collides {
			if dist < closestDistance || !anyCollides {
				closestDistance = dist
				closestNormal = normal
				anyCollides = true
			}
		}
	}
	return anyCollides, closestDistance, closestNormal
}

func (j *JoinedCollider) SphereCollision(center Coord3D, r float64) bool {
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 3; axis++ {
		min := j.min.Array()[axis]
		max := j.max.Array()[axis]
		value := center.Array()[axis]
		if value < min {
			distSquared += (min - value) * (min - value)
		} else if value > max {
			distSquared += (max - value) * (max - value)
		}
	}
	if distSquared > r*r {
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
	minFrac := math.Inf(-1)
	maxFrac := math.Inf(1)
	for axis := 0; axis < 3; axis++ {
		origin := r.Origin.Array()[axis]
		rate := r.Direction.Array()[axis]
		if rate == 0 {
			if origin < j.min.Array()[axis] || origin > j.max.Array()[axis] {
				return false
			}
			continue
		}
		t1 := (j.min.Array()[axis] - origin) / rate
		t2 := (j.max.Array()[axis] - origin) / rate
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t2 < 0 {
			// Short-circuit optimization.
			return false
		}
		if t1 > minFrac {
			minFrac = t1
		}
		if t2 < maxFrac {
			maxFrac = t2
		}
	}

	return minFrac <= maxFrac && maxFrac >= 0
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

func (n nullCollider) RayCollisions(r *Ray) int {
	return 0
}

func (n nullCollider) FirstRayCollision(r *Ray) (collides bool, distance float64, normal Coord3D) {
	return false, 0, Coord3D{}
}

func (n nullCollider) SphereCollision(c Coord3D, r float64) bool {
	return false
}

func (n nullCollider) TriangleCollisions(t *Triangle) []Segment {
	return nil
}
