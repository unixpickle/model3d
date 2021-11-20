// Generated from templates/shapes.template

package model3d

import (
	"math"

	"github.com/unixpickle/model3d/numerical"
)

// A Sphere is a spherical 3D primitive.
type Sphere struct {
	Center Coord3D
	Radius float64
}

// Min gets the minimum point of the bounding box.
func (s *Sphere) Min() Coord3D {
	return s.Center.AddScalar(-s.Radius)
}

// Max gets the maximum point of the bounding box.
func (s *Sphere) Max() Coord3D {
	return s.Center.AddScalar(s.Radius)
}

// Contains checks if a point c is inside the sphere.
func (s *Sphere) Contains(coord Coord3D) bool {
	return coord.Dist(s.Center) <= s.Radius
}

// FirstRayCollision gets the first ray collision with the
// sphere, if one occurs.
func (s *Sphere) FirstRayCollision(r *Ray) (RayCollision, bool) {
	var res RayCollision
	var ok bool
	s.RayCollisions(r, func(rc RayCollision) {
		// Collisions are sorted from first to last.
		if !ok {
			res = rc
			ok = true
		}
	})
	return res, ok
}

// RayCollisions calls f (if non-nil) with every ray
// collision.
//
// It returns the total number of collisions.
func (s *Sphere) RayCollisions(r *Ray, f func(RayCollision)) int {
	// Want to find where ||(o+a*d)-c||^2 = r^2
	// Let's call o = (o-c) for the rest of this.
	// ||a*d+o||^2 = r^2
	// a^2*d^2 + 2*a*d*o + o^2 = r^2
	// a^2*(d^2) + a*(2*d*o) + (o^2 - r^2) = 0
	// quadratic equation: a=(d^2), b=(2*d*o), c = (o^2 - r^2)
	o := r.Origin.Sub(s.Center)
	d := r.Direction
	a := d.Dot(d)
	b := 2 * d.Dot(o)
	c_ := o.Dot(o) - s.Radius*s.Radius

	discriminant := b*b - 4*a*c_
	if discriminant <= 0 {
		return 0
	}

	sqrtDisc := math.Sqrt(discriminant)
	t1 := (-b + sqrtDisc) / (2 * a)
	t2 := (-b - sqrtDisc) / (2 * a)
	if t1 > t2 {
		t1, t2 = t2, t1
	}

	var count int
	for _, t := range []float64{t1, t2} {
		if t < 0 {
			continue
		}
		count++
		if f != nil {
			point := r.Origin.Add(r.Direction.Scale(t))
			normal := point.Sub(s.Center).Normalize()
			f(RayCollision{Normal: normal, Scale: t})
		}
	}

	return count
}

// SphereCollision checks if the surface of s collides
// with a solid sphere centered at c with radius r.
func (s *Sphere) SphereCollision(center Coord3D, r float64) bool {
	return math.Abs(s.SDF(center)) <= r
}

// SDF gets the signed distance relative to the sphere.
func (s *Sphere) SDF(coord Coord3D) float64 {
	return s.Radius - coord.Dist(s.Center)
}

// PointSDF gets the signed distance function at coord
// and also returns the nearest point to it on the sphere.
func (s *Sphere) PointSDF(coord Coord3D) (Coord3D, float64) {
	direction := coord.Sub(s.Center)
	if norm := direction.Norm(); norm == 0 {
		// Pick an arbitrary point
		return s.Center.Add(X(s.Radius)), s.Radius
	} else {
		return s.Center.Add(direction.Scale(s.Radius / norm)), s.SDF(coord)
	}
}

// NormalSDF gets the signed distance function at coord
// and also returns the normal at the nearest point to it
// on the sphere.
func (s *Sphere) NormalSDF(coord Coord3D) (Coord3D, float64) {
	direction := coord.Sub(s.Center)
	if norm := direction.Norm(); norm == 0 {
		// Pick an arbitrary normal
		return X(1), s.Radius
	} else {
		return direction.Scale(1 / norm), s.SDF(coord)
	}
}

