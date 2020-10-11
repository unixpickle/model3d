// Generated from model2d/transform.go

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
