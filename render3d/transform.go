package render3d

import "github.com/unixpickle/model3d"

// Translate moves the object by an additive offset.
func Translate(obj Object, offset model3d.Coord3D) Object {
	return &translatedObject{
		Object: obj,
		Offset: offset,
	}
}

type translatedObject struct {
	Object Object
	Offset model3d.Coord3D
}

func (t *translatedObject) Min() model3d.Coord3D {
	return t.Object.Min().Add(t.Offset)
}

func (t *translatedObject) Max() model3d.Coord3D {
	return t.Object.Max().Add(t.Offset)
}

func (t *translatedObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	return t.Object.Cast(&model3d.Ray{
		Origin:    r.Origin.Sub(t.Offset),
		Direction: r.Direction,
	})
}
