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
	return InSolidBounds(t, c) && t.solid.Contains(c.Sub(t.offset))
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