// A Rect is a 3D primitive that fills an axis-aligned
// rectangular volume.
type Rect struct {
	MinVal Coord3D
	MaxVal Coord3D
}

// NewRect creates a Rect with a min and a max value.
func NewRect(min, max Coord3D) *Rect {
	return &Rect{MinVal: min, MaxVal: max}
}

// BoundsRect creates a Rect from a Bounder's bounds.
func BoundsRect(b Bounder) *Rect {
	return NewRect(b.Min(), b.Max())
}

// Min yields r.MinVal.
func (r *Rect) Min() Coord3D {
	return r.MinVal
}

// Max yields r.MaxVal.
func (r *Rect) Max() Coord3D {
	return r.MaxVal
}

// Contains checks if c is inside of r.
func (r *Rect) Contains(c Coord3D) bool {
	return c.Min(r.MinVal) == r.MinVal && c.Max(r.MaxVal) == r.MaxVal
}

// FirstRayCollision gets the first ray collision with the
// rectangular surface.
func (r *Rect) FirstRayCollision(ray *Ray) (RayCollision, bool) {
	tMin, tMax := rayCollisionWithBounds(ray, r.MinVal, r.MaxVal)
	if tMax < tMin || tMax < 0 {
		return RayCollision{}, false
	}

	t := tMin
	if t < 0 {
		t = tMax
	}

	return RayCollision{
		Scale:  t,
		Normal: r.normalAt(ray.Origin.Add(ray.Direction.Scale(t))),
	}, true
}

// RayCollisions calls f (if non-nil) with each ray
// collision with the rectangular surface.
// It returns the number of collisions.
func (r *Rect) RayCollisions(ray *Ray, f func(RayCollision)) int {
	tMin, tMax := rayCollisionWithBounds(ray, r.MinVal, r.MaxVal)
	if tMax < tMin || tMax < 0 {
		return 0
	}

	var count int
	for _, t := range []float64{tMin, tMax} {
		if t < 0 {
			continue
		}
		count++
		if f != nil {
			f(RayCollision{
				Scale:  t,
				Normal: r.normalAt(ray.Origin.Add(ray.Direction.Scale(t))),
			})
		}
	}
	return count
}

func (r *Rect) normalAt(c Coord3D) Coord3D {
	var axis int
	var sign float64
	minDist := math.Inf(1)

	minArr := r.MinVal.Array()
	maxArr := r.MaxVal.Array()
	cArr := c.Array()
	for i, cVal := range cArr {
		if d := math.Abs(cVal - minArr[i]); d < minDist {
			minDist = d
			sign = -1
			axis = i
		}
		if d := math.Abs(cVal - maxArr[i]); d < minDist {
			minDist = d
			sign = 1
			axis = i
		}
	}

	var resArr [3]float64
	resArr[axis] = sign
	return NewCoord3DArray(resArr)
}

// SphereCollision checks if a solid sphere touches any
// part of the rectangular surface.
func (r *Rect) SphereCollision(c Coord3D, radius float64) bool {
	return math.Abs(r.SDF(c)) <= radius
}

// SDF gets the signed distance to the surface of the
// rectangular volume.
func (r *Rect) SDF(c Coord3D) float64 {
	return r.genericSDF(c, nil, nil)
}

// PointSDF gets the nearest point on the surface of the
// cube and the corresponding SDF.
func (r *Rect) PointSDF(c Coord3D) (Coord3D, float64) {
	var p Coord3D
	res := r.genericSDF(c, nil, &p)
	return p, res
}

// NormalSDF gets the nearest point on the surface of the
// cube and the corresponding SDF.
func (r *Rect) NormalSDF(c Coord3D) (Coord3D, float64) {
	var n Coord3D
	res := r.genericSDF(c, &n, nil)
	return n, res
}

