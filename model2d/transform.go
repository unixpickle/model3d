// Generated from templates/transform.template

package model2d

// Transform is an invertible coordinate transformation.
type Transform interface {
	// Apply applies the transformation to c.
	Apply(c Coord) Coord

	// ApplyBounds gets a new bounding rectangle that is
	// guaranteed to bound the old bounding rectangle when
	// it is transformed.
	ApplyBounds(min, max Coord) (Coord, Coord)

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
	Offset Coord
}

func (t *Translate) Apply(c Coord) Coord {
	return c.Add(t.Offset)
}

func (t *Translate) ApplyBounds(min, max Coord) (Coord, Coord) {
	return min.Add(t.Offset), max.Add(t.Offset)
}

func (t *Translate) Inverse() Transform {
	return &Translate{Offset: t.Offset.Scale(-1)}
}

func (t *Translate) ApplyDistance(d float64) float64 {
	return d
}

// Matrix2Transform is a Transform that applies a matrix
// to coordinates.
type Matrix2Transform struct {
	Matrix *Matrix2
}

func (m *Matrix2Transform) Apply(c Coord) Coord {
	return m.Matrix.MulColumn(c)
}

func (m *Matrix2Transform) ApplyBounds(min, max Coord) (Coord, Coord) {
	var newMin, newMax Coord
	for i, x := range []float64{min.X, max.X} {
		for j, y := range []float64{min.Y, max.Y} {
			c := m.Matrix.MulColumn(XY(x, y))
			if i == 0 && j == 0 {
				newMin, newMax = c, c
			} else {
				newMin = newMin.Min(c)
				newMax = newMax.Max(c)
			}
		}
	}
	return newMin, newMax
}

func (m *Matrix2Transform) Inverse() Transform {
	return &Matrix2Transform{Matrix: m.Matrix.Inverse()}
}

// orthoMatrix2Transform is a Matrix2Transform
// for distance-preserving matrices.
type orthoMatrix2Transform struct {
	Matrix2Transform
}

// Rotation creates a rotation transformation using an
// angle in radians.
func Rotation(theta float64) DistTransform {
	return &orthoMatrix2Transform{
		Matrix2Transform{
			Matrix: NewMatrix2Rotation(theta),
		},
	}
}

func (m *orthoMatrix2Transform) ApplyDistance(c float64) float64 {
	return c
}

func (m *orthoMatrix2Transform) Inverse() Transform {
	return &orthoMatrix2Transform{*m.Matrix2Transform.Inverse().(*Matrix2Transform)}
}

// A JoinedTransform composes transformations from left to
// right.
type JoinedTransform []Transform

func (j JoinedTransform) Apply(c Coord) Coord {
	for _, t := range j {
		c = t.Apply(c)
	}
	return c
}

func (j JoinedTransform) ApplyBounds(min Coord, max Coord) (Coord, Coord) {
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

func (s *Scale) Apply(c Coord) Coord {
	return c.Scale(s.Scale)
}

func (s *Scale) ApplyBounds(min Coord, max Coord) (Coord, Coord) {
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

// TranslateSolid creates a new Solid by translating a
// Solid by a given offset.
func TranslateSolid(solid Solid, offset Coord) Solid {
	return TransformSolid(&Translate{Offset: offset}, solid)
}

// RotateSolid creates a new Solid by rotating a Solid by
// a given angle (in radians).
func RotateSolid(solid Solid, angle float64) Solid {
	return TransformSolid(Rotation(angle), solid)
}

// TransformSolid applies t to the solid s to produce a
// new, transformed solid.
func TransformSolid(t Transform, s Solid) Solid {
	inv := t.Inverse()
	min, max := t.ApplyBounds(s.Min(), s.Max())
	return CheckedFuncSolid(min, max, func(c Coord) bool {
		return s.Contains(inv.Apply(c))
	})
}

// TransformSDF applies t to the SDF s to produce a new,
// transformed SDF.
func TransformSDF(t DistTransform, s SDF) SDF {
	inv := t.Inverse()
	min, max := t.ApplyBounds(s.Min(), s.Max())
	return FuncSDF(min, max, func(c Coord) float64 {
		return t.ApplyDistance(s.SDF(inv.Apply(c)))
	})
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

type transformedCollider struct {
	min Coord
	max Coord
	c   Collider
	t   DistTransform
	inv DistTransform
}

func (t *transformedCollider) Min() Coord {
	return t.min
}

func (t *transformedCollider) Max() Coord {
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

func (t *transformedCollider) CircleCollision(c Coord, r float64) bool {
	return t.c.CircleCollision(t.inv.Apply(c), t.inv.ApplyDistance(r))
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
