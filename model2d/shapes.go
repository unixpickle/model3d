// Generated from templates/shapes.template

package model2d

import (
	"math"
	"sort"
)

// A Circle is a 2D perfect circle.
type Circle struct {
	Center Coord
	Radius float64
}

// Min gets the minimum point of the bounding box.
func (c *Circle) Min() Coord {
	return c.Center.AddScalar(-c.Radius)
}

// Max gets the maximum point of the bounding box.
func (c *Circle) Max() Coord {
	return c.Center.AddScalar(c.Radius)
}

// Contains checks if a point c is inside the circle.
func (c *Circle) Contains(coord Coord) bool {
	return coord.Dist(c.Center) <= c.Radius
}

// FirstRayCollision gets the first ray collision with the
// circle, if one occurs.
func (c *Circle) FirstRayCollision(r *Ray) (RayCollision, bool) {
	var res RayCollision
	var ok bool
	c.RayCollisions(r, func(rc RayCollision) {
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
func (c *Circle) RayCollisions(r *Ray, f func(RayCollision)) int {
	// Want to find where ||(o+a*d)-c||^2 = r^2
	// Let's call o = (o-c) for the rest of this.
	// ||a*d+o||^2 = r^2
	// a^2*d^2 + 2*a*d*o + o^2 = r^2
	// a^2*(d^2) + a*(2*d*o) + (o^2 - r^2) = 0
	// quadratic equation: a=(d^2), b=(2*d*o), c = (o^2 - r^2)
	o := r.Origin.Sub(c.Center)
	d := r.Direction
	a := d.Dot(d)
	b := 2 * d.Dot(o)
	c_ := o.Dot(o) - c.Radius*c.Radius

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
			normal := point.Sub(c.Center).Normalize()
			f(RayCollision{Normal: normal, Scale: t, Extra: c})
		}
	}

	return count
}

// CircleCollision checks if the surface of c collides
// with a solid circle centered at c with radius r.
func (c *Circle) CircleCollision(center Coord, r float64) bool {
	return math.Abs(c.SDF(center)) <= r
}

// SDF gets the signed distance relative to the circle.
func (c *Circle) SDF(coord Coord) float64 {
	return c.Radius - coord.Dist(c.Center)
}

// PointSDF gets the signed distance function at coord
// and also returns the nearest point to it on the circle.
func (c *Circle) PointSDF(coord Coord) (Coord, float64) {
	direction := coord.Sub(c.Center)
	if norm := direction.Norm(); norm == 0 {
		// Pick an arbitrary point
		return c.Center.Add(X(c.Radius)), c.Radius
	} else {
		return c.Center.Add(direction.Scale(c.Radius / norm)), c.SDF(coord)
	}
}

// NormalSDF gets the signed distance function at coord
// and also returns the normal at the nearest point to it
// on the circle.
func (c *Circle) NormalSDF(coord Coord) (Coord, float64) {
	direction := coord.Sub(c.Center)
	if norm := direction.Norm(); norm == 0 {
		// Pick an arbitrary normal
		return X(1), c.Radius
	} else {
		return direction.Scale(1 / norm), c.SDF(coord)
	}
}

// MetaballField returns positive values outside of the
// surface, and these values increase linearly with
// distance to the surface.
func (c *Circle) MetaballField(coord Coord) float64 {
	return -c.SDF(coord)
}

// MetaballDistBound returns d always, since the metaball
// implemented by MetaballField() is defined in terms of
// standard Euclidean coordinates.
func (c *Circle) MetaballDistBound(d float64) float64 {
	return d
}

// A Rect is a 2D primitive that fills an axis-aligned
// rectangular space.
type Rect struct {
	MinVal Coord
	MaxVal Coord
}

// NewRect creates a Rect with a min and a max value.
func NewRect(min, max Coord) *Rect {
	return &Rect{MinVal: min, MaxVal: max}
}

// BoundsRect creates a Rect from a Bounder's bounds.
func BoundsRect(b Bounder) *Rect {
	return NewRect(b.Min(), b.Max())
}

// Min yields r.MinVal.
func (r *Rect) Min() Coord {
	return r.MinVal
}

// Max yields r.MaxVal.
func (r *Rect) Max() Coord {
	return r.MaxVal
}

// Contains checks if c is inside of r.
func (r *Rect) Contains(c Coord) bool {
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
		Extra:  r,
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
				Extra:  r,
			})
		}
	}
	return count
}

