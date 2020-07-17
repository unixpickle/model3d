package model2d

import (
	"math"
	"sort"
)

// A Ray is a line originating at a point and extending
// infinitely in some (positive) direction.
type Ray struct {
	Origin    Coord
	Direction Coord
}

// RayCollision is a point where a ray intersects a
// 2-dimensional outline.
type RayCollision struct {
	// The amount of the ray direction to add to the ray
	// origin to hit the point in question.
	//
	// The scale should be non-negative.
	Scale float64

	// The normal pointing outward from the outline at the
	// point of collision.
	Normal Coord
}

// A Collider is the outline of a 2-dimensional shape.
// It can count its intersections with a ray, and check if
// any part of the outline is inside a circle.
//
// All methods of a Collider are safe for concurrency.
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

	// CircleCollision checks if the collider touches a
	// circle with origin c and radius r.
	CircleCollision(c Coord, r float64) bool
}

// ColliderContains checks if a point is within a Collider
// and at least margin away from the border.
//
// If the margin is negative, points are also conatined if
// the point is less than -margin away from the surface.
func ColliderContains(c Collider, coord Coord, margin float64) bool {
	r := &Ray{
		Origin: coord,
		// Random direction; any direction should work, but we
		// want to avoid edge cases and rounding errors.
		Direction: Coord{0.5224892708603626, 0.10494477243214506},
	}
	collisions := c.RayCollisions(r, nil)
	if collisions%2 == 0 {
		if margin < 0 {
			return c.CircleCollision(coord, -margin)
		}
		return false
	}
	return margin <= 0 || !c.CircleCollision(coord, margin)
}

// A SegmentCollider is a 2-dimensional outline which can
// detect if a line segment collides with the outline.
type SegmentCollider interface {
	// SegmentCollision returns true if the segment
	// collides with the outline.
	SegmentCollision(s *Segment) bool
}

// A RectCollider is a 2-dimensional outline which can
// detect if a 2D axis-aligned rectangular area collides
// with the outline.
type RectCollider interface {
	// RectCollision returns true if any part of the
	// outline is inside the rect.
	RectCollision(r *Rect) bool
}

type MultiCollider interface {
	Collider
	SegmentCollider
	RectCollider
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
	var anyCollides bool
	var closest RayCollision
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

func (j *JoinedCollider) CircleCollision(center Coord, r float64) bool {
	if len(j.colliders) == 0 {
		return false
	}
	if !circleTouchesBounds(center, r, j.min, j.max) {
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

func circleTouchesBounds(center Coord, r float64, min, max Coord) bool {
	return pointToBoundsDistSquared(center, min, max) <= r*r
}

func pointToBoundsDistSquared(center Coord, min, max Coord) float64 {
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 2; axis++ {
		min := min.Array()[axis]
		max := max.Array()[axis]
		value := center.Array()[axis]
		d := 0.0
		if value < min {
			d = min - value
		} else if value > max {
			d = max - value
		}
		distSquared += d * d
	}
	return distSquared
}
