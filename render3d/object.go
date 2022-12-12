package render3d

import (
	"math"
	"math/rand"
	"sort"

	"github.com/unixpickle/model3d/model3d"
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

// Max gets the maximum of the bounding box.
func (c *ColliderObject) Max() model3d.Coord3D {
	return c.Collider.Max()
}

// Cast returns the first ray collision.
func (c *ColliderObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	coll, ok := c.Collider.FirstRayCollision(r)
	return coll, c.Material, ok
}

// ParticipatingMedium is a volume in which a ray has a
// probability of hitting a particle, in which the
// collision probability increases with distance.
//
// It is recommended that you use an HGMaterial with this
// object type.
//
// Normals reported for collisions are random and have no
// bearing on how rays are scattered, since the medium
// simulates complex particles which either reflect or
// refract light.
// Hence, materials which use normals should not be
// employed.
type ParticipatingMedium struct {
	Collider model3d.Collider
	Material Material

	// Lambda controls how likely a collision is.
	// Larger lambda means lower probability.
	// Mean distance is 1 / lambda.
	Lambda float64
}

// Min gets the minimum of the bounding box.
func (p *ParticipatingMedium) Min() model3d.Coord3D {
	return p.Collider.Min()
}

// Max gets the maximum of the bounding box.
func (p *ParticipatingMedium) Max() model3d.Coord3D {
	return p.Collider.Max()
}

// Cast returns the first probabilistic ray collision.
func (p *ParticipatingMedium) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	t := -math.Log(rand.Float64()) / p.Lambda
	t /= r.Direction.Norm()

	var collisions []model3d.RayCollision
	p.Collider.RayCollisions(r, func(rc model3d.RayCollision) {
		collisions = append(collisions, rc)
	})
	sort.Slice(collisions, func(i, j int) bool {
		return collisions[i].Scale < collisions[j].Scale
	})

	inside := len(collisions)%2 == 1
	lastT := 0.0
	for _, c := range collisions {
		if inside {
			passed := c.Scale - lastT
			t -= passed
			if t < 0 {
				return model3d.RayCollision{
					Scale: c.Scale + t,

					// Normal could be anything, but we randomize
					// it so that the normal cosine term is very
					// unlikely to be 0.
					Normal: model3d.NewCoord3DRandUnit(),
				}, p.Material, true
			}
		}
		inside = !inside
		lastT = c.Scale
	}

	return model3d.RayCollision{}, nil, false
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

// Max gets the maximum of the bounding box.
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

// A FilteredObject only reports ray collisions if the ray
// collides with some bounds object.
type FilteredObject struct {
	Object
	Bounds model3d.Collider
}

func (f *FilteredObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	if _, ok := f.Bounds.FirstRayCollision(r); !ok {
		return model3d.RayCollision{}, nil, false
	}
	return f.Object.Cast(r)
}

// BVHToObject creates a single object from a BVH by
// wrapping each node in a (bounding-box filtered) joined
// object.
func BVHToObject(objs *model3d.BVH[Object]) Object {
	if objs.Leaf != nil {
		return objs.Leaf
	} else {
		var branches JoinedObject
		for _, x := range objs.Branch {
			branches = append(branches, BVHToObject(x))
		}
		return &FilteredObject{
			Object: branches,
			Bounds: model3d.BoundsRect(branches),
		}
	}
}
