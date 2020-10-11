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
	min Coord
	max Coord
	s   Solid
	inv Transform
}

func (t *transformedSolid) Min() Coord {
	return t.min
}

func (t *transformedSolid) Max() Coord {
	return t.max
}

func (t *transformedSolid) Contains(c Coord) bool {
	return InBounds(t, c) && t.s.Contains(t.inv.Apply(c))
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
			// add-codegen: for k, z := range []float64{min.Z, max.Z} {
			c := m.Matrix.MulColumn(XY(x, y))
			// replace-codegen: c := m.Matrix.MulColumn(XYZ(x, y, z))
			if i == 0 && j == 0 {
				// replace-codegen: if i == 0 && j == 0 && k == 0 {
				newMin, newMax = c, c
			} else {
				newMin = newMin.Min(c)
				newMax = newMax.Max(c)
			}
			// add-codegen: }
		}
	}
	return newMin, newMax
}

func (m *Matrix2Transform) Inverse() Transform {
	return &Matrix2Transform{Matrix: m.Matrix.Inverse()}
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