func (r *Rect) normalAt(c Coord) Coord {
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

	var resArr [2]float64
	resArr[axis] = sign
	return NewCoordArray(resArr)
}

// CircleCollision checks if a solid circle touches any
// part of the rectangular surface.
func (r *Rect) CircleCollision(c Coord, radius float64) bool {
	return math.Abs(r.SDF(c)) <= radius
}

// SDF gets the signed distance to the surface of the
// rectangular volume.
func (r *Rect) SDF(c Coord) float64 {
	return r.genericSDF(c, nil, nil)
}

// PointSDF gets the nearest point on the surface of the
// rect and the corresponding SDF.
func (r *Rect) PointSDF(c Coord) (Coord, float64) {
	var p Coord
	res := r.genericSDF(c, nil, &p)
	return p, res
}

// NormalSDF gets the signed distance to the rect and the
// normal at the closest point on the surface.
func (r *Rect) NormalSDF(c Coord) (Coord, float64) {
	var n Coord
	res := r.genericSDF(c, &n, nil)
	return n, res
}

func (r *Rect) genericSDF(c Coord, normalOut, pointOut *Coord) float64 {
	if !r.Contains(c) {
		// We can project directly to the rect to hit the surface.
		nearest := c.Min(r.MaxVal).Max(r.MinVal)
		if pointOut != nil {
			*pointOut = nearest
		}
		if normalOut != nil {
			*normalOut = r.normalAt(nearest)
		}
		return -c.Dist(nearest)
	}

	// Find the closest side of the rect.
	minArr := r.MinVal.Array()
	maxArr := r.MaxVal.Array()
	cArr := c.Array()
	dist := math.Inf(1)
	for i := 0; i < 2; i++ {
		minD := cArr[i] - minArr[i]
		maxD := maxArr[i] - cArr[i]
		axisD := math.Min(minD, maxD)
		if axisD < dist {
			dist = axisD
			if normalOut != nil {
				var arr [2]float64
				if minD < maxD {
					arr[i] = -1.0
				} else {
					arr[i] = 1.0
				}
				*normalOut = NewCoordArray(arr)
			}
			if pointOut != nil {
				arr := cArr
				if minD < maxD {
					arr[i] = minArr[i]
				} else {
					arr[i] = maxArr[i]
				}
				*pointOut = NewCoordArray(arr)
			}
		}
	}
	return dist
}

// MetaballField returns positive values outside of the
// surface, and these values increase linearly with
// distance to the surface.
func (r *Rect) MetaballField(coord Coord) float64 {
	return -r.SDF(coord)
}

