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
func (m *Mesh) Volume() float64 {
	var result float64
	m.Iterate(func(t *Triangle) {
		result += 1.0 / 6.0 * t[0].Cross(t[1]).Dot(t[2])
	})
	return math.Abs(result)
}
