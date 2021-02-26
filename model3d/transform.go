// Generated from templates/transform.template

package model3d

// Transform is an invertible coordinate transformation.
type Transform interface {
	// Apply applies the transformation to c.
	Apply(c Coord3D) Coord3D

	// ApplyBounds gets a new bounding rectangle that is
	// guaranteed to bound the old bounding rectangle when
	// it is transformed.
	ApplyBounds(min, max Coord3D) (Coord3D, Coord3D)

	// Inverse gets an inverse transformation.
	//
	// The inverse may not perfectly invert bounds
	// transformations, since some information may be lost
	// during such a transformation.
	Inverse() Transform
}

// DistTransform is a Transform that changes Euclidean
// distances in a coordinate-independent fashion.
//
// The inverse of a DistTransform should also be a
// DistTransform.
type DistTransform interface {
	Transform

	// ApplyDistance computes the distance between
	// t.Apply(c1) and t.Apply(c2) given the distance
	// between c1 and c2, where c1 and c2 are arbitrary
	// points.
	ApplyDistance(d float64) float64
}

// Translate is a Transform that adds an offset to
// coordinates.
type Translate struct {
	Offset Coord3D
}

func (t *Translate) Apply(c Coord3D) Coord3D {
	return c.Add(t.Offset)
}

func (t *Translate) ApplyBounds(min, max Coord3D) (Coord3D, Coord3D) {
	return min.Add(t.Offset), max.Add(t.Offset)
}

func (t *Translate) Inverse() Transform {
	return &Translate{Offset: t.Offset.Scale(-1)}
}

func (t *Translate) ApplyDistance(d float64) float64 {
	return d
}

// Matrix3Transform is a Transform that applies a matrix
// to coordinates.
type Matrix3Transform struct {
	Matrix *Matrix3
}

func (m *Matrix3Transform) Apply(c Coord3D) Coord3D {
	return m.Matrix.MulColumn(c)
}

func (m *Matrix3Transform) ApplyBounds(min, max Coord3D) (Coord3D, Coord3D) {
	var newMin, newMax Coord3D
	for i, x := range []float64{min.X, max.X} {
		for j, y := range []float64{min.Y, max.Y} {
			for k, z := range []float64{min.Z, max.Z} {
				c := m.Matrix.MulColumn(XYZ(x, y, z))
				if i == 0 && j == 0 && k == 0 {
					newMin, newMax = c, c
				} else {
					newMin = newMin.Min(c)
					newMax = newMax.Max(c)
				}
			}
		}
	}
	return newMin, newMax
}

func (m *Matrix3Transform) Inverse() Transform {
	return &Matrix3Transform{Matrix: m.Matrix.Inverse()}
}

// orthoMatrix3Transform is a Matrix3Transform
// for distance-preserving matrices.
type orthoMatrix3Transform struct {
	Matrix3Transform
}

// Rotation creates a rotation transformation using an
// angle in radians around a unit vector direction.
func Rotation(axis Coord3D, theta float64) DistTransform {
	return &orthoMatrix3Transform{
		Matrix3Transform{
			Matrix: NewMatrix3Rotation(axis, theta),
		},
	}
}

func (m *orthoMatrix3Transform) ApplyDistance(c float64) float64 {
	return c
}

// A JoinedTransform composes transformations from left to
// right.
type JoinedTransform []Transform

func (j JoinedTransform) Apply(c Coord3D) Coord3D {
	for _, t := range j {
		c = t.Apply(c)
	}
	return c
}

func (j JoinedTransform) ApplyBounds(min Coord3D, max Coord3D) (Coord3D, Coord3D) {
	for _, t := range j {
		min, max = t.ApplyBounds(min, max)
	}
	return min, max
}

func (j JoinedTransform) Inverse() Transform {
	res := JoinedTransform{}
	for i := len(j) - 1; i >= 0; i-- {
		res = append(res, j[i].Inverse())
	}
	return res
}

// ApplyDistance transforms a distance.
//
// It panic()s if any transforms don't implement
// DistTransform.
func (j JoinedTransform) ApplyDistance(d float64) float64 {
	for _, t := range j {
		d = t.(DistTransform).ApplyDistance(d)
	}
	return d
}

