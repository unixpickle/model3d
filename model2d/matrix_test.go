package model2d

import (
	"math"
	"math/cmplx"
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
