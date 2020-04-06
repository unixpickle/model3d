package render3d

import (
	"math"

	"github.com/unixpickle/model3d"
)

// An Object is a renderable 3D object.
type Object interface {
	// Cast finds the first collision with ray r.
	//
	// It returns not only the ray collision, but also the
	// material on the surface of the object at the point.
	//
	// The final return value indicates if there was a
	// collision or not.
	Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool)
}

// A ColliderObject wraps a model3d.Collider in the Object
// interface, using a constant material.
type ColliderObject struct {
	Collider model3d.Collider
	Material Material
}

// Cast returns the first ray collision.
func (c *ColliderObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	coll, ok := c.Collider.FirstRayCollision(r)
	return coll, c.Material, ok
}

// A JoinedObject combines multiple Objects.
type JoinedObject []Object

// Cast casts the ray onto the objects and chooses the
// closest ray collision.
func (j JoinedObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	var coll model3d.RayCollision
	var mat Material
	var found bool
	for _, o := range j {
		if c, m, f := o.Cast(r); f && (!found || c.Scale < coll.Scale) {
			coll = c
			mat = m
			found = true
		}
	}
	return coll, mat, found
}

// A Sphere is an Object implementing a perfect, single-
// material sphere.
type Sphere struct {
	Center   model3d.Coord3D
	Radius   float64
	Material Material
}

// Cast casts the ray onto the surface of the sphere.
func (s *Sphere) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
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
	c := o.Dot(o) - s.Radius*s.Radius

	discriminant := b*b - 4*a*c
	if discriminant <= 0 {
		return model3d.RayCollision{}, nil, false
	}

	sqrtDisc := math.Sqrt(discriminant)
	t1 := (-b + sqrtDisc) / (2 * a)
	t2 := (-b - sqrtDisc) / (2 * a)
	if t1 > t2 {
		t1, t2 = t2, t1
	}

	if t2 <= 0 {
		return model3d.RayCollision{}, nil, false
	}

	var returnT float64

	returnT = t1
	if t1 < 0 {
		returnT = t2
	}

	point := r.Origin.Add(r.Direction.Scale(returnT))
	normal := point.Sub(s.Center).Normalize()

	return model3d.RayCollision{Normal: normal, Scale: returnT}, s.Material, true
}

// A Plane is an Object implementing an infinity, flat
// plane.
//
// Points x on the plane satisfy: x*Normal - Bias = 0.
type Plane struct {
	Normal   model3d.Coord3D
	Bias     float64
	Material Material
}

// Cast gets the collision with r and the plane.
func (p *Plane) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	// Want to solve for t such that:
	//
	//     (o+t*d)*n - b = 0
	//     o*n + t*(d*n) - b = 0
	//     t = (b - o*n) / (d*n)
	//
	dDot := r.Direction.Dot(p.Normal)

	// Rays parallel to plane have no intersection.
	if math.Abs(dDot) < 1e-8*r.Direction.Norm()*p.Normal.Norm() {
		return model3d.RayCollision{}, nil, false
	}

	scale := (p.Bias - r.Origin.Dot(p.Normal)) / dDot
	if scale < 0 {
		return model3d.RayCollision{}, nil, false
	}

	return model3d.RayCollision{
		Scale:  scale,
		Normal: p.Normal.Normalize(),
	}, p.Material, true
}

// A Cylinder is an Object implementing a perfect round
// cylinder with caps on the ends and a single material.
//
// Defined the same way as model3d.CylinderSolid.
type Cylinder struct {
	P1       model3d.Coord3D
	P2       model3d.Coord3D
	Radius   float64
	Material Material
}

// Cast casts the ray onto the surface of the cylinder.
func (c *Cylinder) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	var hasOne bool
	var nearest model3d.RayCollision
	c.CastAll(r, func(rc model3d.RayCollision, m Material) {
		if !hasOne || rc.Scale < nearest.Scale {
			nearest = rc
			hasOne = true
		}
	})
	return nearest, c.Material, hasOne
}

// CastAll gets all of the ray collisions.
func (c *Cylinder) CastAll(r *model3d.Ray, f func(model3d.RayCollision, Material)) {
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
				f(model3d.RayCollision{
					Scale:  t,
					Normal: p.Sub(v.Scale(frac)).Normalize(),
				}, c.Material)
			}
		}
	}

	// Now detect collisions at the tips.
	plane := Plane{
		Normal:   c.P2.Sub(c.P1),
		Material: c.Material,
	}
	for i, tip := range []model3d.Coord3D{c.P1, c.P2} {
		plane.Bias = plane.Normal.Dot(tip)
		coll, mat, ok := plane.Cast(r)
		if ok {
			p := r.Origin.Add(r.Direction.Scale(coll.Scale))
			if p.Dist(tip) > c.Radius {
				continue
			}
			if i == 0 {
				coll.Normal = coll.Normal.Scale(-1)
			}
			f(coll, mat)
		}
	}
}