func (r *Rect) genericSDF(c Coord3D, normalOut, pointOut *Coord3D) float64 {
	if !r.Contains(c) {
		// We can project directly to the cube to hit the surface.
		nearest := c.Min(r.MaxVal).Max(r.MinVal)
		if pointOut != nil {
			*pointOut = nearest
		}
		if normalOut != nil {
			minArr := r.MinVal.Array()
			maxArr := r.MaxVal.Array()
			var normal [3]float64
			for i, x := range nearest.Array() {
				if x == minArr[i] {
					normal = [3]float64{}
					normal[i] = -1
				} else if x == maxArr[i] {
					normal = [3]float64{}
					normal[i] = 1
				}
			}
			*normalOut = NewCoord3DArray(normal)
		}
		return -c.Dist(nearest)
	}

	// Find the closest side of the cube.
	minArr := r.MinVal.Array()
	maxArr := r.MaxVal.Array()
	cArr := c.Array()
	dist := math.Inf(1)
	for i := 0; i < 3; i++ {
		minD := cArr[i] - minArr[i]
		maxD := maxArr[i] - cArr[i]
		axisD := math.Min(minD, maxD)
		if axisD < dist {
			dist = axisD
			if normalOut != nil {
				var arr [3]float64
				if minD < maxD {
					arr[i] = -1.0
				} else {
					arr[i] = 1.0
				}
				*normalOut = NewCoord3DArray(arr)
			}
			if pointOut != nil {
				arr := cArr
				if minD < maxD {
					arr[i] = minArr[i]
				} else {
					arr[i] = maxArr[i]
				}
				*pointOut = NewCoord3DArray(arr)
			}
		}
	}
	return dist
}

// Expand returns a new Rect that is delta units further
// along in every direction, making it a total of 2*delta
// units longer along each axis.
func (r *Rect) Expand(delta float64) *Rect {
	return &Rect{
		MinVal: r.MinVal.AddScalar(-delta),
		MaxVal: r.MaxVal.AddScalar(delta),
	}
}

// A Cylinder is a cylindrical 3D primitive.
//
// The cylinder is defined as all the positions within a
// radius distance from the line segment between P1 and
// P2.
type Cylinder struct {
	P1     Coord3D
	P2     Coord3D
	Radius float64
}

// Min gets the minimum point of the bounding box.
func (c *Cylinder) Min() Coord3D {
	minCenter := c.P1.Min(c.P2)
	axis := c.P2.Sub(c.P1)
	minOffsets := (Coord3D{
		circleAxisBound(0, axis, -1),
		circleAxisBound(1, axis, -1),
		circleAxisBound(2, axis, -1),
	}).Scale(c.Radius)
	return minCenter.Add(minOffsets)
}

// Max gets the maximum point of the bounding box.
func (c *Cylinder) Max() Coord3D {
	maxCenter := c.P1.Max(c.P2)
	axis := c.P2.Sub(c.P1)
	maxOffsets := (Coord3D{
		circleAxisBound(0, axis, 1),
		circleAxisBound(1, axis, 1),
		circleAxisBound(2, axis, 1),
	}).Scale(c.Radius)
	return maxCenter.Add(maxOffsets)
}

// circleAxisBound gets the furthest along an axis
// (x, y, or z) you can move while remaining inside a unit
// circle which is normal to a given vector.
// The sign argument indicates if we are moving in the
// negative or positive direction.
func circleAxisBound(axis int, normal Coord3D, sign float64) float64 {
	var arr [3]float64
	arr[axis] = sign
	proj := NewCoord3DArray(arr).ProjectOut(normal)

	// Care taken to deal with numerical issues.
	proj = proj.Scale(1 / (proj.Norm() + 1e-8))
	return sign * (math.Abs(proj.Array()[axis]) + 1e-8)
}

// Contains checks if a point p is within the cylinder.
func (c *Cylinder) Contains(p Coord3D) bool {
	diff := c.P1.Sub(c.P2)
	direction := diff.Normalize()
	frac := p.Sub(c.P2).Dot(direction)
	if frac < 0 || frac > diff.Norm() {
		return false
	}
	projection := c.P2.Add(direction.Scale(frac))
	return projection.Dist(p) <= c.Radius
}

