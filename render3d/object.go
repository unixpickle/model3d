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
