package model2d

import (
	"math"
	"sort"
)

// A Ray is a line originating at a point and extending
// infinitely in some direction.
type Ray struct {
	Origin    Coord
	Direction Coord
}

// Collision computes where (and if) the ray intersects
// the segment.
//
// If it returns true as the first value, then the ray or
// its reverse hits the segment.
//
// The second return value is how much of the direction
// must be added to the origin to hit the line containing
// the segment.
// If it is negative, it means the segment is behind the
// ray.
func (r *Ray) Collision(s *Segment) (bool, float64) {
	v := s[1].Sub(s[0])
	matrix := Matrix2{
		v.X, r.Direction.X,
		v.Y, r.Direction.Y,
	}
	if math.Abs(matrix.Det()) < 1e-8*s.Length()*r.Direction.Norm() {
		return false, 0
	}
	matrix.InvertInPlace()
	result := matrix.MulColumn(r.Origin.Sub(s[0]))
	return result.X >= 0 && result.X <= 1, -result.Y
}

// A Collider is the outline of a 2-dimensional shape.
// It can count its intersections with a ray, and check if
// any part of the outline is inside a circle.
type Collider interface {
	Bounder

	// RayCollisions counts the number of collisions with
	// a ray.
	RayCollisions(r *Ray) int

	// FirstRayCollision gets the ray collision with the
	// lowest non-negative distance.
	// It also yields the normal from the outline where
	// the collision took place.
	FirstRayCollision(r *Ray) (collides bool, distance float64, normal Coord)

	// CircleCollision checks if the collider touches a
	// circle with origin c and radius r.
	CircleCollision(c Coord, r float64) bool
}

// ColliderContains checks if a point is within a Collider
// and at least margin away from the border.
func ColliderContains(c Collider, coord Coord, margin float64) bool {
	r := &Ray{
		Origin: coord,
		// Random direction; any direction should work, but we
		// want to avoid edge cases and rounding errors.
		Direction: Coord{0.5224892708603626, 0.10494477243214506},
	}
	collisions := c.RayCollisions(r)
	if collisions%2 == 0 {
		return false
	}
	return margin == 0 || !c.CircleCollision(coord, margin)
}

// RayCollisions returns 1 if the ray collides with the
// segment, or 0 otherwise.
func (s *Segment) RayCollisions(r *Ray) int {
	if collides, pos := r.Collision(s); collides && pos >= 0 {
		return 1
	}
	return 0
}

// FirstRayCollision returns the collision info and the
// segment normal.
func (s *Segment) FirstRayCollision(r *Ray) (bool, float64, Coord) {
	collides, frac := r.Collision(s)
	return collides && frac >= 0, frac, s.Normal()
}

// CircleCollision checks if the circle intersects the
// segment s.
func (s *Segment) CircleCollision(c Coord, r float64) bool {
	if s[0].Dist(c) < r || s[1].Dist(c) < r {
		return true
	}

	// The segment may pass through the circle without
	// either endpoint being contained.
	v := s[1].Sub(s[0])
	frac := (c.Dot(v) - s[0].Dot(v)) / v.Dot(v)
	closest := s[0].Add(v.Scale(frac))
	return frac >= 0 && frac <= 1 && closest.Dist(c) < r
}

// MeshToCollider converts a mesh to an efficient
// Collider.
func MeshToCollider(m *Mesh) Collider {
	segs := m.SegmentsSlice()
	GroupSegments(segs)
	return GroupedSegmentsToCollider(segs)
}

// GroupSegments sorts the segments recursively by their x
// and y values.
// This can be used to prepare segments for
// GroupedSegmentsToCollider.
func GroupSegments(segs []*Segment) {
	groupSegmentsAxis(segs, 0)
}

func groupSegmentsAxis(segs []*Segment, axis int) {
	if len(segs) <= 1 {
		return
	}
	sort.Slice(segs, func(i, j int) bool {
		a1 := segs[i][0].Array()
		a2 := segs[j][0].Array()
		return a1[axis] < a2[axis]
	})
	mid := len(segs) / 2
	groupSegmentsAxis(segs[:mid], (axis+1)%2)
	groupSegmentsAxis(segs[mid:], (axis+1)%2)
}

// GroupedSegmentsToCollider converts pre-grouped segments
// into an efficient collider.
// If the segments were not grouped with GroupSegments,
// then the resulting collider may be highly inefficient.
func GroupedSegmentsToCollider(segs []*Segment) Collider {
	if len(segs) == 0 {
		return NewJoinedCollider(nil)
	} else if len(segs) == 1 {
		return segs[0]
	} else {
		mid := len(segs) / 2
		c1 := GroupedSegmentsToCollider(segs[:mid])
		c2 := GroupedSegmentsToCollider(segs[mid:])
		return NewJoinedCollider([]Collider{c1, c2})
	}
}

////////////////////////////////////////////////////////////
// NOTE: almost all JoinedCollider code was able to be    //
// copied from model3d. This code duplication cannot be   //
// helped, although perhaps `go generate` should be used. //
////////////////////////////////////////////////////////////

// A JoinedCollider wraps multiple other Colliders and
// only passes along rays and circles that enter their
// combined bounding box.
type JoinedCollider struct {
	min       Coord
	max       Coord
	colliders []Collider
}

// NewJoinedCollider creates a JoinedCollider which
// combines zero or more other colliders.
func NewJoinedCollider(other []Collider) *JoinedCollider {
	if len(other) == 0 {
		return &JoinedCollider{}
	}
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

func (j *JoinedCollider) Min() Coord {
	return j.min
}

func (j *JoinedCollider) Max() Coord {
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

func (j *JoinedCollider) FirstRayCollision(r *Ray) (bool, float64, Coord) {
	if !j.rayCollidesWithBounds(r) {
		return false, 0, Coord{}
	}
	var anyCollides bool
	var closestDistance float64
	var closestNormal Coord
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

func (j *JoinedCollider) CircleCollision(center Coord, r float64) bool {
	if len(j.colliders) == 0 {
		return false
	}
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 2; axis++ {
		min := j.min.Array()[axis]
		max := j.max.Array()[axis]
		value := center.Array()[axis]
		d := 0.0
		if value < min {
			d = min - value
		} else if value > max {
			d = max - value
		}
		distSquared += d * d
	}
	if distSquared > r*r {
		return false
	}
	for _, c := range j.colliders {
		if c.CircleCollision(center, r) {
			return true
		}
	}
	return false
}

func (j *JoinedCollider) rayCollidesWithBounds(r *Ray) bool {
	if len(j.colliders) == 0 {
		return false
	}
	minFrac := math.Inf(-1)
	maxFrac := math.Inf(1)
	for axis := 0; axis < 2; axis++ {
		origin := r.Origin.Array()[axis]
		rate := r.Direction.Array()[axis]
		if rate == 0 {
			if origin < j.min.Array()[axis] || origin > j.max.Array()[axis] {
				return false
			}
			continue
		}
		invRate := 1 / rate
		t1 := (j.min.Array()[axis] - origin) * invRate
		t2 := (j.max.Array()[axis] - origin) * invRate
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t2 < 0 {
			// No collision is possible, so we can short-circuit
			// everything else.
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
