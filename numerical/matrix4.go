package numerical

import (
	"math"
)

type Matrix4 [16]float64

// NewMatrix4Identity creates an identity matrix.
func NewMatrix4Identity() *Matrix4 {
	return &Matrix4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// NewMatrix4Columns creates a Matrix4 from the columns.
func NewMatrix4Columns(c1, c2, c3, c4 Vec4) *Matrix4 {
	var res Matrix4
	for i := 0; i < 4; i++ {
		offset := i * 4
		res[offset] = c1[i]
		res[offset+1] = c2[i]
		res[offset+2] = c3[i]
		res[offset+3] = c4[i]
	}
	return &res
}

// Add computes m+m1 and returns the sum.
func (m *Matrix4) Add(m1 *Matrix4) *Matrix4 {
	var res Matrix4
	for i, x := range m {
		res[i] = x + m1[i]
	}
	return &res
}

// Sub computes m-m1 and returns the result.
func (m *Matrix4) Sub(m1 *Matrix4) *Matrix4 {
	var res Matrix4
	for i, x := range m {
		res[i] = x - m1[i]
	}
	return &res
}

// Scale returns the scalar-matrix product.
func (m *Matrix4) Scale(s float64) *Matrix4 {
	var res Matrix4
	for i, x := range m {
		res[i] = x * s
	}
	return &res
}

// Mul computes m*m1 and returns the product.
func (m *Matrix4) Mul(m1 *Matrix4) *Matrix4 {
	return &Matrix4{
		m[0]*m1[0] + m[1]*m1[4] + m[2]*m1[8] + m[3]*m1[12],
		m[0]*m1[1] + m[1]*m1[5] + m[2]*m1[9] + m[3]*m1[13],
		m[0]*m1[2] + m[1]*m1[6] + m[2]*m1[10] + m[3]*m1[14],
		m[0]*m1[3] + m[1]*m1[7] + m[2]*m1[11] + m[3]*m1[15],

		m[4]*m1[0] + m[5]*m1[4] + m[6]*m1[8] + m[7]*m1[12],
		m[4]*m1[1] + m[5]*m1[5] + m[6]*m1[9] + m[7]*m1[13],
		m[4]*m1[2] + m[5]*m1[6] + m[6]*m1[10] + m[7]*m1[14],
		m[4]*m1[3] + m[5]*m1[7] + m[6]*m1[11] + m[7]*m1[15],

		m[8]*m1[0] + m[9]*m1[4] + m[10]*m1[8] + m[11]*m1[12],
		m[8]*m1[1] + m[9]*m1[5] + m[10]*m1[9] + m[11]*m1[13],
		m[8]*m1[2] + m[9]*m1[6] + m[10]*m1[10] + m[11]*m1[14],
		m[8]*m1[3] + m[9]*m1[7] + m[10]*m1[11] + m[11]*m1[15],

		m[12]*m1[0] + m[13]*m1[4] + m[14]*m1[8] + m[15]*m1[12],
		m[12]*m1[1] + m[13]*m1[5] + m[14]*m1[9] + m[15]*m1[13],
		m[12]*m1[2] + m[13]*m1[6] + m[14]*m1[10] + m[15]*m1[14],
		m[12]*m1[3] + m[13]*m1[7] + m[14]*m1[11] + m[15]*m1[15],
	}
}

// MulColumn computes a matrix-vector product.
func (m *Matrix4) MulColumn(v Vec4) Vec4 {
	var res Vec4
	for i := 0; i < 4; i++ {
		rowOffset := i * 4
		for j := 0; j < 4; j++ {
			res[i] += m[rowOffset+j] * v[j]
		}
	}
	return res
}

// Transpose computes the matrix transpose.
func (m *Matrix4) Transpose() *Matrix4 {
	var res Matrix4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			res[i*4+j] = m[i+j*4]
		}
	}
	return &res
}

// Cols returns the columns of the matrix as Vec4.
func (m *Matrix4) Cols() [4]Vec4 {
	return m.Transpose().Rows()
}

// Rows returns the rows of the matrix as Vec4.
func (m *Matrix4) Rows() [4]Vec4 {
	return [4]Vec4{
		{m[0], m[1], m[2], m[3]},
		{m[4], m[5], m[6], m[7]},
		{m[8], m[9], m[10], m[11]},
		{m[12], m[13], m[14], m[15]},
	}
}

// Det returns the determinant of m.
func (m *Matrix4) Det() float64 {
	a, b, c, d, e, f, g, h, i, j, k, l, m1, n, o, p := m[0], m[1], m[2], m[3], m[4], m[5], m[6],
		m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15]
	return a*f*k*p - a*f*l*o - a*g*j*p + a*g*l*n + a*h*j*o - a*h*k*n - b*e*k*p + b*e*l*o + b*g*i*p -
		b*g*l*m1 - b*h*i*o + b*h*k*m1 + c*e*j*p - c*e*l*n - c*f*i*p + c*f*l*m1 + c*h*i*n - c*h*j*m1 -
		d*e*j*o + d*e*k*n + d*f*i*o - d*f*k*m1 - d*g*i*n + d*g*j*m1
}

