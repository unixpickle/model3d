package numerical

import (
	"math"
	"math/cmplx"
	"math/rand"
	"testing"
)

func TestMatrix3Inverse(t *testing.T) {
	for i := 0; i < 10; i++ {
		m := Matrix3{}
		for j := range m {
			m[j] = rand.NormFloat64()
		}
		m1 := m.Inverse()
		product := m.Mul(m1)
		product[0] -= 1
		product[4] -= 1
		product[8] -= 1
		for j, x := range product {
			if math.Abs(x) > 1e-8 {
				t.Errorf("entry %d should be 0 but got %f", j, x)
			}
		}
	}
}

func TestMatrix3Rotation(t *testing.T) {
	t.Run("AroundY", func(t *testing.T) {
		theta := 0.1
		mat := NewMatrix3Rotation(Vec3{0, 1, 0}, theta)
		c1 := Vec3{2, 1, 3}
		newTheta := math.Atan2(c1[2], c1[0]) - theta
		norm := Vec2{c1[0], c1[2]}.Norm()
		expected := Vec3{
			norm * math.Cos(newTheta),
			c1[1],
			norm * math.Sin(newTheta),
		}
		actual := mat.MulColumn(c1)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	})

	t.Run("AroundZ", func(t *testing.T) {
		theta := 0.1
		mat := NewMatrix3Rotation(Vec3{0, 0, 1}, theta)
		c1 := Vec3{2, 1, 3}
		newTheta := math.Atan2(c1[1], c1[0]) + theta
		norm := Vec2{c1[0], c1[1]}.Norm()
		expected := Vec3{
			norm * math.Cos(newTheta),
			norm * math.Sin(newTheta),
			3,
		}
		actual := mat.MulColumn(c1)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	})

	t.Run("Flip", func(t *testing.T) {
		axis := NewVec3RandomUnit()
		c := NewVec3RandomNormal()
		c1 := NewMatrix3Rotation(axis, 0.1).MulColumn(c)
		c2 := NewMatrix3Rotation(axis.Scale(-1), -0.1).MulColumn(c)
		if c1.Dist(c2) > 1e-5 {
			t.Error("negating axis negates rotation direction")
		}
	})
}

func TestMatrix3Eigenvalues(t *testing.T) {
	mats := []*Matrix3{
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{0, 1, 0, 0, 0, 1, 1, 0, 0},
	}
	eigs := [][3]complex128{
		{0, complex(0.5*(3*math.Sqrt(33)+15), 0), complex(0.5*(-3*math.Sqrt(33)+15), 0)},
		{1, complex(-1, math.Sqrt(3)) / 2, complex(-1, -math.Sqrt(3)) / 2},
	}
	for i, mat := range mats {
		expected := eigs[i]
		actual := map[complex128]int{}
		for _, x := range mat.Eigenvalues() {
			actual[x]++
		}
		for j, x := range expected {
			var notFirst bool
			var a complex128
			for anA, count := range actual {
				if count == 0 {
					continue
				}
				if !notFirst || cmplx.Abs(anA-x) < cmplx.Abs(a-x) {
					a = anA
					notFirst = true
				}
			}
			actual[a]--
			if cmplx.Abs(x-a) > 1e-8 {
				t.Errorf("case %d eig %d: should be %f but got %f", i, j, x, a)
			}
		}
	}
}

