package render3d

import (
	"github.com/unixpickle/model3d"
)

// An Object is a renderable 3D object.
type Object interface {
	model3d.Bounder

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

// Min gets the minimum of the bounding box.
func (c *ColliderObject) Min() model3d.Coord3D {
	return c.Collider.Min()
}

// Max gets the minimum of the bounding box.
func (c *ColliderObject) Max() model3d.Coord3D {
	return c.Collider.Max()
}

// Cast returns the first ray collision.
func (c *ColliderObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	coll, ok := c.Collider.FirstRayCollision(r)
	return coll, c.Material, ok
}

// A JoinedObject combines multiple Objects.
type JoinedObject []Object

// Min gets the minimum of the bounding box.
func (j JoinedObject) Min() model3d.Coord3D {
	min := j[0].Min()
	for _, x := range j[1:] {
		min = min.Min(x.Min())
	}
	return min
}

// Max gets the minimum of the bounding box.
func (j JoinedObject) Max() model3d.Coord3D {
	max := j[0].Max()
	for _, x := range j[1:] {
		max = max.Max(x.Max())
	}
	return max
}

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
