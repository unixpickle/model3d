package model2d

import "math"

// Area computes the area inside of a manifold mesh.
func (m *Mesh) Area() float64 {
	var result float64
	m.Iterate(func(s *Segment) {
		mat := Matrix2{
			s[0].X, s[0].Y,
			s[1].X, s[1].Y,
		}
		result += mat.Det() / 2.0
	})
	return math.Abs(result)
}
