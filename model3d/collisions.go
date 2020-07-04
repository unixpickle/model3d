package model3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model2d"
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

	// SphereCollision checks if the collider touches a
	// sphere with origin c and radius r.
	SphereCollision(c Coord3D, r float64) bool
}

// A TriangleCollider is like a Collider, but it can also
// check if and where a triangle intersects the surface.
//
// This is intended to be used for piecewise-linear
// surfaces where triangle-surface intersections can be
// parameterized as line segments.
//
// All methods of a TriangleCollider are safe for
// concurrency.
type TriangleCollider interface {
	Collider

	// TriangleCollisions gets all of the segments on the
	// surface which intersect the triangle t.
	TriangleCollisions(t *Triangle) []Segment
}

// A SegmentCollider is a surface which can detect if a
// line segment collides with the surface.
type SegmentCollider interface {
	// SegmentCollision returns true if the segment
	// collides with the surface.
	SegmentCollision(s Segment) bool
}

// A RectCollider is a surface which can detect if a 3D
// axis-aligned rectangular volume collides with the
// surface.
type RectCollider interface {
	// RectCollision returns true if any part of the
	// surface is inside the rect.
	RectCollision(r *Rect) bool
}

// MeshToCollider creates an efficient TriangleCollider
// out of a mesh.
func MeshToCollider(m *Mesh) TriangleCollider {
	tris := m.TriangleSlice()
	GroupTriangles(tris)
	return GroupedTrianglesToCollider(tris)
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

// BVHToCollider converts a BVH into a TriangleCollider in
// a hierarchical way.
func BVHToCollider(b *BVH) TriangleCollider {
	if b.Leaf != nil {
		return b.Leaf
	}
	other := make([]Collider, len(b.Branch))
	for i, b1 := range b.Branch {
		other[i] = BVHToCollider(b1)
	}
	return joinedTriangleCollider{NewJoinedCollider(other)}
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
		min: other[0].Min(),
		max: other[0].Max(),
	}
	for _, c := range other[1:] {
		res.min = res.min.Min(c.Min())
		res.max = res.max.Max(c.Max())
	}

	// Flatten out other joined colliders with the same
	// bounds.
	for _, c := range other {
		var jc *JoinedCollider
		switch c := c.(type) {
		case *JoinedCollider:
			jc = c
		case joinedTriangleCollider:
			jc = c.JoinedCollider
		}
		if jc != nil && jc.min == res.min && jc.max == res.max {
			res.colliders = append(res.colliders, jc.colliders...)
		} else {
			res.colliders = append(res.colliders, c)
		}
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
	minFrac = math.Max(0, minFrac)
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
	minFrac = math.Max(0, minFrac)
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
				coord := XYZ(x, y, z)
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

type profileCollider struct {
	Collider2D model2d.Collider
	Solid2D    model2d.Solid
	MinVal     Coord3D
	MaxVal     Coord3D
}

// ProfileCollider turns a 2D Collider into a 3D Collider
// by elongating the 2D Collider along the Z axis.
func ProfileCollider(coll2d model2d.Collider, minZ, maxZ float64) Collider {
	min, max := coll2d.Min(), coll2d.Max()
	return &profileCollider{
		Collider2D: coll2d,
		Solid2D:    model2d.NewColliderSolid(coll2d),
		MinVal:     XYZ(min.X, min.Y, minZ),
		MaxVal:     XYZ(max.X, max.Y, maxZ),
	}
}

func (p *profileCollider) Min() Coord3D {
	return p.MinVal
}

func (p *profileCollider) Max() Coord3D {
	return p.MaxVal
}

func (p *profileCollider) RayCollisions(r *Ray, f func(RayCollision)) int {
	collisions2d := make([]model2d.RayCollision, 0, 2)
	var inside2d func(t float64) bool

	if r.Direction.X == 0 && r.Direction.Y == 0 {
		// Special case for ray coming in directly towards the faces.
		if !p.Solid2D.Contains(r.Origin.XY()) {
			return 0
		}
		inside2d = func(t float64) bool {
			return true
		}
	} else {
		ray2d := &model2d.Ray{
			Origin:    r.Origin.XY(),
			Direction: r.Direction.XY(),
		}
		p.Collider2D.RayCollisions(ray2d, func(rc model2d.RayCollision) {
			collisions2d = append(collisions2d, rc)
		})

		inside2d = func(t float64) bool {
			var numColl int
			for _, rc := range collisions2d {
				if rc.Scale == t {
					// No matter what, we will report the 2D
					// version of this collision.
					return false
				}
				if rc.Scale > t {
					numColl++
				}
			}
			return numColl%2 == 1
		}
	}

	if r.Direction.Z == 0 {
		// Special cases for flat ray collisions, since these
		// will never collide with the faces.
		if r.Origin.Z < p.MinVal.Z || r.Origin.Z > p.MaxVal.Z {
			return 0
		}
		for _, rc := range collisions2d {
			f(RayCollision{
				Normal: XY(rc.Normal.X, rc.Normal.Y),
				Scale:  rc.Scale,
			})
		}
		return len(collisions2d)
	}

	minT := (p.MinVal.Z - r.Origin.Z) / r.Direction.Z
	maxT := (p.MaxVal.Z - r.Origin.Z) / r.Direction.Z
	minTNormal := Z(-1)
	maxTNormal := Z(1)
	if maxT < minT {
		minT, maxT = maxT, minT
		minTNormal, maxTNormal = maxTNormal, minTNormal
	}

	var count int
	if minT >= 0 && inside2d(minT) {
		f(RayCollision{
			Normal: minTNormal,
			Scale:  minT,
		})
		count++
	}
	for _, rc := range collisions2d {
		if rc.Scale >= minT && rc.Scale <= maxT {
			f(RayCollision{
				Normal: XY(rc.Normal.X, rc.Normal.Y),
				Scale:  rc.Scale,
			})
			count++
		}
	}
	if maxT >= 0 && inside2d(maxT) {
		f(RayCollision{
			Normal: maxTNormal,
			Scale:  maxT,
		})
		count++
	}

	return count
}

func (p *profileCollider) FirstRayCollision(r *Ray) (RayCollision, bool) {
	var firstRC RayCollision
	var collides bool
	p.RayCollisions(r, func(rc RayCollision) {
		if !collides || rc.Scale < firstRC.Scale {
			collides = true
			firstRC = rc
		}
	})
	return firstRC, collides
}

func (p *profileCollider) SphereCollision(c Coord3D, r float64) bool {
	faceDistance := 0.0
	if c.Z < p.MinVal.Z {
		faceDistance = p.MinVal.Z - c.Z
	} else if c.Z > p.MaxVal.Z {
		faceDistance = c.Z - p.MaxVal.Z
	}
	if faceDistance >= r {
		return false
	}

	largestR := math.Sqrt(r*r - faceDistance*faceDistance)
	if p.Collider2D.CircleCollision(c.XY(), largestR) {
		return true
	}

	// The sphere may collide with the face of the solid
	// even if it doesn't collide with an edge of the 2D
	// shape.
	absFaceDist := math.Min(math.Abs(c.Z-p.MinVal.Z), math.Abs(c.Z-p.MaxVal.Z))
	return absFaceDist < r && p.Solid2D.Contains(c.XY())
}
