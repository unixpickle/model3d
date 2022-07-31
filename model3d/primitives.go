package model3d

import (
	"math"
)

// A Triangle is a triangle in 3D Euclidean space.
type Triangle [3]Coord3D

// AtBarycentric computes the point at the barycentric
// coordinates.
func (t *Triangle) AtBarycentric(c [3]float64) Coord3D {
	var res Coord3D
	for i, v := range t {
		res = res.Add(v.Scale(c[i]))
	}
	return res
}

// Area computes the area of the triangle.
func (t *Triangle) Area() float64 {
	return t.crossProduct().Norm() / 2
}

// Normal computes a normal vector for the triangle using
// the right-hand rule.
func (t *Triangle) Normal() Coord3D {
	return t.crossProduct().Normalize()
}

func (t *Triangle) crossProduct() Coord3D {
	return t[1].Sub(t[0]).Cross(t[2].Sub(t[0]))
}

// Min gets the element-wise minimum of all the points.
func (t *Triangle) Min() Coord3D {
	return t[0].Min(t[1]).Min(t[2])
}

// Max gets the element-wise maximum of all the points.
func (t *Triangle) Max() Coord3D {
	return t[0].Max(t[1]).Max(t[2])
}

// Segments gets all three line segments in the triangle.
func (t *Triangle) Segments() [3]Segment {
	var res [3]Segment
	for i := 0; i < 3; i++ {
		res[i] = NewSegment(t[i], t[(i+1)%3])
	}
	return res
}

// SharesEdge checks if p shares exactly one edge with p1.
func (t *Triangle) SharesEdge(t1 *Triangle) bool {
	return t.inCommon(t1) == 2
}

func (t *Triangle) inCommon(t1 *Triangle) int {
	inCommon := 0
	// Unrolled because this is a hot path in some algorithms.
	if t[0] == t1[0] || t[0] == t1[1] || t[0] == t1[2] {
		inCommon++
	}
	if t[1] == t1[0] || t[1] == t1[1] || t[1] == t1[2] {
		inCommon++
	}
	if t[2] == t1[0] || t[2] == t1[1] || t[2] == t1[2] {
		inCommon++
	}
	return inCommon
}

func (t *Triangle) sharedSegment(t1 *Triangle) Segment {
	var shared [2]Coord3D
	var num int
	for _, c := range t {
		for _, c1 := range t1 {
			if c == c1 {
				shared[num] = c
				num++
				break
			}
		}
	}
	if num != 2 {
		panic("no shared edge detected")
	}
	return NewSegment(shared[0], shared[1])
}

func (t *Triangle) otherSegment(v Coord3D) Segment {
	for i, c := range t {
		if c == v {
			return NewSegment(t[(i+1)%3], t[(i+2)%3])
		}
	}
	panic("triangle does not contain point")
}

// AreaGradient computes the gradient of the triangle's
// area with respect to every coordinate.
func (t *Triangle) AreaGradient() *Triangle {
	var grad Triangle
	for i, p := range t {
		p1 := t[(i+1)%3]
		p2 := t[(i+2)%3]
		base := p2.Sub(p1)
		baseNorm := base.Norm()
		if baseNorm == 0 {
			continue
		}
		// Project the base out of v to get the height
		// vector of the triangle.
		normed := base.Scale(1 / baseNorm)
		v := p.Sub(p1)
		v = v.Sub(normed.Scale(normed.Dot(v)))

		vNorm := v.Norm()
		if vNorm == 0 {
			continue
		}
		grad[i] = v.Scale(baseNorm / (2 * vNorm))
	}
	return &grad
}

// Dist gets the minimum distance from c to a point on the
// triangle.
func (t *Triangle) Dist(c Coord3D) float64 {
	// Special case where closest point is inside the
	// triangle, rather than on an edge.
	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])
	mat := NewMatrix3Columns(v1, v2, t.Normal())
	mat.InvertInPlace()
	components := mat.MulColumn(c.Sub(t[0]))
	if components.X >= 0 && components.Y >= 0 && components.X+components.Y <= 1 {
		return math.Abs(components.Z)
	}

	// Check all three edges.
	result := math.Inf(1)
	for _, s := range t.Segments() {
		if d := s.Dist(c); d < result {
			result = d
		}
	}
	return result
}

// Closest gets the point on the triangle closest to c.
func (t *Triangle) Closest(c Coord3D) Coord3D {
	// This is similar to Dist(), but returns a point.

	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])
	mat := NewMatrix3Columns(v1, v2, t.Normal())
	mat.InvertInPlace()
	components := mat.MulColumn(c.Sub(t[0]))
	if components.X >= 0 && components.Y >= 0 && components.X+components.Y <= 1 {
		return t[0].Add(v1.Scale(components.X)).Add(v2.Scale(components.Y))
	}

	closestDist := math.Inf(1)
	closestPoint := Coord3D{}
	for _, s := range t.Segments() {
		c1 := s.Closest(c)
		if d := c1.Dist(c); d < closestDist {
			closestDist = d
			closestPoint = c1
		}
	}
	return closestPoint
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

