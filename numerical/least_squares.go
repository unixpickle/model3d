package numerical

// LeastSquares3 computes the least squares solution to the
// equation Ax = b, where A is a matrix with rows vs, and b
// is a column matrix.
//
// The epsilon argument is a lower bound for singular values
// in the psuedoinverse.
func LeastSquares3(a []Vec3, b []float64, epsilon float64) Vec3 {
	return LeastSquaresReg3(a, b, 0, epsilon)
}

// LeastSquaresReg3 is like LeastSquares3, but uses ridge
// regression with a penalty equal to lambda.
func LeastSquaresReg3(a []Vec3, b []float64, lambda, epsilon float64) Vec3 {
	// A^T * A = A^T * b
	var leftSide Matrix3
	var rightSide Vec3
	for i, v := range a {
		rightSide = rightSide.Add(v.Scale(b[i]))
		outIdx := 0
		for j := 0; j < 3; j++ {
			for i := 0; i < 3; i++ {
				leftSide[outIdx] += v[i] * v[j]
				outIdx++
			}
		}
	}

	leftSide[0] += lambda
	leftSide[4] += lambda
	leftSide[8] += lambda

	var s, v Matrix3
	leftSide.symEigDecomp(&s, &v)

	// V*S*V^T * x = rightSide
	// x = V*inv(S)*V^T * rightSide
	for i := 0; i < 3; i++ {
		if s[i*4] > epsilon {
			s[i*4] = 1 / s[i*4]
		} else {
			s[i*4] = 0
		}
	}
	return v.Mul(&s).Mul(v.Transpose()).MulColumn(rightSide)
}
