package model3d

// Matrix2 is a 2x2 matrix, stored in row-major order.
type Matrix2 [4]float64

// NewMatrix2Columns creates a Matrix2 with the given
// coordinates as column entries.
func NewMatrix2Columns(c1, c2 Coord2D) *Matrix2 {
	return &Matrix2{
		c1.X, c2.X,
		c1.Y, c2.Y,
	}
}

// Det computes the determinant of the matrix.
func (m *Matrix2) Det() float64 {
	return m[0]*m[3] - m[1]*m[2]
}

// Scale scales m by a factor s.
func (m *Matrix2) Scale(s float64) {
	for i, x := range m {
		m[i] = x * s
	}
}

// Inverse computes the inverse matrix.
func (m *Matrix2) Inverse() *Matrix2 {
	coeff := 1 / m.Det()
	res := Matrix2{
		m[3], -m[1],
		-m[2], m[0],
	}
	res.Scale(coeff)
	return &res
}

// Mul computes m*m1 and returns the product.
func (m *Matrix2) Mul(m1 *Matrix2) *Matrix2 {
	return &Matrix2{
		m[0]*m1[0] + m[1]*m1[2],
		m[0]*m1[1] + m[1]*m1[3],
		m[2]*m1[0] + m[3]*m1[2],
		m[2]*m1[1] + m[3]*m1[3],
	}
}

// MulColumn multiplies the matrix m by a column vector
// represented by c.
func (m *Matrix2) MulColumn(c Coord2D) Coord2D {
	return Coord2D{
		X: m[0]*c.X + m[1]*c.Y,
		Y: m[2]*c.X + m[3]*c.Y,
	}
}

// Matrix3 is a 3x3 matrix, stored in row-major order.
type Matrix3 [9]float64

// NewMatrix3Columns creates a Matrix3 with the given
// coordinates as column entries.
func NewMatrix3Columns(c1, c2, c3 Coord3D) *Matrix3 {
	return &Matrix3{
		c1.X, c2.X, c3.X,
		c1.Y, c2.Y, c3.Y,
		c1.Z, c2.Z, c3.Z,
	}
}

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
