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
