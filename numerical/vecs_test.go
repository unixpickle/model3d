package numerical

import (
	"math"
	"testing"
)

func TestVec4OrthoBasis(t *testing.T) {
	checkVec := func(t *testing.T, v Vec4) {
		v1, v2, v3 := v.OrthoBasis()
		mat := NewMatrix4Columns(v.Normalize(), v1, v2, v3)
		checkOrthogonal4(t, mat)
		if det := mat.Det(); math.IsInf(det, 0) || math.IsNaN(det) || math.Abs(det-1.0) > 1e-5 {
			t.Errorf("expected determinant 1.0 but got %f", det)
		}
	}
	t.Run("AxisAligned", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			for _, value := range []float64{-7.0, 7.0, -1.0, 1.0} {
				var v Vec4
				v[i] = value
				checkVec(t, v)
			}
		}
	})
	t.Run("Random", func(t *testing.T) {
		v := Vec4{10, -17, 32, 5}
		checkVec(t, v)
		for i := 0; i < 100; i++ {
			checkVec(t, NewVec4RandomUnit())
		}
	})
}

func checkOrthogonal4(t *testing.T, mat *Matrix4) {
	mtm := mat.Transpose().Mul(mat)
	eye := NewMatrix4Identity()
	for i, x := range eye {
		a := mtm[i]
		if math.Abs(a-x) > 1e-8 {
			t.Errorf("matrix %v is not orthogonal (A^T * A = %v)", *mat, *mtm)
			return
		}
	}
}
