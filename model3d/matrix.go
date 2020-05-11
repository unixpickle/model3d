package model3d

import (
	"math"
	"math/cmplx"

	"github.com/unixpickle/model3d/model2d"
)

type Matrix2 = model2d.Matrix2

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

// NewMatrix3Rotation creates a 3D rotation matrix.
// Points are rotated around the given vector in a
// right-handed direction.
//
// The axis is assumed to be normalized.
// The angle is measured in radians.
func NewMatrix3Rotation(axis Coord3D, angle float64) *Matrix3 {
	x, y := axis.OrthoBasis()
	basis := NewMatrix3Columns(axis, x, y)
	rotation := &Matrix3{
		1, 0, 0,
		0, math.Cos(angle), math.Sin(angle),
		0, -math.Sin(angle), math.Cos(angle),
	}
	return basis.Mul(rotation).Mul(basis.Transpose())
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
	res := *m
	res.InvertInPlace()
	return &res
}

// InvertInPlace moves the inverse of m into m without
// causing any new allocations.
func (m *Matrix3) InvertInPlace() {
	coeff := 1 / m.Det()
	*m = Matrix3{
		m[4]*m[8] - m[5]*m[7], m[2]*m[7] - m[1]*m[8], m[1]*m[5] - m[2]*m[4],
		m[5]*m[6] - m[3]*m[8], m[0]*m[8] - m[2]*m[6], m[2]*m[3] - m[0]*m[5],
		m[3]*m[7] - m[4]*m[6], m[1]*m[6] - m[0]*m[7], m[0]*m[4] - m[1]*m[3],
	}
	m.Scale(coeff)
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

// Transpose computes the matrix transpose.
func (m *Matrix3) Transpose() *Matrix3 {
	return &Matrix3{
		m[0], m[3], m[6],
		m[1], m[4], m[7],
		m[2], m[5], m[8],
	}
}

// Eigenvalues computes the eigenvalues of the matrix.
//
// There may be repeated eigenvalues, but for numerical
// reasons three are always returned.
func (m *Matrix3) Eigenvalues() [3]complex128 {
	trace := m[0] + m[4] + m[8]
	sqTrace := (m[0]*m[0] + m[1]*m[3] + m[2]*m[6]) +
		(m[1]*m[3] + m[4]*m[4] + m[7]*m[5]) +
		(m[2]*m[6] + m[5]*m[7] + m[8]*m[8])

	// Characteristic polynomial coefficients.
	a := -complex128(1)
	b := complex(trace, 0)
	c := complex(0.5*(sqTrace-trace*trace), 0)
	d := complex(m.Det(), 0)

	// Cubic formula: https://en.wikipedia.org/wiki/Cubic_equation#General_cubic_formula

	disc0 := b*b - 3*a*c
	disc1 := 2*b*b*b - 9*a*b*c + 27*a*a*d
	addOrSub := cmplx.Sqrt(disc1*disc1 - 4*disc0*disc0*disc0)
	// For numerical stability, choose the C with the largest
	// absolute value.
	c1 := (disc1 + addOrSub) / 2
	c2 := (disc1 - addOrSub) / 2
	bigC := c1
	if cmplx.Abs(c2) > cmplx.Abs(c1) {
		bigC = c2
	}
	bigC = cmplx.Pow(bigC, 1.0/3.0)

	xForPhase := func(phase complex128) complex128 {
		thisC := phase * bigC
		return (-1.0 / (3 * a)) * (b + thisC + disc0/thisC)
	}

	return [3]complex128{
		xForPhase(1),
		xForPhase(-0.5 + 0.8660254037844386i),
		xForPhase(-0.5 - 0.8660254037844386i),
	}
}