// MetaballDistBound returns d always, since the metaball
// implemented by MetaballField() is defined in terms of
// standard Euclidean coordinates.
func (r *Rect) MetaballDistBound(d float64) float64 {
	return d
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

// A Capsule is a shape which contains all of the points
// within a given distance of a line segment.
type Capsule struct {
	P1     Coord
	P2     Coord
	Radius float64
}

// Min gets the minimum point of the bounding box.
func (c *Capsule) Min() Coord {
	return c.P1.Min(c.P2).AddScalar(-c.Radius)
}

// Max gets the maximum point of the bounding box.
func (c *Capsule) Max() Coord {
	return c.P1.Max(c.P2).AddScalar(c.Radius)
}

// Contains checks if c is within the capsule.
func (c *Capsule) Contains(coord Coord) bool {
	segment := Segment{c.P1, c.P2}
	return segment.Dist(coord) <= c.Radius
}

// FirstRayCollision gets the first ray collision with the
// capsule, if one occurs.
func (c *Capsule) FirstRayCollision(r *Ray) (RayCollision, bool) {
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
func (c *Capsule) RayCollisions(r *Ray, f func(RayCollision)) int {
	var colls []RayCollision
	for _, p := range []Coord{c.P1, c.P2} {
		circle := &Circle{Center: p, Radius: c.Radius}
		circle.RayCollisions(r, func(rc RayCollision) {
			colls = append(colls, rc)
		})
	}

	// Borrowed from Cylinder.RayCollisions.
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
				colls = append(colls, RayCollision{
					Scale:  t,
					Normal: p.Sub(v.Scale(frac)).Normalize(),
					Extra:  c,
				})
			}
		}
	}

	if len(colls) == 0 {
		return 0
	} else if len(colls) == 1 {
		if f != nil {
			f(colls[0])
		}
		return 1
	}

	sort.Slice(colls, func(i, j int) bool {
		return colls[i].Scale < colls[j].Scale
	})

	// There can be at most two collisions: entering and leaving.
	// There may be more (phantom) collisions in between since we
	// tracked collisions with spheres at the endpoints, not
	// hemispheres.
	count := 1
	if !c.Contains(r.Origin) {
		if f != nil {
			f(colls[0])
		}
		count += 1
	}
	if f != nil {
		f(colls[len(colls)-1])
	}
	return count
}

// CircleCollision checks if the surface of c collides
// with a solid circle centered at c with radius r.
func (c *Capsule) CircleCollision(center Coord, r float64) bool {
	return math.Abs(c.SDF(center)) <= r
}

// SDF gets the signed distance to the surface of the capsule.
func (c *Capsule) SDF(coord Coord) float64 {
	return c.genericSDF(coord, nil, nil)
}

// PointSDF gets the nearest point on the surface of the
// capsule and the corresponding SDF.
func (c *Capsule) PointSDF(coord Coord) (Coord, float64) {
	var p Coord
	res := c.genericSDF(coord, nil, &p)
	return p, res
}

// NormalSDF gets the signed distance to the capsule and
// the normal at the closest point on the surface.
func (c *Capsule) NormalSDF(coord Coord) (Coord, float64) {
	var n Coord
	res := c.genericSDF(coord, &n, nil)
	return n, res
}

func (c *Capsule) genericSDF(coord Coord, normalOut, pointOut *Coord) float64 {
	v := c.P2.Sub(c.P1)
	norm := v.Norm()
	axis := v.Scale(1 / norm)
	dot := coord.Sub(c.P1).Dot(axis)
	if dot < 0 || dot > norm {
		proxy := &Circle{Radius: c.Radius}
		if dot < 0 {
			proxy.Center = c.P1
		} else {
			proxy.Center = c.P2
		}
		if normalOut != nil {
			*normalOut, _ = proxy.NormalSDF(coord)
		}
		if pointOut != nil {
			*pointOut, _ = proxy.PointSDF(coord)
		}
		return proxy.SDF(coord)
	}

	sdf := c.Radius - Segment{c.P1, c.P2}.Dist(coord)
	if normalOut != nil || pointOut != nil {
		projPoint := c.P1.Add(axis.Scale(dot))
		delta := coord.Sub(projPoint)

		b1 := XY(-axis.Y, axis.X)
		normal := safeNormal(delta, b1, axis)
		if normalOut != nil {
			*normalOut = normal
		}
		if pointOut != nil {
			*pointOut = projPoint.Add(normal.Scale(c.Radius))
		}
	}
	return sdf
}

// MetaballField returns positive values outside of the
// surface, and these values increase linearly with
// distance to the surface.
func (c *Capsule) MetaballField(coord Coord) float64 {
	return -c.SDF(coord)
}

// MetaballDistBound returns d always, since the metaball
// implemented by MetaballField() is defined in terms of
// standard Euclidean coordinates.
func (c *Capsule) MetaballDistBound(d float64) float64 {
	return d
}