// FirstRayCollision gets the first ray collision with the
// cylinder, if one occurs.
func (c *Cylinder) FirstRayCollision(r *Ray) (RayCollision, bool) {
	var res RayCollision
	var ok bool
	c.RayCollisions(r, func(rc RayCollision) {
		if !ok || rc.Scale < res.Scale {
			res = rc
			ok = true
		}
	})
	return res, ok
}

// RayCollisions calls f (if non-nil) with every ray
// collision.
//
// It returns the total number of collisions.
func (c *Cylinder) RayCollisions(r *Ray, f func(RayCollision)) int {
	// First, detect collisions with the rounded sides.
	//
	// For simplicity, set P1 = 0 and v = P2 - P1 and
	// ||v|| = 1.
	//
	//     dist = min_a  ||a*v - p||
	//
	// We can solve for distance like so:
	//
	//     0 = dist'
	//       = v * (a*v - p)
	//       = a*||v||^2 - p*v
	//     a = p*v / ||v||^2
	//       = p*v
	//     dist = ||v*(p*v) - p||
	//
	// An intersection occurs when the distance is exactly
	// equal to the radius r. Thus, with ray scale t:
	//
	//     let p = o + t*d
	//     let v1 = v*(o*v) - o
	//     let v2 = v*(d*v) - d
	//     r^2 = ||v*(p*v) - p||^2
	//         = ||v*((o+t*d)*v) - (o+t*d)||^2
	//         = ||v*(o*v+t*(d*v)) - (o+t*d)||^2
	//         = ||v*(o*v) - o + t*(v*(d*v) - d)||^2
	//         = ||v1 + t*v2||^2
	//         = ||v1||^2 + 2*t*(v1*v2) + t^2*||v2||^2
	//     quad eq: a=||v2||^2, b=2*(v1*v2), c=||v1||^2-||r||^2
	//     solutions are (-b +- sqrt(b^2 - 4*a*c)) / (2*a)
	//

	var count int

	v := c.P2.Sub(c.P1).Normalize()
	o := r.Origin.Sub(c.P1)
	d := r.Direction
	v1 := v.Scale(o.Dot(v)).Sub(o)
	v2 := v.Scale(d.Dot(v)).Sub(d)
	a := v2.Dot(v2)
	b := 2 * v1.Dot(v2)
	cVal := v1.Dot(v1) - c.Radius*c.Radius
	discriminant := b*b - 4*a*cVal

	if discriminant > 0 {
		sqrt := math.Sqrt(discriminant)
		maxScale := c.P2.Sub(c.P1).Norm()
		for _, sign := range []float64{-1, 1} {
			t := (-b + sign*sqrt) / (2 * a)
			if t < 0 {
				// Colisions cannot occur before ray start.
				continue
			}
			// Make sure the collision happens between P1 and P2.
			p := o.Add(d.Scale(t))
			if frac := v.Dot(p); frac >= 0 && frac < maxScale {
				count++
				if f != nil {
					f(RayCollision{
						Scale:  t,
						Normal: p.Sub(v.Scale(frac)).Normalize(),
					})
				}
			}
		}
	}

	// Now detect collisions at the tips.
	for i, tip := range []Coord3D{c.P1, c.P2} {
		normal := v
		if i == 0 {
			normal = normal.Scale(-1)
		}
		bias := normal.Dot(tip)
		coll, ok := castPlane(normal, bias, r)
		if !ok {
			continue
		}
		p := r.Origin.Add(r.Direction.Scale(coll.Scale))
		if p.Dist(tip) > c.Radius {
			continue
		}
		count++
		if f != nil {
			f(coll)
		}
	}

	return count
}

// SphereCollision detects if a sphere collides with the
// cylinder.
func (c *Cylinder) SphereCollision(center Coord3D, r float64) bool {
	return math.Abs(c.SDF(center)) <= r
}

// SDF gets the signed distance to the cylinder.
func (c *Cylinder) SDF(coord Coord3D) float64 {
	return c.genericSDF(coord, nil, nil)
}

// PointSDF gets the signed distance to the cylinder
// and the closest point on the surface.
func (c *Cylinder) PointSDF(coord Coord3D) (Coord3D, float64) {
	var p Coord3D
	sdf := c.genericSDF(coord, nil, &p)
	return p, sdf
}

