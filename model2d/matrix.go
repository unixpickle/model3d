package model2d

import (
	"math"
	"math/cmplx"
)

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

// NewMatrix2Rotation creates a rotation matrix that
// rotates column vectors by theta.
func NewMatrix2Rotation(theta float64) *Matrix2 {
	cos := math.Cos(theta)
	sin := math.Sin(theta)
	return &Matrix2{cos, -sin, sin, cos}
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
	res := *m
	res.InvertInPlace()
	return &res
}

// InvertInPlace moves the inverse of m into m without
// causing any new allocations.
func (m *Matrix2) InvertInPlace() {
	m.InvertInPlaceDet(m.Det())
}

// InvertInPlaceDet is an optimization for InvertInPlace
// when the determinant has been pre-computed.
func (m *Matrix2) InvertInPlaceDet(det float64) {
	*m = Matrix2{
		m[3], -m[1],
		-m[2], m[0],
	}
	m.Scale(1 / det)
}

// MulColumnInv multiplies the inverse of m by the column
// c, given the determinant of m.
func (m *Matrix2) MulColumnInv(c Coord, det float64) Coord {
	m1 := Matrix2{
		m[3], -m[1],
		-m[2], m[0],
	}
	return m1.MulColumn(c.Scale(1 / det))
}

// Add computes m+m1 and returns the sum.
func (m *Matrix2) Add(m1 *Matrix2) *Matrix2 {
	var res Matrix2
	for i, x := range m {
		res[i] = x + m1[i]
	}
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

// Transpose computes the matrix transpose.
func (m *Matrix2) Transpose() *Matrix2 {
	return &Matrix2{
		m[0], m[2],
		m[1], m[3],
	}
}

// Eigenvalues computes the eigenvalues of the matrix.
//
// There may be a repeated eigenvalue, but for numerical
// reasons two are always returned.
func (m *Matrix2) Eigenvalues() [2]complex128 {
	// Quadratic formula for the characteristic polynomial.
	a := complex128(1)
	b := complex(-(m[0] + m[3]), 0)
	c := complex(m.Det(), 0)
	sqrtDisc := cmplx.Sqrt(b*b - 4*a*c)
	return [2]complex128{
		(-b - sqrtDisc) / (2 * a),
		(-b + sqrtDisc) / (2 * a),
	}
}

// SVD computes the singular value decomposition of the
// matrix.
//
// It populates matrices u, s, and v, such that
//
//     m = u*s*v.Transpose()
//
// The singular values in s are sorted largest to
// smallest.
func (m *Matrix2) SVD(u, s, v *Matrix2) {
	ata := m.Transpose().Mul(m)
	aat := m.Mul(m.Transpose())
	eigVals := ata.Eigenvalues()

	if real(eigVals[0]) < real(eigVals[1]) {
		eigVals[0], eigVals[1] = eigVals[1], eigVals[0]
	}

	v1, v2 := ata.symEigs(eigVals)

	var u1, u2 Coord
	u1 = m.MulColumn(v1)
	if n := u1.Norm(); n == 0 {
		u1, u2 = aat.symEigs(eigVals)
	} else {
		u1 = u1.Scale(1 / n)
		u2 = Coord{X: -u1.Y, Y: u1.X}
	}

	*s = Matrix2{
		math.Sqrt(math.Max(0, real(eigVals[0]))), 0,
		0, math.Sqrt(math.Max(0, real(eigVals[1]))),
	}
	if m.MulColumn(v1).Dot(u1) < 0 {
		u1 = u1.Scale(-1)
	}
	if m.MulColumn(v2).Dot(u2) < 0 {
		u2 = u2.Scale(-1)
	}

	*u = Matrix2{
		u1.X, u2.X,
		u1.Y, u2.Y,
	}
	*v = Matrix2{
		v1.X, v2.X,
		v1.Y, v2.Y,
	}
}

// symEigs computes the eigenvectors for the eigenvalues
// of a symmetric matrix.
func (m *Matrix2) symEigs(vals [2]complex128) (v1, v2 Coord) {
	r1 := Coord{X: m[0] - real(vals[0]), Y: m[1]}
	r2 := Coord{X: m[2], Y: m[3] - real(vals[0])}
	n1, n2 := r1.Norm(), r2.Norm()
	if n1 == 0 && n2 == 0 {
		return Coord{X: 1}, Coord{Y: 1}
	}

	secondEig := r1.Scale(1 / n1)
	if n2 > n1 {
		secondEig = r2.Scale(1 / n2)
	}
	firstEig := Coord{X: -secondEig.Y, Y: secondEig.X}
	return firstEig, secondEig
}
