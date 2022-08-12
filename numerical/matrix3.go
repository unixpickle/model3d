package numerical

import (
	"math"
	"math/cmplx"
)

// Matrix3 is a 3x3 matrix, stored in row-major order.
type Matrix3 [9]float64

// NewMatrix3Columns creates a Matrix3 with the given
// coordinates as column entries.
func NewMatrix3Columns(c1, c2, c3 Vec3) *Matrix3 {
	return &Matrix3{
		c1[0], c2[0], c3[0],
		c1[1], c2[1], c3[1],
		c1[2], c2[2], c3[2],
	}
}

// NewMatrix3Rotation creates a 3D rotation matrix.
// Points are rotated around the given vector in a
// right-handed direction.
//
// The axis is assumed to be normalized.
// The angle is measured in radians.
func NewMatrix3Rotation(axis Vec3, angle float64) *Matrix3 {
	x, y := axis.OrthoBasis()
	basis := NewMatrix3Columns(axis, x, y)
	rotation := &Matrix3{
		1, 0, 0,
		0, math.Cos(angle), math.Sin(angle),
		0, -math.Sin(angle), math.Cos(angle),
	}
	return basis.Mul(rotation).Mul(basis.Transpose())
}

// Cols returns the columns of the matrix as Vec4.
func (m *Matrix3) Cols() [3]Vec3 {
	return m.Transpose().Rows()
}