// NormalSDF gets the signed distance to the cylinder
// and the normal at the closest point on the surface.
func (c *Cylinder) NormalSDF(coord Coord3D) (Coord3D, float64) {
	var n Coord3D
	sdf := c.genericSDF(coord, &n, nil)
	return n, sdf
}

func (c *Cylinder) genericSDF(coord Coord3D, normalOut, pointOut *Coord3D) float64 {
	axis := c.P2.Sub(c.P1)
	norm := axis.Norm()
	axis = axis.Scale(1 / norm)

	dist := math.Inf(1)
	contained := false
	if d := axis.Dot(coord.Sub(c.P1)); d >= 0 && d < norm {
		sd := c.Radius - Segment{c.P1, c.P2}.Dist(coord)
		if sd > 0 {
			contained = true
			dist = sd
		} else {
			dist = -sd
		}
		if normalOut != nil || pointOut != nil {
			projPoint := c.P1.Add(axis.Scale(d))
			delta := coord.Sub(projPoint)
			b1, _ := axis.OrthoBasis()
			normal := safeNormal(delta, b1, axis)
			if normalOut != nil {
				*normalOut = normal
			}
			if pointOut != nil {
				*pointOut = projPoint.Add(normal.Scale(c.Radius))
			}
		}
	}

	filledCircleDist(coord, c.P1, axis.Scale(-1), c.Radius, &dist, normalOut, pointOut)
	filledCircleDist(coord, c.P2, axis, c.Radius, &dist, normalOut, pointOut)
	if !contained {
		dist = -dist
	}
	return dist
}

func filledCircleDist(c, center, axis Coord3D, radius float64, curDist *float64, normalOut, pointOut *Coord3D) {
	b1, b2 := axis.OrthoBasis()
	mat := NewMatrix3Columns(b1, b2, axis).Transpose()
	proj := mat.MulColumn(c.Sub(center))
	norm2 := proj.XY().Norm()
	if norm2 < radius {
		dist := math.Abs(proj.Z)
		if dist <= *curDist {
			*curDist = dist
			if normalOut != nil {
				*normalOut = axis
			}
			if pointOut != nil {
				*pointOut = center.Add(b1.Scale(proj.X)).Add(b2.Scale(proj.Y))
			}
		}
	} else {
		norm2 -= radius
		dist := math.Sqrt(norm2*norm2 + proj.Z*proj.Z)
		if dist <= *curDist {
			*curDist = dist
			if normalOut != nil {
				*normalOut = axis
			}
			if pointOut != nil {
				dir2d := proj.XY().Normalize().Scale(radius)
				*pointOut = center.Add(b1.Scale(dir2d.X)).Add(b2.Scale(dir2d.Y))
			}
		}
	}
}

// castPlane gets the collision with r and a plane defined
// by:
//
//     normal*x = bias
//
func castPlane(normal Coord3D, bias float64, r *Ray) (RayCollision, bool) {
	// Want to solve for t such that:
	//
	//     (o+t*d)*n - b = 0
	//     o*n + t*(d*n) - b = 0
	//     t = (b - o*n) / (d*n)
	//
	dDot := r.Direction.Dot(normal)

	// Rays parallel to plane have no intersection.
	if math.Abs(dDot) < 1e-8*r.Direction.Norm()*normal.Norm() {
		return RayCollision{}, false
	}

	scale := (bias - r.Origin.Dot(normal)) / dDot
	if scale < 0 {
		return RayCollision{}, false
	}

	return RayCollision{
		Scale:  scale,
		Normal: normal,
	}, true
}

// A Cone is a 3D cone, eminating from a point towards the
// center of a base, where the base has a given radius.
type Cone struct {
	Tip    Coord3D
	Base   Coord3D
	Radius float64
}

func (c *Cone) Min() Coord3D {
	axis := c.Tip.Sub(c.Base)
	minOffsets := (Coord3D{
		circleAxisBound(0, axis, -1),
		circleAxisBound(1, axis, -1),
		circleAxisBound(2, axis, -1),
	}).Scale(c.Radius)
	return minOffsets.Add(c.Base).Min(c.Tip)
}

