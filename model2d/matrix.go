package model2d

// Matrix2 is a 2x2 matrix, stored in row-major order.
type Matrix2 [4]float64

// NewMatrix2Columns creates a Matrix2 with the given
// coordinates as column entries.
func NewMatrix2Columns(c1, c2 Coord) *Matrix2 {
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
func (m *Matrix2) MulColumn(c Coord) Coord {
	return Coord{
		X: m[0]*c.X + m[1]*c.Y,
		Y: m[2]*c.X + m[3]*c.Y,
	}
}