// rayCollision is like FirstRayCollision, but it reports
// negative scales for use in SDFs and the like.
func (t *Triangle) rayCollision(r *Ray) (tc *TriangleCollision, scale float64) {
	if d := t.Normal().Dot(r.Direction.Normalize()); d < 1e-8 && d > -1e-8 {
		// Ray is parallel to the triangle.
		return nil, 0
	}

	// Adapted from here, but made more readable by trying
	// to expose what the variables are:
	// https://en.m.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm

	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])

	cross1 := r.Direction.Cross(v2)
	det := cross1.Dot(v1)
	if det == 0 {
		// While the ray is not parallel to the triangle,
		// somehow this determinant is 0, likely due to
		// floating-point underflow.
		return nil, 0
	}
	invDet := 1 / det

	o := r.Origin.Sub(t[0])
	bary1 := invDet * o.Dot(cross1)
	if bary1 < 0 || bary1 > 1 {
		return nil, 0
	}

	cross2 := o.Cross(v1)
	bary2 := invDet * r.Direction.Dot(cross2)
	if bary2 < 0 || bary1+bary2 > 1 {
		return nil, 0
	}

	tc = &TriangleCollision{
		Triangle:    t,
		Barycentric: [3]float64{1 - (bary1 + bary2), bary1, bary2},
	}
	scale = invDet * v2.Dot(cross2)

	return
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

// SegmentCollision checks if the segment collides with
// the triangle.
func (t *Triangle) SegmentCollision(s Segment) bool {
	coll, ok := t.FirstRayCollision(&Ray{Origin: s[0], Direction: s[1].Sub(s[0])})
	return ok && coll.Scale <= 1
}

// RectCollision checks if any part of the triangle is
// inside the rect.
func (t *Triangle) RectCollision(rect *Rect) bool {
	// Check if any of the triangle's segments intersect
	// the rect.
	for _, s := range t.Segments() {
		if s.RectCollision(rect) {
			return true
		}
	}
	// The rect could still collide with the plane of the
	// triangle, so check the segments of the rect.
	for _, x := range [2]float64{rect.MinVal.X, rect.MaxVal.X} {
		segs := [4]Segment{
			NewSegment(XYZ(x, rect.MinVal.Y, rect.MinVal.Z), XYZ(x, rect.MinVal.Y, rect.MaxVal.Z)),
			NewSegment(XYZ(x, rect.MinVal.Y, rect.MinVal.Z), XYZ(x, rect.MaxVal.Y, rect.MinVal.Z)),
			NewSegment(XYZ(x, rect.MaxVal.Y, rect.MinVal.Z), XYZ(x, rect.MaxVal.Y, rect.MaxVal.Z)),
			NewSegment(XYZ(x, rect.MinVal.Y, rect.MaxVal.Z), XYZ(x, rect.MaxVal.Y, rect.MaxVal.Z)),
		}
		for _, s := range segs {
			if t.SegmentCollision(s) {
				return true
			}
		}
	}
	segs := [4]Segment{
		NewSegment(
			XYZ(rect.MinVal.X, rect.MinVal.Y, rect.MinVal.Z),
			XYZ(rect.MaxVal.X, rect.MinVal.Y, rect.MinVal.Z),
		),
		NewSegment(
			XYZ(rect.MinVal.X, rect.MaxVal.Y, rect.MinVal.Z),
			XYZ(rect.MaxVal.X, rect.MaxVal.Y, rect.MinVal.Z),
		),
		NewSegment(
			XYZ(rect.MinVal.X, rect.MinVal.Y, rect.MaxVal.Z),
			XYZ(rect.MaxVal.X, rect.MinVal.Y, rect.MaxVal.Z),
		),
		NewSegment(
			XYZ(rect.MinVal.X, rect.MaxVal.Y, rect.MaxVal.Z),
			XYZ(rect.MaxVal.X, rect.MaxVal.Y, rect.MaxVal.Z),
		),
	}
	for _, s := range segs {
		if t.SegmentCollision(s) {
			return true
		}
	}
	return false
}

// A InterpNormalTriangle is a Triangle which returns
// smoothed normals for ray collisions.
type InterpNormalTriangle struct {
	Triangle

	VertexNormals [3]Coord3D
}

// InterpNormal returns the interpolated normal for the
// barycentric coordinates within the triangle.
func (i *InterpNormalTriangle) InterpNormal(barycentric [3]float64) Coord3D {
	var normal Coord3D
	for j, vn := range i.VertexNormals {
		normal = normal.Add(vn.Scale(barycentric[j]))
	}
	return normal.Normalize()
}

