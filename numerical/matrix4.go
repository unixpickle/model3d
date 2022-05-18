package numerical

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

// Rows returns the rows of the matrix as Vec4.
func (m *Matrix4) Rows() [4]Vec4 {
	return [4]Vec4{
		{m[0], m[1], m[2], m[3]},
		{m[4], m[5], m[6], m[7]},
		{m[8], m[9], m[10], m[11]},
		{m[12], m[13], m[14], m[15]},
	}
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
