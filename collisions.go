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
	matrix := Matrix3{
		t[1].X - t[0].X, t[2].X - t[0].X, r.Direction.X,
		t[1].Y - t[0].Y, t[2].Y - t[0].Y, r.Direction.Y,
		t[1].Z - t[0].Z, t[2].Z - t[0].Z, r.Direction.Z,
	}
	if math.Abs(matrix.Det()) < 1e-8 {
		return false, 0
	}
	result := matrix.Inverse().MulColumn(r.Origin.Add(t[0].Scale(-1)))
	return result.X >= 0 && result.Y >= 0 && result.X+result.Y <= 1, -result.Z
}

// A Collider is a surface which can count the number of
// times it intersects a ray, and check if any part of it
// is inside of a sphere.
type Collider interface {
	// RayCollisions counts the number of collisions with
	// a ray.
	RayCollisions(r *Ray) int

	// SphereCollision checks if the collider touches a
	// sphere with origin c and radius r.
	SphereCollision(c Coord3D, r float64) bool
}

// MeshToCollider creates an efficient Collider out of a
// mesh.
func MeshToCollider(m *Mesh) Collider {
	return colliderForTriangles(m.TriangleSlice(), 0)
}

func colliderForTriangles(tris []*Triangle, axis int) Collider {
	if len(tris) == 1 {
		return tris[0]
	}
	essentials.VoodooSort(tris, func(i, j int) bool {
		min1 := tris[i][0].array()[axis]
		min2 := tris[j][0].array()[axis]
		for k := 1; k < 3; k++ {
			min1 = math.Min(min1, tris[i][k].array()[axis])
			min2 = math.Min(min2, tris[j][k].array()[axis])
		}
		return min1 < min2
	})
	c1 := colliderForTriangles(tris[:len(tris)/2], (axis+1)%3)
	c2 := colliderForTriangles(tris[len(tris)/2:], (axis+1)%3)
	var min, max Coord3D
	if b, ok := c1.(*BoundedCollider); ok {
		min = b.Min
		max = b.Max
	} else {
		min = c1.(*Triangle).Min()
		max = c1.(*Triangle).Max()
	}
	if b, ok := c2.(*BoundedCollider); ok {
		min = b.Min.Min(min)
		max = b.Max.Max(max)
	} else {
		min = c2.(*Triangle).Min().Min(min)
		max = c2.(*Triangle).Max().Max(max)
	}
	return &BoundedCollider{
		Min:       min,
		Max:       max,
		Colliders: []Collider{c1, c2},
	}
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

// A BoundedCollider wraps multiple other Colliders and
// only passes along rays and spheres that enter a cube.
type BoundedCollider struct {
	Min       Coord3D
	Max       Coord3D
	Colliders []Collider
}

func (b *BoundedCollider) RayCollisions(r *Ray) int {
	minFrac := math.Inf(-1)
	maxFrac := math.Inf(1)
	for axis := 0; axis < 3; axis++ {
		origin := r.Origin.array()[axis]
		rate := r.Direction.array()[axis]
		if rate == 0 {
			if origin < b.Min.array()[axis] || origin > b.Max.array()[axis] {
				return 0
			}
			continue
		}
		t1 := (b.Min.array()[axis] - origin) / rate
		t2 := (b.Max.array()[axis] - origin) / rate
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		minFrac = math.Max(minFrac, t1)
		maxFrac = math.Min(maxFrac, t2)
	}

	if minFrac > maxFrac || maxFrac < 0 {
		return 0
	}

	var count int
	for _, c := range b.Colliders {
		count += c.RayCollisions(r)
	}
	return count
}

func (b *BoundedCollider) SphereCollision(center Coord3D, r float64) bool {
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 3; axis++ {
		min := b.Min.array()[axis]
		max := b.Max.array()[axis]
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

	for _, c := range b.Colliders {
		if c.SphereCollision(center, r) {
			return true
		}
	}
	return false
}