func (c *Cone) Max() Coord3D {
	axis := c.Tip.Sub(c.Base)
	maxOffsets := (Coord3D{
		circleAxisBound(0, axis, 1),
		circleAxisBound(1, axis, 1),
		circleAxisBound(2, axis, 1),
	}).Scale(c.Radius)
	return maxOffsets.Add(c.Base).Max(c.Tip)
}

func (c *Cone) Contains(p Coord3D) bool {
	diff := c.Tip.Sub(c.Base)
	direction := diff.Normalize()
	frac := p.Sub(c.Base).Dot(direction)
	radiusFrac := 1 - frac/diff.Norm()
	if radiusFrac < 0 || radiusFrac > 1 {
		return false
	}
	projection := c.Base.Add(direction.Scale(frac))
	return projection.Dist(p) <= c.Radius*radiusFrac
}

func (c *Cone) SDF(p Coord3D) float64 {
	baseDist := math.Inf(1)
	filledCircleDist(p, c.Base, c.Tip.Sub(c.Base).Normalize(), c.Radius, &baseDist, nil, nil)

	centerLine := c.Tip.Sub(c.Base)
	centerOffset := p.Sub(c.Base)
	proj := centerOffset.ProjectOut(centerLine)
	if proj.Norm() == 0 {
		// We are nearly at the centerline.
		proj.X = 1
	}
	axis := proj.Normalize()
	edgeSegment := NewSegment(c.Tip, c.Base.Add(axis.Scale(c.Radius)))
	edgeDist := edgeSegment.Dist(p)

	dist := math.Min(baseDist, edgeDist)
	if c.Contains(p) {
		return dist
	} else {
		return -dist
	}
}

// A Torus is a 3D primitive that represents a torus.
//
// The torus is defined by revolving a sphere of radius
// InnerRadius around the point Center and around the
// axis Axis, at a distance of OuterRadius from Center.
//
// The Torus is only valid if the inner radius is lower
// than the outer radius.
// Otherwise, invalid ray collisions and SDF values may be
// reported.
type Torus struct {
	Center      Coord3D
	Axis        Coord3D
	OuterRadius float64
	InnerRadius float64
}

// Min gets the minimum point of the bounding box.
func (t *Torus) Min() Coord3D {
	extra := XYZ(t.InnerRadius, t.InnerRadius, t.InnerRadius)
	minOffsets := (Coord3D{
		circleAxisBound(0, t.Axis, -1),
		circleAxisBound(1, t.Axis, -1),
		circleAxisBound(2, t.Axis, -1),
	}).Scale(t.OuterRadius)
	return minOffsets.Add(t.Center).Sub(extra)
}

// Max gets the maximum point of the bounding box.
func (t *Torus) Max() Coord3D {
	extra := XYZ(t.InnerRadius, t.InnerRadius, t.InnerRadius)
	minOffsets := (Coord3D{
		circleAxisBound(0, t.Axis, 1),
		circleAxisBound(1, t.Axis, 1),
		circleAxisBound(2, t.Axis, 1),
	}).Scale(t.OuterRadius)
	return minOffsets.Add(t.Center).Add(extra)
}

// Contains determines if c is within the torus.
func (t *Torus) Contains(c Coord3D) bool {
	return t.SDF(c) >= 0
}

// FirstRayCollision gets the first ray collision with the
// surface of the torus.
func (t *Torus) FirstRayCollision(ray *Ray) (RayCollision, bool) {
	var scale float64
	var result RayCollision
	var collides bool
	t.RayCollisions(ray, func(rc RayCollision) {
		if !collides || rc.Scale < scale {
			scale = rc.Scale
			result = rc
			collides = true
		}
	})
	return result, collides
}

