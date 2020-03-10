package model3d

// Transform is an invertible coordinate transformation.
type Transform interface {
	Apply(c Coord3D) Coord3D
	ApplySolid(s Solid) Solid
	Inverse() Transform
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