func (i *InterpNormalTriangle) FirstRayCollision(r *Ray) (RayCollision, bool) {
	info, scale := i.Triangle.rayCollision(r)
	if info != nil && scale >= 0 {
		return RayCollision{
			Scale:  scale,
			Normal: i.InterpNormal(info.Barycentric),
			Extra:  info,
		}, true
	} else {
		return RayCollision{}, false
	}
}

func (i *InterpNormalTriangle) RayCollisions(r *Ray, f func(RayCollision)) int {
	info, scale := i.Triangle.rayCollision(r)
	if info == nil || scale < 0 {
		return 0
	}
	if f != nil {
		f(RayCollision{Scale: scale, Normal: i.InterpNormal(info.Barycentric), Extra: info})
	}
	return 1
}

// A Segment is a line segment in a canonical ordering,
// such that segments can be compared via the == operator
// even if they were created with their points in the
// opposite order.
type Segment [2]Coord3D

// NewSegment creates a segment with the canonical
// ordering.
func NewSegment(p1, p2 Coord3D) Segment {
	if p1.X < p2.X || (p1.X == p2.X && p1.Y < p2.Y) ||
		(p1.X == p2.X && p1.Y == p2.Y && p1.Z < p2.Z) {
		return Segment{p1, p2}
	} else {
		return Segment{p2, p1}
	}
}

// Mid gets the midpoint of the segment.
func (s Segment) Mid() Coord3D {
	return s[0].Mid(s[1])
}

// Length gets the length of the segment.
func (s Segment) Length() float64 {
	return s[0].Dist(s[1])
}

// Dist gets the minimum distance from c to a point on the
// line segment.
func (s Segment) Dist(c Coord3D) float64 {
	return c.Dist(s.Closest(c))
}

// L1Dist gets the minimum L1 distance from c to a point
// on the line segment.
func (s *Segment) L1Dist(c Coord3D) float64 {
	return c.L1Dist(s.ClosestL1(c))
}

// Closest gets the point on the segment closest to c.
func (s Segment) Closest(c Coord3D) Coord3D {
	v1 := s[1].Sub(s[0])
	norm := v1.Norm()
	v := v1.Scale(1 / norm)

	v2 := c.Sub(s[0])
	mag := v.Dot(v2)
	if mag > norm {
		return s[1]
	} else if mag < 0 {
		return s[0]
	}

	return v.Scale(mag).Add(s[0])
}

// ClosestL1 gets the point on the segment closest to c as
// measured in L1 distance.
func (s *Segment) ClosestL1(c Coord3D) Coord3D {
	p0 := s[0]
	v := s[1].Sub(s[0])

	dist := math.Inf(1)
	closest := Coord3D{}
	checkT := func(t float64) {
		tPoint := p0.Add(v.Scale(t))
		d := tPoint.L1Dist(c)
		if d < dist {
			dist = d
			closest = tPoint
		}
	}

	checkT(0.0)
	checkT(1.0)

	vArr := v.Array()
	cArr := c.Sub(p0).Array()

	for i := 0; i < 3; i++ {
		if vArr[i] == 0 {
			continue
		}
		t := cArr[i] / vArr[i]
		if t > 0 && t < 1 {
			checkT(t)
		}
	}

	return closest
}

// RectCollision checks if the segment intersects a Rect.
func (s Segment) RectCollision(r *Rect) bool {
	minFrac, maxFrac := rayCollisionWithBounds(&Ray{
		Origin:    s[0],
		Direction: s[1].Sub(s[0]),
	}, r.MinVal, r.MaxVal)
	return maxFrac >= minFrac && maxFrac >= 0 && minFrac <= 1
}

// Other gets a point in t that is not in s, provided that
// t is not degenerate.
func (s Segment) Other(t *Triangle) Coord3D {
	if t[0] != s[0] && t[0] != s[1] {
		return t[0]
	} else if t[1] != s[0] && t[1] != s[1] {
		return t[1]
	} else {
		return t[2]
	}
}

// union finds the point that s and s1 have in common,
// assuming that they have exactly one point in common.
func (s Segment) union(s1 Segment) Coord3D {
	if s1[0] == s[0] || s1[0] == s[1] {
		return s1[0]
	} else {
		return s1[1]
	}
}

// inverseUnion finds the two points that s and s1 do not
// have in common, assuming that they have exactly one
// point in common.
//
// The first point is from s, and the second is from s1.
func (s Segment) inverseUnion(s1 Segment) (Coord3D, Coord3D) {
	union := s.union(s1)
	if union == s1[0] {
		if union == s[0] {
			return s[1], s1[1]
		} else {
			return s[0], s1[1]
		}
	} else {
		if union == s[0] {
			return s[1], s1[0]
		} else {
			return s[0], s1[0]
		}
	}
}