// A Triangle is a 2D primitive containing all of the
// convex combinations of three points.
type Triangle struct {
	coords [3]Coord
	min    Coord
	max    Coord
	invMat Matrix2
}

// NewTriangle creates a Triangle from three points.
// The triangle's behavior will be mostly identical (up to
// rounding error) regardless of the order of the points.
func NewTriangle(p1, p2, p3 Coord) *Triangle {
	v1 := p2.Sub(p1)
	v2 := p3.Sub(p1)
	mat := NewMatrix2Columns(v1, v2)

	// Handle degenerate triangles using a pseudoinverse.
	// This will result in non-NaN (but possibly incorrect)
	// barycentric coordinates.
	maxDet := math.Sqrt(v1.NormSquared() * v2.NormSquared())
	const eps = 1e-12
	if det := mat.Det(); math.Abs(det) > eps*maxDet {
		mat.InvertInPlaceDet(det)
	} else {
		var u, s, v Matrix2
		mat.SVD(&u, &s, &v)
		for i, x := range s {
			if x > eps {
				s[i] = 1 / x
			}
		}
		*mat = *v.Mul(&s).Mul(u.Transpose())
	}

	return &Triangle{
		coords: [3]Coord{p1, p2, p3},
		min:    p1.Min(p2).Min(p3),
		max:    p1.Max(p2).Max(p3),
		invMat: *mat,
	}
}

// Coords gets the coordinates passed to NewTriangle().
func (t *Triangle) Coords() [3]Coord {
	return t.coords
}

// Min gets the minimum of the corners.
func (t *Triangle) Min() Coord {
	return t.min
}

// Max gets the maximum of the corners.
func (t *Triangle) Max() Coord {
	return t.max
}

// Contains returns true if the triangle contains the point
// c. Behavior may be slightly inaccurate due to rounding
// errors, for example if c is on one of the edges.
func (t *Triangle) Contains(c Coord) bool {
	if !InBounds(t, c) {
		return false
	}
	solution := t.invMat.MulColumn(c.Sub(t.coords[0]))
	if solution.X < 0 || solution.Y < 0 || solution.X+solution.Y > 1 {
		return false
	}
	return true
}

// Barycentric returns the barycentric coordinates of c.
//
// The result should always sum to 1.
//
// If the triangle is degenerate, the behavior is
// undefined and the results may be unstable.
func (t *Triangle) Barycentric(c Coord) [3]float64 {
	solution := t.invMat.MulColumn(c.Sub(t.coords[0]))
	return [3]float64{
		1 - (solution.X + solution.Y),
		solution.X,
		solution.Y,
	}
}

// AtBarycentric computes the point at the barycentric
// coordinates.
func (t *Triangle) AtBarycentric(c [3]float64) Coord {
	var res Coord
	for i, v := range t.coords {
		res = res.Add(v.Scale(c[i]))
	}
	return res
}

// Area returns the area of the triangle.
func (t *Triangle) Area() float64 {
	c := t.coords
	return NewMatrix2Columns(c[1].Sub(c[0]), c[2].Sub(c[0])).Det() / 2
}

