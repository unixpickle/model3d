package numerical

import (
	"math"
	"math/cmplx"
	"math/rand"
	"testing"
)

func TestMatrix2Eigenvalues(t *testing.T) {
	mats := []*Matrix2{
		{1, 2, 3, 4},
		{1, 0, 0, 1},
		{1, 0, 0, 2},
		{0, 1, 1, 0},
		{math.Sqrt2, math.Sqrt2, -math.Sqrt2, math.Sqrt2},
	}
	eigs := [][]complex128{
		{-.37228132326901432992, 5.37228132326901432992},
		{1, 1},
		{1, 2},
		{-1, 1},
		{complex(math.Sqrt2, math.Sqrt2), complex(math.Sqrt2, -math.Sqrt2)},
	}
	for i, mat := range mats {
		expected := eigs[i]
		actual := mat.Eigenvalues()
		for j, x := range expected {
			a := actual[j]
			if cmplx.Abs(x-a) > 1e-8 {
				t.Errorf("case %d eig %d: should be %f but got %f", i, j, x, a)
			}
		}
	}
}

func TestMatrix2SVD(t *testing.T) {
	testDecomp := func(t *testing.T, mat *Matrix2) {
		var u, s, v Matrix2
		mat.SVD(&u, &s, &v)

		if s[0] < s[3] {
			t.Errorf("singular values not sorted: %f,%f", s[0], s[3])
		}

		eye := &Matrix2{1, 0, 0, 1}
		if !matrixClose2(u.Transpose().Mul(&u), eye) {
			t.Errorf("u is not orthonormal: %v", u)
		}
		if !matrixClose2(v.Transpose().Mul(&v), eye) {
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
			mat := &Matrix2{}
			for j := range mat {
				mat[j] = rand.NormFloat64()
			}
			testDecomp(t, mat)
		}
	})

	t.Run("Ortho", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			c1 := NewVec2RandomUnit()
			c2 := NewVec2RandomUnit().ProjectOut(c1).Normalize()
			testDecomp(t, NewMatrix2Columns(c1, c2))
		}
	})
}

func BenchmarkMatrix2SVD(b *testing.B) {
	matrix := &Matrix2{1, 2, 3, 4}
	for i := 0; i < b.N; i++ {
		var u, s, v Matrix2
		matrix.SVD(&u, &s, &v)
	}
}

func matrixClose2(m1, m2 *Matrix2) bool {
	for i, x := range m1 {
		y := m2[i]
		if math.Abs(x-y) > 1e-5 {
			return false
		}
	}
	return true
}