// Scale is a transform that scales an object.
type Scale struct {
	Scale float64
}

func (s *Scale) Apply(c Coord3D) Coord3D {
	return c.Scale(s.Scale)
}

func (s *Scale) ApplyBounds(min Coord3D, max Coord3D) (Coord3D, Coord3D) {
	return min.Scale(s.Scale), max.Scale(s.Scale)
}

func (s *Scale) Inverse() Transform {
	return &Scale{Scale: 1 / s.Scale}
}

func (s *Scale) ApplyDistance(d float64) float64 {
	return d * s.Scale
}

// ScaleSolid creates a new Solid that scales incoming
// coordinates c by 1/s.
// Thus, the new solid is s times larger.
func ScaleSolid(solid Solid, s float64) Solid {
	return TransformSolid(&Scale{Scale: s}, solid)
}

// TransformSolid applies t to the solid s to produce a
// new, transformed solid.
func TransformSolid(t Transform, s Solid) Solid {
	min, max := t.ApplyBounds(s.Min(), s.Max())
	return &transformedSolid{
		min: min,
		max: max,
		s:   s,
		inv: t.Inverse(),
	}
}

// TransformSDF applies t to the SDF s to produce a new,
// transformed SDF.
func TransformSDF(t DistTransform, s SDF) SDF {
	min, max := t.ApplyBounds(s.Min(), s.Max())
	return &transformedSDF{
		min: min,
		max: max,
		s:   s,
		t:   t,
		inv: t.Inverse().(DistTransform),
	}
}

// TransformCollider applies t to the Collider c to
// produce a new, transformed Collider.
func TransformCollider(t DistTransform, c Collider) Collider {
	min, max := t.ApplyBounds(c.Min(), c.Max())
	return &transformedCollider{
		min: min,
		max: max,
		c:   c,
		t:   t,
		inv: t.Inverse().(DistTransform),
	}
}

type transformedSolid struct {
	min Coord3D
	max Coord3D
	s   Solid
	inv Transform
}

func (t *transformedSolid) Min() Coord3D {
	return t.min
}

func (t *transformedSolid) Max() Coord3D {
	return t.max
}

func (t *transformedSolid) Contains(c Coord3D) bool {
	return InBounds(t, c) && t.s.Contains(t.inv.Apply(c))
}

type transformedSDF struct {
	min Coord3D
	max Coord3D
	s   SDF
	t   DistTransform
	inv DistTransform
}

func (t *transformedSDF) Min() Coord3D {
	return t.min
}

func (t *transformedSDF) Max() Coord3D {
	return t.max
}

func (t *transformedSDF) SDF(c Coord3D) float64 {
	return t.t.ApplyDistance(t.s.SDF(t.inv.Apply(c)))
}

type transformedCollider struct {
	min Coord3D
	max Coord3D
	c   Collider
	t   DistTransform
	inv DistTransform
}

func (t *transformedCollider) Min() Coord3D {
	return t.min
}

func (t *transformedCollider) Max() Coord3D {
	return t.max
}

func (t *transformedCollider) RayCollisions(r *Ray, f func(RayCollision)) int {
	return t.c.RayCollisions(t.innerRay(r), func(rc RayCollision) {
		f(t.outerCollision(rc))
	})
}

func (t *transformedCollider) FirstRayCollision(r *Ray) (RayCollision, bool) {
	rc, collides := t.c.FirstRayCollision(t.innerRay(r))
	return t.outerCollision(rc), collides
}

func (t *transformedCollider) SphereCollision(c Coord3D, r float64) bool {
	return t.c.SphereCollision(t.inv.Apply(c), t.inv.ApplyDistance(r))
}

func (t *transformedCollider) innerRay(r *Ray) *Ray {
	return &Ray{
		Origin:    t.inv.Apply(r.Origin),
		Direction: t.inv.Apply(r.Direction),
	}
}

func (t *transformedCollider) outerCollision(rc RayCollision) RayCollision {
	return RayCollision{
		Scale:  t.t.ApplyDistance(rc.Scale),
		Normal: t.t.Apply(rc.Normal),
		Extra:  rc.Extra,
	}
}