func TestMatrix3SVD(t *testing.T) {
	ensureEquivalence := func(t *testing.T, mat *Matrix3) {
		var u, s, v Matrix3
		mat.SVD(&u, &s, &v)

		if s[0] < s[4] || s[4] < s[8] {
			t.Errorf("singular values not sorted: %f,%f,%f", s[0], s[4], s[8])
		}

		eye := &Matrix3{1, 0, 0, 0, 1, 0, 0, 0, 1}
		if !matrixClose3(u.Transpose().Mul(&u), eye) {
			t.Errorf("u is not orthonormal: %v", u)
		}
		if !matrixClose3(v.Transpose().Mul(&v), eye) {
			t.Errorf("v is not orthonormal: %v", v)
		}

		recon := u.Mul(s.Mul(v.Transpose()))
		for j, x := range mat {
			a := recon[j]
			if math.Abs(x-a) > 1e-5 || math.IsNaN(a) {
				t.Errorf("got %v but expected %v", recon, mat)
				break
			}
		}
	}

	t.Run("Random", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			mat := &Matrix3{}
			for j := range mat {
				mat[j] = rand.NormFloat64()
			}
			ensureEquivalence(t, mat)
		}
	})

	t.Run("RepeatedEig", func(t *testing.T) {
		ensureEquivalence(t, &Matrix3{1, 0, 0, 0, 1, 0, 0, 0, 1})
		ensureEquivalence(t, &Matrix3{1, 0, 0, 1, 0, 0, 1, 0, 0})
		ensureEquivalence(t, &Matrix3{0, 0, 0, 0, 0, 0, 0, 0, 0})
		ensureEquivalence(t, &Matrix3{0, 1, 0, 0, 1, 1, 1, 0, 0})
	})

	t.Run("Ortho", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			c1 := NewVec3RandomUnit()
			c2, c3 := c1.OrthoBasis()
			for _, c := range []*Vec3{&c1, &c2, &c3} {
				if rand.Intn(2) == 0 {
					*c = c.Scale(-1)
				}
			}
			ensureEquivalence(t, NewMatrix3Columns(c1, c2, c3))
		}
	})
}

func TestMatrix3SymEigDecomp(t *testing.T) {
	matrix := &Matrix3{18, 13, 9, 13, 14, 14, 9, 14, 17}
	expectedEigs := []float64{40.33886832, 8.58897527, 0.07215641}

	var s, v Matrix3
	matrix.symEigDecomp(&s, &v)

	for i, x := range expectedEigs {
		a := s[i*4]
		if math.Abs(a-x) > 1e-4 {
			t.Errorf("expected eigenvalue %d to be %f but got %f", i, x, a)
		}
	}

	testEquivalent := func(matrix *Matrix3) {
		var s, v Matrix3
		matrix.symEigDecomp(&s, &v)
		product := v.Mul(s.Mul(v.Transpose()))
		for i, x := range matrix {
			a := product[i]
			if math.Abs(a-x) > 1e-4 {
				t.Errorf("expected entry %d to be %f but got %f", i, x, a)
			}
		}
	}
	testEquivalent(matrix)

	// Randomly generated positive definite.
	testEquivalent(&Matrix3{
		3.52262922, 3.46358736, 2.85937811, 3.46358736, 4.83930523,
		2.95558449, 2.85937811, 2.95558449, 2.84897598,
	})
	testEquivalent(&Matrix3{
		3.25572161, 0.58485646, 0.55781699, 0.58485646, 2.93323824,
		0.23630776, 0.55781699, 0.23630776, 0.57796781,
	})
	testEquivalent(&Matrix3{
		10.08859077, 0.27557785, 6.08485816, 0.27557785, 2.14838461,
		-0.20111911, 6.08485816, -0.20111911, 3.73788934,
	})

	// Randomly generated negative eigenvalues.
	testEquivalent(&Matrix3{
		-4.60804473, -3.50402888, 0.56751734, -3.50402888, -1.77185477,
		-0.41722716, 0.56751734, -0.41722716, 0.84882731,
	})
	testEquivalent(&Matrix3{
		4.60804473, 3.50402888, -0.56751734, 3.50402888, 1.77185477,
		0.41722716, -0.56751734, 0.41722716, -0.84882731,
	})
}

func BenchmarkMatrix3SVD(b *testing.B) {
	mat := &Matrix3{1, 2, 3, 4, 5, 6, 7, 8, 10}
	for i := 0; i < b.N; i++ {
		var u, s, v Matrix3
		mat.SVD(&u, &s, &v)
	}
}

func matrixClose3(m1, m2 *Matrix3) bool {
	for i, x := range m1 {
		y := m2[i]
		if math.Abs(x-y) > 1e-5 {
			return false
		}
	}
	return true
}
