package numerical

type Matrix4 [16]float64

// Add computes m+m1 and returns the sum.
func (m *Matrix4) Add(m1 *Matrix4) *Matrix4 {
	var res Matrix4
	for i, x := range m {
		res[i] = x + m1[i]
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

// SingularValues computes the eigenvalues of M^T * M.
func (m *Matrix4) SingularValues() [4]float64 {
	panic("unimplemented")
}
