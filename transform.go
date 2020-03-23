package model3d

// Transform is an invertible coordinate transformation.
type Transform interface {
	Apply(c Coord3D) Coord3D
	ApplySolid(s Solid) Solid
	Inverse() Transform
}

// Translate is a Transform that adds an offset to
// coordinates.
type Translate struct {
	Offset Coord3D
}

func (t *Translate) Apply(c Coord3D) Coord3D {
	return c.Add(t.Offset)
}

func (t *Translate) ApplySolid(s Solid) Solid {
	return &translatedSolid{
		offset: t.Offset,
		min:    s.Min().Add(t.Offset),
		max:    s.Max().Add(t.Offset),
		solid:  s,
	}
}

func (t *Translate) Inverse() Transform {
	return &Translate{Offset: t.Offset.Scale(-1)}
}

type translatedSolid struct {
	offset Coord3D
	min    Coord3D
	max    Coord3D
	solid  Solid
}

func (t *translatedSolid) Min() Coord3D {
	return t.min
}

func (t *translatedSolid) Max() Coord3D {
	return t.max
}

func (t *translatedSolid) Contains(c Coord3D) bool {
	return InBounds(t, c) && t.solid.Contains(c.Sub(t.offset))
}

// Matrix3Transform is a Transform that applies a matrix
// to coordinates.
type Matrix3Transform struct {
	Matrix *Matrix3
}

func (m *Matrix3Transform) Apply(c Coord3D) Coord3D {
	return m.Matrix.MulColumn(c)
}

func (m *Matrix3Transform) ApplySolid(s Solid) Solid {
	min := s.Min()
	max := s.Max()
	var newMin, newMax Coord3D
	for i, x := range []float64{min.X, max.X} {
		for j, y := range []float64{min.Y, max.Y} {
			for k, z := range []float64{min.Z, max.Z} {
				c := m.Matrix.MulColumn(Coord3D{X: x, Y: y, Z: z})
				if i == 0 && j == 0 && k == 0 {
					newMin, newMax = c, c
				} else {
					newMin = newMin.Min(c)
					newMax = newMax.Max(c)
				}
			}
		}
	}
	return &matrix3Solid{
		inverse: m.Matrix.Inverse(),
		min:     newMin,
		max:     newMax,
		solid:   s,
	}
}

func (m *Matrix3Transform) Inverse() Transform {
	return &Matrix3Transform{Matrix: m.Matrix.Inverse()}
}

type matrix3Solid struct {
	inverse *Matrix3
	min     Coord3D
	max     Coord3D
	solid   Solid
}

func (m *matrix3Solid) Min() Coord3D {
	return m.min
}

func (m *matrix3Solid) Max() Coord3D {
	return m.max
}

func (m *matrix3Solid) Contains(c Coord3D) bool {
	return InBounds(m, c) && m.solid.Contains(m.inverse.MulColumn(c))
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

func (j JoinedTransform) ApplySolid(s Solid) Solid {
	for _, t := range j {
		s = t.ApplySolid(s)
	}
	return s
}

func (j JoinedTransform) Inverse() Transform {
	res := JoinedTransform{}
	for i := len(j) - 1; i >= 0; i-- {
		res = append(res, j[i].Inverse())
	}
	return res
}
