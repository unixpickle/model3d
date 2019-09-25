package model3d

// A 3x3 matrix, stored in row-major order.
type Matrix3 [9]float64

// Det computes the determinant of the matrix.
func (m *Matrix3) Det() float64 {
	return m[0]*(m[4]*m[8]-m[5]*m[7]) - m[1]*(m[3]*m[8]-m[5]*m[6]) + m[2]*(m[3]*m[7]-m[4]*m[6])
}

// Scale scales m by a factor s.
func (m *Matrix3) Scale(s float64) {
	for i, x := range m {
		m[i] = x * s
	}
}

// Inverse computes the inverse matrix.
func (m *Matrix3) Inverse() *Matrix3 {
	coeff := 1 / m.Det()
	res := Matrix3{
		m[4]*m[8] - m[5]*m[7], m[2]*m[7] - m[1]*m[8], m[1]*m[5] - m[2]*m[4],
		m[5]*m[6] - m[3]*m[8], m[0]*m[8] - m[2]*m[6], m[2]*m[3] - m[0]*m[5],
		m[3]*m[7] - m[4]*m[6], m[1]*m[6] - m[0]*m[7], m[0]*m[4] - m[1]*m[3],
	}
	res.Scale(coeff)
	return &res
}

// Mul computes m*m1 and returns the product.
func (m *Matrix3) Mul(m1 *Matrix3) *Matrix3 {
	return &Matrix3{
		m[0]*m1[0] + m[1]*m1[3] + m[2]*m1[6],
		m[0]*m1[1] + m[1]*m1[4] + m[2]*m1[7],
		m[0]*m1[2] + m[1]*m1[5] + m[2]*m1[8],

		m[3]*m1[0] + m[4]*m1[3] + m[5]*m1[6],
		m[3]*m1[1] + m[4]*m1[4] + m[5]*m1[7],
		m[3]*m1[2] + m[4]*m1[5] + m[5]*m1[8],

		m[6]*m1[0] + m[7]*m1[3] + m[8]*m1[6],
		m[6]*m1[1] + m[7]*m1[4] + m[8]*m1[7],
		m[6]*m1[2] + m[7]*m1[5] + m[8]*m1[8],
	}
}

// MulColumn multiplies the matrix m by a column vector
// represented by c.
func (m *Matrix3) MulColumn(c Coord3D) Coord3D {
	return Coord3D{
		X: m[0]*c.X + m[1]*c.Y + m[2]*c.Z,
		Y: m[3]*c.X + m[4]*c.Y + m[5]*c.Z,
		Z: m[6]*c.X + m[7]*c.Y + m[8]*c.Z,
	}
}
