package numerical

// A QEF4 is a four-dimensional quadric error function.
//
// The storage is in row-major order, skipping below-diagonal entries.
type QEF4 [10]float64

// NewQEF4Outer creates a QEF4 as an outer product of a 4-D vector.
func NewQEF4Outer(v Vec4) *QEF4 {
	return &QEF4{
		v[0] * v[0], v[0] * v[1], v[0] * v[2], v[0] * v[3],
		/*v[1] * v[0],*/ v[1] * v[1], v[1] * v[2], v[1] * v[3],
		/*v[2] * v[0], v[2]*v[1],*/ v[2] * v[2], v[2] * v[3],
		/*v[3] * v[0], v[3]*v[1], v[3]*v[2],*/ v[3] * v[3],
	}
}

// NewQEF4Dist creates a QEF4 that implements the function
//
//	sum_i (c[i] - v[i])^2
//
// For the first three dimensions of v, assuming the fourth
// dimension of v is 1.
func NewQEF4Dist(c Vec3) *QEF4 {
	return &QEF4{
		1, 0, 0, -c[0],
		/*0,*/ 1, 0, -c[1],
		/*0, 0,*/ 1, -c[2],
		/*-c[0], -c[1], -c[2],*/ c.Dot(c),
	}
}

// Matrix4 returns a raw Matrix4 for q.
func (q *QEF4) Matrix4() *Matrix4 {
	return &Matrix4{
		q[0], q[1], q[2], q[3],
		q[1], q[4], q[5], q[6],
		q[2], q[5], q[7], q[8],
		q[3], q[6], q[8], q[9],
	}
}

func (q *QEF4) topLeft() *Matrix3 {
	return &Matrix3{
		q[0], q[1], q[2],
		q[1], q[4], q[5],
		q[2], q[5], q[7],
	}
}

// Eval evaluates the error of a 3D vector by filling the
// last dimension with 1 and evaluating v^T * Q * v.
func (q *QEF4) Eval(v Vec3) float64 {
	m4 := q.Matrix4()
	dp := m4.MulColumn(Vec4{v[0], v[1], v[2], 1})
	return dp[0]*v[0] + dp[1]*v[1] + dp[2]*v[2] + dp[3]
}

// Minimize solves for the value of v that minimizes
//
//	q.Eval(v)
//
// Also returns the determinant of the linear system, which
// is near-zero if the solution is unlikely to be stable.
func (q *QEF4) Minimize() (Vec3, float64) {
	m := q.topLeft()
	v := Vec3{-q[3], -q[6], -q[8]}
	det := m.Det()
	return m.MulColumnInv(v, det), det
}

// Add returns the sum of q and q1.
func (q *QEF4) Add(q1 *QEF4) *QEF4 {
	res := *q
	for i, x := range q1 {
		res[i] += x
	}
	return &res
}

// Add returns the sum of q and q1.
func (q *QEF4) Scale(s float64) *QEF4 {
	res := *q
	for i, x := range q {
		res[i] = x * s
	}
	return &res
}
