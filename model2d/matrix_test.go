package model2d

import (
	"math"
	"math/cmplx"
	"math/rand"
	"testing"
)

func TestEigenvalues(t *testing.T) {
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

func TestSVD(t *testing.T) {
	for i := 0; i < 100; i++ {
		mat := &Matrix2{}
		for j := range mat {
			mat[j] = rand.NormFloat64()
		}
		var u, s, v Matrix2
		mat.SVD(&u, &s, &v)

		eye := &Matrix2{1, 0, 0, 1}
		if !matrixClose(u.Transpose().Mul(&u), eye) {
			t.Errorf("u is not orthonormal: %v", u)
		}
		if !matrixClose(v.Transpose().Mul(&v), eye) {
			t.Errorf("v is not orthonormal: %v", v)
		}

		recon := u.Mul(s.Mul(v.Transpose()))
		for j, x := range mat {
			a := recon[j]
			if math.Abs(x-a) > 1e-5 {
				t.Errorf("got %v but expected %v", recon, mat)
				break
			}
		}
	}
}

func BenchmarkSVD(b *testing.B) {
	matrix := &Matrix2{1, 2, 3, 4}
	for i := 0; i < b.N; i++ {
		var u, s, v Matrix2
		matrix.SVD(&u, &s, &v)
	}
}

func matrixClose(m1, m2 *Matrix2) bool {
	for i, x := range m1 {
		y := m2[i]
		if math.Abs(x-y) > 1e-5 {
			return false
		}
	}
	return true
}