// RayCollisions calls f (if non-nil) with each ray
// collision with the surface of the torus.
// It returns the number of collisions.
func (t *Torus) RayCollisions(ray *Ray, f func(RayCollision)) int {
	// First, transform the torus to be centered and make
	// y the central axis
	x, z := t.Axis.OrthoBasis()
	y := t.Axis.Normalize()
	basis := NewMatrix3Columns(x, y, z).Transpose()
	o := basis.MulColumn(ray.Origin.Sub(t.Center))
	d := basis.MulColumn(ray.Direction)
	r := t.InnerRadius
	R := t.OuterRadius

	// Based on http://blog.marcinchwedczuk.pl/ray-tracing-torus.
	o2SubR2 := o.NormSquared() - (r*r + R*R)
	oDotD := o.Dot(d)
	dSquared := d.NormSquared()
	polynomial := numerical.Polynomial{
		o2SubR2*o2SubR2 - 4*R*R*(r*r-o.Y*o.Y),
		4*o2SubR2*oDotD + 8*R*R*o.Y*d.Y,
		2*dSquared*o2SubR2 + 4*oDotD*oDotD + 4*R*R*d.Y*d.Y,
		4 * dSquared * oDotD,
		dSquared * dSquared,
	}
	n := 0
	polynomial.IterRealRoots(func(x float64) bool {
		if x >= 0 {
			n += 1
			p := ray.Origin.Add(ray.Direction.Scale(x))
			normal, _ := t.NormalSDF(p)
			f(RayCollision{Scale: x, Normal: normal})
		}
		return true
	})
	return n
}

// SphereCollision checks if any part of the surface of the
// torus is contained in a sphere.
func (t *Torus) SphereCollision(c Coord3D, r float64) bool {
	return math.Abs(t.SDF(c)) <= r
}

// SDF determines the minimum distance from a point to the
// surface of the torus.
func (t *Torus) SDF(c Coord3D) float64 {
	return t.genericSDF(c, nil, nil)
}

// PointSDF is like SDF, but also returns the closest point
// on the surface of the torus.
func (t *Torus) PointSDF(c Coord3D) (Coord3D, float64) {
	var point Coord3D
	dist := t.genericSDF(c, &point, nil)
	return point, dist
}

// NormalSDF is like SDF, but also returns the normal on
// the surface of the torus at the closest point to c.
func (t *Torus) NormalSDF(c Coord3D) (Coord3D, float64) {
	var normal Coord3D
	dist := t.genericSDF(c, nil, &normal)
	return normal, dist
}

func (t *Torus) genericSDF(c Coord3D, closest, normal *Coord3D) float64 {
	b1, b2 := t.Axis.OrthoBasis()
	centered := c.Sub(t.Center)

	// Compute the closest point on the ring around
	// the center of the torus.
	x := b1.Dot(centered)
	y := b2.Dot(centered)
	outerNorm := math.Sqrt(x*x + y*y)
	if outerNorm == 0 {
		// Degenerate case in the exact center of the torus.
		x = 1
		y = 0
		outerNorm = 1
	}
	scale := t.OuterRadius / outerNorm
	x *= scale
	y *= scale
	ringPoint := b1.Scale(x).Add(b2.Scale(y))

	if closest != nil || normal != nil {
		direction := centered.Sub(ringPoint)
		invalidDirection := ringPoint.Cross(t.Axis)
		direction = safeNormal(direction, t.Axis.Normalize(), invalidDirection)
		if normal != nil {
			*normal = direction
		}
		if closest != nil {
			*closest = ringPoint.Add(direction.Scale(t.InnerRadius)).Add(t.Center)
		}
	}

	return t.InnerRadius - ringPoint.Dist(centered)
}

func safeNormal(direction, fallbackDirection, invalidDirection Coord3D) Coord3D {
	if norm := direction.Norm(); norm == 0 {
		// Fully degenerate case.
		direction = fallbackDirection
	} else {
		direction = direction.Scale(1 / norm)

		// When direction was very small, it might be pointing in
		// an invalid direction once we normalize it.
		direction = direction.ProjectOut(invalidDirection)
		if norm := direction.Norm(); norm < 1e-5 {
			// Most of the direction was due to rounding error.
			direction = fallbackDirection
		} else {
			direction = direction.Scale(1 / norm)
		}
	}
	return direction
}