// FirstRayCollision gets the first ray collision with the
// triangle, if one occurs.
func (t *Triangle) FirstRayCollision(r *Ray) (RayCollision, bool) {
	var res RayCollision
	var ok bool
	t.RayCollisions(r, func(rc RayCollision) {
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
func (t *Triangle) RayCollisions(r *Ray, f func(RayCollision)) int {
	n := 0
	for i := 0; i < 3; i++ {
		seg := Segment{t.coords[i], t.coords[(i+1)%3]}
		rc, ok := seg.FirstRayCollision(r)
		if ok {
			n++
			if f != nil {
				rc.Extra = t
				f(rc)
			}
		}
	}
	return n
}

// CircleCollision checks if the surface of c collides
// with a solid Circle centered at c with radius r.
func (t *Triangle) CircleCollision(center Coord, r float64) bool {
	return math.Abs(t.SDF(center)) <= r
}

// SDF gets the signed distance to the surface of the capsule.
func (t *Triangle) SDF(c Coord) float64 {
	return t.genericSDF(c, nil, nil, nil)
}

// PointSDF gets the nearest point on the border of the
// triangle and the corresponding SDF.
func (t *Triangle) PointSDF(c Coord) (Coord, float64) {
	var p Coord
	res := t.genericSDF(c, nil, &p, nil)
	return p, res
}

// NormalSDF gets the signed distance to the triangle and
// the normal at the closest point on the border.
func (t *Triangle) NormalSDF(c Coord) (Coord, float64) {
	var n Coord
	res := t.genericSDF(c, &n, nil, nil)
	return n, res
}

// BarycentricSDF gets the signed distance to the triangle
// and the barycentric coordinates of the closest point on
// the border.
func (t *Triangle) BarycentricSDF(c Coord) ([3]float64, float64) {
	var bary [3]float64
	res := t.genericSDF(c, nil, nil, &bary)
	return bary, res
}

func (t *Triangle) genericSDF(coord Coord, normalOut, pointOut *Coord, baryOut *[3]float64) float64 {
	closest := math.Inf(1)
	var closestPoint Coord
	var closestDot float64
	closestVertex := -1
	closestEdge := 0
	for i := 0; i < 3; i++ {
		p1 := t.coords[i]
		p2 := t.coords[(i+1)%3]
		v := p2.Sub(p1)
		dot := v.Dot(coord.Sub(p1)) / v.NormSquared()

		var foundPoint Coord
		foundVertex := -1
		if dot <= 0 {
			foundPoint = p1
			foundVertex = i
		} else if dot >= 1 {
			foundPoint = p2
			foundVertex = (i + 1) % 3
		} else {
			foundPoint = p1.Add(v.Scale(dot))
		}
		distSq := foundPoint.SquaredDist(coord)
		if distSq < closest {
			closest = distSq
			closestPoint = foundPoint
			closestDot = dot
			closestVertex = foundVertex
			closestEdge = i
		}
	}

	if normalOut != nil {
		if closestVertex != -1 {
			// Point away from the midpoint of opposite edge.
			p0 := t.coords[closestVertex]
			p1 := t.coords[(closestVertex+1)%3]
			p2 := t.coords[(closestVertex+2)%3]
			*normalOut = p0.Sub(p1.Mid(p2)).Normalize()
		} else {
			p0 := t.coords[closestEdge]
			p1 := t.coords[(closestEdge+1)%3]
			p2 := t.coords[(closestEdge+2)%3]
			v := p1.Sub(p0)
			normal := XY(v.Y, -v.X).Normalize()
			if normal.Dot(p2.Sub(p0)) > 0 {
				normal = normal.Scale(-1)
			}
			*normalOut = normal
		}
	}
	if pointOut != nil {
		*pointOut = closestPoint
	}
	if baryOut != nil {
		var bary [3]float64
		if closestVertex != -1 {
			bary[closestVertex] = 1
		} else {
			bary[closestEdge] = 1 - closestDot
			bary[(closestEdge+1)%3] = closestDot
		}
		*baryOut = bary
	}
	dist := math.Sqrt(closest)
	if t.Contains(coord) {
		return dist
	} else {
		return -dist
	}
}

// MetaballField returns positive values outside of the
// surface, and these values increase linearly with
// distance to the surface.
func (t *Triangle) MetaballField(coord Coord) float64 {
	return -t.SDF(coord)
}

// MetaballDistBound returns d always, since the metaball
// implemented by MetaballField() is defined in terms of
// standard Euclidean coordinates.
func (t *Triangle) MetaballDistBound(d float64) float64 {
	return d
}

func safeNormal(direction, fallbackDirection, invalidDirection Coord) Coord {
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
