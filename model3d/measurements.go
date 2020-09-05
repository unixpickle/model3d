package model3d

import "math"

// Area computes the total surface area of the mesh.
func (m *Mesh) Area() float64 {
	var result float64
	m.Iterate(func(t *Triangle) {
		result += t.Area()
	})
	return result
}

// Volume measures the volume of the mesh.
//
// This assumes that the mesh is manifold and the normals
// are consistent.
func (m *Mesh) Volume() float64 {
	var result float64
	m.Iterate(func(t *Triangle) {
		mat := Matrix3{
			t[0].X, t[0].Y, t[0].Z,
			t[1].X, t[1].Y, t[1].Z,
			t[2].X, t[2].Y, t[2].Z,
		}
		result += mat.Det() / 6.0
	})
	return math.Abs(result)
}