// Rows returns the rows of the matrix as Vec3.
func (m *Matrix3) Rows() [3]Vec3 {
	return [3]Vec3{
		{m[0], m[1], m[2]},
		{m[3], m[4], m[5]},
		{m[6], m[7], m[8]},
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
	res := *m
	res.InvertInPlace()
	return &res
}

// InvertInPlace moves the inverse of m into m without
// causing any new allocations.
func (m *Matrix3) InvertInPlace() {
	m.InvertInPlaceDet(m.Det())
}

// InvertInPlaceDet is an optimization for InvertInPlace
// when the determinant has been pre-computed.
func (m *Matrix3) InvertInPlaceDet(det float64) {
	*m = Matrix3{
		m[4]*m[8] - m[5]*m[7], m[2]*m[7] - m[1]*m[8], m[1]*m[5] - m[2]*m[4],
		m[5]*m[6] - m[3]*m[8], m[0]*m[8] - m[2]*m[6], m[2]*m[3] - m[0]*m[5],
		m[3]*m[7] - m[4]*m[6], m[1]*m[6] - m[0]*m[7], m[0]*m[4] - m[1]*m[3],
	}
	m.Scale(1 / det)
}

// MulColumnInv multiplies the inverse of m by the column
// c, given the determinant of m.
func (m *Matrix3) MulColumnInv(c Vec3, det float64) Vec3 {
	m1 := Matrix3{
		m[4]*m[8] - m[5]*m[7], m[2]*m[7] - m[1]*m[8], m[1]*m[5] - m[2]*m[4],
		m[5]*m[6] - m[3]*m[8], m[0]*m[8] - m[2]*m[6], m[2]*m[3] - m[0]*m[5],
		m[3]*m[7] - m[4]*m[6], m[1]*m[6] - m[0]*m[7], m[0]*m[4] - m[1]*m[3],
	}
	return m1.MulColumn(c.Scale(1 / det))
}

// Add computes m+m1 and returns the sum.
func (m *Matrix3) Add(m1 *Matrix3) *Matrix3 {
	var res Matrix3
	for i, x := range m {
		res[i] = x + m1[i]
	}
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
func (m *Matrix3) MulColumn(c Vec3) Vec3 {
	return Vec3{
		m[0]*c[0] + m[1]*c[1] + m[2]*c[2],
		m[3]*c[0] + m[4]*c[1] + m[5]*c[2],
		m[6]*c[0] + m[7]*c[1] + m[8]*c[2],
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
		if thisC == 0 {
			return (-1.0 / (3 * a)) * b
		}
		return (-1.0 / (3 * a)) * (b + thisC + disc0/thisC)
	}

	return [3]complex128{
		xForPhase(1),
		xForPhase(-0.5 + 0.8660254037844386i),
		xForPhase(-0.5 - 0.8660254037844386i),
	}
}

// SVD computes the singular value decomposition of the
// matrix.
//
// It populates matrices u, s, and v, such that
//
//	m = u*s*v.Transpose()
//
// The singular values in s are sorted largest to
// smallest.
func (m *Matrix3) SVD(u, s, v *Matrix3) {
	ata := m.Transpose().Mul(m)
	aat := m.Mul(m.Transpose())
	eigVals := ata.Eigenvalues()

	largestEig := math.Max(real(eigVals[0]), math.Max(real(eigVals[1]), real(eigVals[2])))

	inVector := ata.symEigVector(largestEig)
	outVector := m.MulColumn(inVector)
	if n := outVector.Norm(); n == 0 {
		outVector = aat.symEigVector(largestEig)
		if m.MulColumn(inVector).Dot(outVector) < 0 {
			outVector = outVector.Scale(-1)
		}
	} else {
		outVector = outVector.Scale(1 / n)
	}

	// Find other two singular components using a 2x2
	// matrix in a different basis.
	inBasis1, inBasis2 := inVector.OrthoBasis()
	outBasis1, outBasis2 := outVector.OrthoBasis()
	out1 := m.MulColumn(inBasis1)
	out2 := m.MulColumn(inBasis2)
	mat2x2 := Matrix2{
		outBasis1.Dot(out1), outBasis1.Dot(out2),
		outBasis2.Dot(out1), outBasis2.Dot(out2),
	}

	var subU, subS, subV Matrix2
	mat2x2.SVD(&subU, &subS, &subV)

	subU1 := outBasis1.Scale(subU[0]).Add(outBasis2.Scale(subU[2]))
	subU2 := outBasis1.Scale(subU[1]).Add(outBasis2.Scale(subU[3]))
	subV1 := inBasis1.Scale(subV[0]).Add(inBasis2.Scale(subV[2]))
	subV2 := inBasis1.Scale(subV[1]).Add(inBasis2.Scale(subV[3]))

	*u = Matrix3{
		outVector[0], subU1[0], subU2[0],
		outVector[1], subU1[1], subU2[1],
		outVector[2], subU1[2], subU2[2],
	}
	*s = Matrix3{
		math.Sqrt(math.Max(0, real(eigVals[0]))), 0, 0,
		0, subS[0], 0,
		0, 0, subS[3],
	}
	*v = Matrix3{
		inVector[0], subV1[0], subV2[0],
		inVector[1], subV1[1], subV2[1],
		inVector[2], subV1[2], subV2[2],
	}

	// s might not be sorted due to rounding errors, but in
	// those cases the values really should be equal.
	if s[4] < s[8] {
		s[4] = s[8]
	}
	if s[0] < s[4] {
		s[0] = s[4]
	}
}

// symEigDecomp computes the eigendecomposition of the
// symmetric matrix.
//
// It populates matrices v and s such that
//
//	m = v*s*v.Transpose()
func (m *Matrix3) symEigDecomp(s, v *Matrix3) {
	eigVals := m.Eigenvalues()
	largestEig := math.Max(real(eigVals[0]), math.Max(real(eigVals[1]), real(eigVals[2])))
	inVector := m.symEigVector(largestEig)

	// Find other two singular components using a 2x2
	// matrix in a different basis.
	inBasis1, inBasis2 := inVector.OrthoBasis()
	out1 := m.MulColumn(inBasis1)
	out2 := m.MulColumn(inBasis2)

	// Not gonna lie: I didn't know this order would be right.
	mat2x2 := Matrix2{
		out1.Dot(inBasis1), out2.Dot(inBasis1),
		out1.Dot(inBasis2), out2.Dot(inBasis2),
	}

	var subV, subS Matrix2
	mat2x2.symEigDecomp(&subS, &subV)

	// Both of these work and I honestly have no idea why.
	//
	// subV1 := inBasis1.Scale(subV[0]).Add(inBasis2.Scale(subV[2]))
	// subV2 := inBasis1.Scale(subV[1]).Add(inBasis2.Scale(subV[3]))
	subV1 := inBasis1.Scale(subV[0]).Add(inBasis2.Scale(subV[1]))
	subV2 := inBasis1.Scale(subV[2]).Add(inBasis2.Scale(subV[3]))

	*v = *NewMatrix3Columns(inVector, subV1, subV2)
	*s = Matrix3{
		real(eigVals[0]), 0, 0,
		0, subS[0], 0,
		0, 0, subS[3],
	}
}

func (m *Matrix3) symEigVector(val float64) Vec3 {
	row1 := Vec3{m[0] - val, m[1], m[2]}
	row2 := Vec3{m[3], m[4] - val, m[5]}
	row3 := Vec3{m[6], m[7], m[8] - val}

	// Search for the null-space by trying a bunch of
	// different possibilities and choosing the most
	// null-spacey one of them.
	var bestVector Vec3
	var bestResult float64
	var triedAny bool
	tryVector := func(c Vec3) {
		norm := c.Norm()
		if norm == 0 {
			return
		}
		v := c.Scale(1 / norm)
		out := math.Max(math.Max(math.Abs(row1.Dot(v)), math.Abs(row2.Dot(v))),
			math.Abs(row3.Dot(v)))
		if !triedAny || out < bestResult {
			bestVector = v
			bestResult = out
			triedAny = true
		}
	}
	tryOrtho := func(c Vec3) {
		if (c == Vec3{}) {
			return
		}
		v1, _ := c.OrthoBasis()
		tryVector(v1)
	}

	// Rank 1 matrices will have a null-space as any
	// vector orthogonal to a non-zero row.
	tryOrtho(row1)
	tryOrtho(row2)
	tryOrtho(row3)

	// Rank 2 matrices will have a null-space as the
	// cross-product of some two rows.
	tryVector(row1.Cross(row2))
	tryVector(row1.Cross(row3))
	tryVector(row2.Cross(row3))

	if !triedAny {
		// It's a rank-zero matrix.
		return Vec3{1, 0, 0}
	}
	return bestVector
}