// CharPoly computes the characteristic polynomial of the
// matrix.
func (m *Matrix4) CharPoly() Polynomial {
	a, b, c, d, e, f, g, h, i, j, k, l, m1, n, o, p := m[0], m[1], m[2], m[3], m[4], m[5], m[6],
		m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15]
	return Polynomial{
		a*f*k*p - a*f*l*o - a*g*j*p + a*g*l*n + a*h*j*o - a*h*k*n - b*e*k*p + b*e*l*o + b*g*i*p -
			b*g*l*m1 - b*h*i*o + b*h*k*m1 + c*e*j*p - c*e*l*n - c*f*i*p + c*f*l*m1 + c*h*i*n - c*h*j*m1 -
			d*e*j*o + d*e*k*n + d*f*i*o - d*f*k*m1 - d*g*i*n + d*g*j*m1,
		-a*f*k - a*f*p + a*g*j + a*h*n - a*k*p + a*l*o + b*e*k + b*e*p - b*g*i - b*h*m1 - c*e*j + c*f*i + c*i*p - c*l*m1 - d*e*n + d*f*m1 - d*i*o + d*k*m1 - f*k*p + f*l*o + g*j*p - g*l*n - h*j*o + h*k*n,
		a*f + a*k + a*p - b*e - c*i - d*m1 + f*k + f*p - g*j - h*n + k*p - l*o,
		-a - f - k - p,
		1,
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
func (m *Matrix4) SVD(u, s, v *Matrix4) {
	mtm := m.Transpose().Mul(m)

	// Reduce matrix to be block diagonal, with a 3x3 in the top-left
	// and a 1x1 in the bottom right.
	root := mtm.findSymEig()
	nullMatrix := mtm.Add(NewMatrix4Identity().Scale(-root))
	rightMatrix := nullMatrix.basisAndNullVec()
	leftVector := m.MulColumn(rightMatrix.Cols()[3]).Normalize()
	leftB1, leftB2, leftB3 := leftVector.OrthoBasis()
	leftMatrix := NewMatrix4Columns(leftB1, leftB2, leftB3, leftVector)
	reduced := leftMatrix.Transpose().Mul(m.Mul(rightMatrix))

	// Compute SVD of sub-block of 3x3 matrix in remaining orthogonal basis.
	reduced3x3 := Matrix3{
		reduced[0], reduced[1], reduced[2],
		reduced[4], reduced[5], reduced[6],
		reduced[8], reduced[9], reduced[10],
	}
	var subU, subS, subV Matrix3
	reduced3x3.SVD(&subU, &subS, &subV)

	*s = Matrix4{
		subS[0], 0, 0, 0,
		0, subS[4], 0, 0,
		0, 0, subS[8], 0,
		0, 0, 0, math.Sqrt(math.Max(0, root)),
	}

	// Convert U and V back into the full basis.
	subU4 := Matrix4{
		subU[0], subU[1], subU[2], 0,
		subU[3], subU[4], subU[5], 0,
		subU[6], subU[7], subU[8], 0,
		0, 0, 0, 1,
	}
	subV4 := Matrix4{
		subV[0], subV[1], subV[2], 0,
		subV[3], subV[4], subV[5], 0,
		subV[6], subV[7], subV[8], 0,
		0, 0, 0, 1,
	}

	// leftMatrix^T * m * rightMatrix = subU4 * sigma * subV4^T
	// m = leftMatrix * subU4 * sigma * subV4^T * rightMatrix^T
	//   = leftMatrix * subU4 * sigma * (rightMatrix*subV4)^T
	*u = *(leftMatrix.Mul(&subU4))
	*v = *(rightMatrix.Mul(&subV4))
}

// findSymEig finds one real eigenvalue of m, assuming m is
// symmetric and therefore has one.
func (m *Matrix4) findSymEig() float64 {
	poly := m.CharPoly()

	for len(poly) >= 2 {
		roots := poly.RealRoots()
		if len(roots) > 0 {
			return roots[0]
		}
		// The polynomial must have roots we aren't seeing because they
		// just barely brush the x axis. This means that the roots are
		// probably repeated, and are therefore a local minimum.
		poly = poly.Derivative()
	}

	// A root should have been found, and this case
	// should not be reached.
	return 0.0
}

// basisAndNullVec returns a matrix with four orthonormal
// columns, such that the final column is the closest to
// the right null-space of the matrix.
func (m *Matrix4) basisAndNullVec() *Matrix4 {
	vecs := m.Rows()

	// Start with a normalized set of vectors so that we can
	// assert later that small values (e.g. 1e-10) really do
	// correspond to the null-space.
	for i, v := range vecs {
		norm := v.Norm()
		if norm > 0 {
			vecs[i] = v.Scale(1 / norm)
		}
	}

	// Apply Gram-Schmidt, but re-order the vectors to
	// always choose the next vector such that it has the
	// largest magnitude in a direction not spanned by the
	// previous vectors.
	for i := 0; i < 4; i++ {
		maxNorm := -1.0
		maxIdx := 0
		for j := i; j < 4; j++ {
			norm := vecs[j].Norm()
			if norm >= maxNorm {
				maxIdx = j
				maxNorm = norm
			}
		}
		vecs[i], vecs[maxIdx] = vecs[maxIdx], vecs[i]
		if maxNorm < 1e-10 {
			// Create a random orthogonal vector since this is
			// a null-space vector. Without doing this, the
			// direction we get will depend significantly on
			// rounding error, and may not be orthogonal at all.
			vecs[i] = NewVec4RandomNormal()
			for j := 0; j < i; j++ {
				vecs[i] = vecs[i].Sub(vecs[j].Scale(vecs[i].Dot(vecs[j])))
			}
			maxNorm = vecs[i].Norm()
		}
		vecs[i] = vecs[i].Scale(1 / maxNorm)

		// Project vecs[i] out of the remaining vectors.
		for j := i + 1; j < 4; j++ {
			vecs[j] = vecs[j].Sub(vecs[i].Scale(vecs[i].Dot(vecs[j])))
		}
	}
	return NewMatrix4Columns(vecs[0], vecs[1], vecs[2], vecs[3])
}
