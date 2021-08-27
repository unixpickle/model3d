package render3d

import "github.com/unixpickle/model3d/model3d"

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

// Rotate creates a new Object by rotating an Object by
// a given angle (in radians) around a given (unit) axis.
func Rotate(obj Object, axis model3d.Coord3D, angle float64) Object {
	return MatrixMultiply(obj, model3d.NewMatrix3Rotation(axis, angle))
}

// Scale creates a new Object by scaling an Object by the
// given factor along all axes.
func Scale(obj Object, scale float64) Object {
	return MatrixMultiply(obj, &model3d.Matrix3{scale, 0, 0, 0, scale, 0, 0, 0, scale})
}

// MatrixMultiply left-multiplies coordinates in an object
// by a matrix m.
// It can be used for rotations, scaling, etc.
func MatrixMultiply(obj Object, m *model3d.Matrix3) Object {
	transform := &model3d.Matrix3Transform{Matrix: m}
	min, max := transform.ApplyBounds(obj.Min(), obj.Max())
	return &matrixObject{
		Object:  obj,
		MinVal:  min,
		MaxVal:  max,
		Matrix:  m,
		Inverse: m.Inverse(),
	}
}

type matrixObject struct {
	Object  Object
	MinVal  model3d.Coord3D
	MaxVal  model3d.Coord3D
	Matrix  *model3d.Matrix3
	Inverse *model3d.Matrix3
}

func (m *matrixObject) Min() model3d.Coord3D {
	return m.MinVal
}

func (m *matrixObject) Max() model3d.Coord3D {
	return m.MaxVal
}

func (m *matrixObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	rc, mat, ok := m.Object.Cast(&model3d.Ray{
		Origin:    m.Inverse.MulColumn(r.Origin),
		Direction: m.Inverse.MulColumn(r.Direction),
	})
	if ok {
		rc.Normal = m.Matrix.MulColumn(rc.Normal).Normalize()
	}
	return rc, mat, ok
}
