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
// must be added to the origin to hit the plane spanned by
// the triangle.
// If it is negative, it means the triangle is behind the
// ray.
func (r *Ray) Collision(t *Triangle) (bool, float64) {
	v1 := t[1].Add(t[0].Scale(-1))
	v2 := t[2].Add(t[0].Scale(-1))
	matrix := Matrix3{
		v1.X, v2.X, r.Direction.X,
		v1.Y, v2.Y, r.Direction.Y,
		v1.Z, v2.Z, r.Direction.Z,
	}
	if math.Abs(matrix.Det()) < 1e-8*v1.Norm()*v2.Norm()*r.Direction.Norm() {
		return false, 0
	}
	result := matrix.Inverse().MulColumn(r.Origin.Add(t[0].Scale(-1)))
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

// MeshToCollider creates an efficient Collider out of a
// mesh.
func MeshToCollider(m *Mesh) Collider {
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
// into a Collider.
//
// The triangles should be sorted by GroupTriangles.
// Otherwise, the resulting Collider may not be efficient.
func GroupedTrianglesToCollider(tris []*Triangle) Collider {
	if len(tris) == 1 {
		return tris[0]
	} else if len(tris) == 0 {
		return nullCollider{}
	}
	midIdx := len(tris) / 2
	c1 := GroupedTrianglesToCollider(tris[:midIdx])
	c2 := GroupedTrianglesToCollider(tris[midIdx:])
	return NewJoinedCollider([]Collider{c1, c2})
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
	v := p2.Add(p1.Scale(-1))
	frac := (c.Dot(v) - p1.Dot(v)) / v.Dot(v)
	closest := p1.Add(v.Scale(frac))
	return frac >= 0 && frac <= 1 && closest.Dist(c) < r
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
		min := j.min.array()[axis]
		max := j.max.array()[axis]
		value := center.array()[axis]
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
		origin := r.Origin.array()[axis]
		rate := r.Direction.array()[axis]
		if rate == 0 {
			if origin < j.min.array()[axis] || origin > j.max.array()[axis] {
				return false
			}
			continue
		}
		t1 := (j.min.array()[axis] - origin) / rate
		t2 := (j.max.array()[axis] - origin) / rate
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		minFrac = math.Max(minFrac, t1)
		maxFrac = math.Min(maxFrac, t2)
	}

	return minFrac <= maxFrac && maxFrac >= 0
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
