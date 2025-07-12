package numerical

import (
	"math"
	"testing"
)

func TestMatrix4Mul(t *testing.T) {
	m1 := &Matrix4{-2.1130000, 1.4820000, 0.0370000, 0.3030000, 1.4960000, 0.3140000, 1.0620000, -0.8200000, -0.6650000, 0.5030000, -0.6730000, -0.7730000, 0.6070000, -0.4350000, 0.7850000, 2.0240000}
	m2 := &Matrix4{-0.4530000, 0.1840000, -1.0770000, 0.1830000, -0.3520000, 1.9300000, 0.4620000, 0.3640000, -1.0870000, -0.1670000, -0.5330000, -0.8320000, -1.3760000, 1.1500000, 2.0760000, 0.4800000}
	expected := &Matrix4{-0.0216220, 2.8137390, 3.5696920, 0.2674250, -0.8142900, -0.2390700, -3.7344900, -0.8891200, 1.9193880, 0.0718710, -0.2974480, 0.2502930, -3.7601700, 1.4686430, 2.9287100, 0.2711410}
	actual := m1.Mul(m2)

	for i, x := range expected {
		a := actual[i]
		if math.Abs(a-x) > 1e-7 {
			t.Errorf("entry %d: expected %f but got %f", i, x, a)
		}
	}
}

func TestMatrix4CharPoly(t *testing.T) {
	check := func(t *testing.T, actual, expected Polynomial) {
		if len(actual) != len(expected) {
			t.Fatalf("expected len %d but got %d", len(expected), len(actual))
		}
		for i, x := range expected {
			a := actual[i]
			if math.Abs(a-x) > 1e-3 {
				t.Errorf("term %d should be %f but got %f", i, x, a)
			}
		}
	}

	t.Run("NoThirdDegree", func(t *testing.T) {
		mat := &Matrix4{
			1.0, 3.0, 2.0, -7.0,
			5.0, -6.0, 3.0, -2.0,
			-4.0, 3.0, 2.0, 3.0,
			9.0, 1.0, 2.0, 3.0,
		}
		actual := mat.CharPoly()
		expected := Polynomial{-2420.0, 399.0, 18.0, 0.0, 1.0}
		check(t, actual, expected)
	})

	t.Run("Full", func(t *testing.T) {
		mat := &Matrix4{
			1.0, 3.0, 2.0, -7.0,
			5.0, -6.0, 3.0, -2.0,
			-4.0, 3.0, 2.0, 3.0,
			9.0, 1.0, 2.0, 4.0,
		}
		actual := mat.CharPoly()
		expected := Polynomial{-2525.0, 431.0, 15.0, -1.0, 1.0}
		check(t, actual, expected)
	})
}

func TestMatrix4SVD(t *testing.T) {
	testSVD := func(t *testing.T, prec float64, mat *Matrix4) {
		var u, s, v Matrix4
		mat.SVD(&u, &s, &v)
		checkOrthogonal4(t, &u)
		checkOrthogonal4(t, &v)

		product := u.Mul(&s).Mul(v.Transpose())
		for i, x := range mat {
			a := product[i]
			if math.Abs(a-x) > prec {
				t.Errorf("expected matrix %v but got U*S*V^T = %v (diff=%v)", *mat, *product, product.Sub(mat))
				break
			}
		}
	}
	t.Run("Basic", func(t *testing.T) {
		testSVD(t, 1e-8, &Matrix4{
			1.0, 3.0, 2.0, -7.0,
			5.0, -6.0, 3.0, -2.0,
			-4.0, 3.0, 2.0, 3.0,
			9.0, 1.0, 2.0, 4.0,
		})
	})
	t.Run("Zeros", func(t *testing.T) {
		testSVD(t, 1e-8, &Matrix4{})
	})
	t.Run("SingleSV", func(t *testing.T) {
		// All singular values are 2.
		// The precision we use is lower than normal, because the characteristic
		// polynomial is essentially (x-4)^4, so finding the root is very poorly
		// conditioned.
		testSVD(t, 2e-4, &Matrix4{
			0.21466134858214758, -1.5439761178918934, 1.1357222139082424, 0.5293328873589178,
			0.18422399983270635, 0.7634112609593281, 0.14852742239738262, 1.833358767214458,
			0.4830940705809065, -0.9578566217750428, -1.6178405574598163, 0.48137588403386594,
			-1.9200526404971558, -0.3403699254365289, -0.2658318797086999, 0.35620160486551505,
		})
	})
}

func BenchmarkMatrix4SVD(b *testing.B) {
	mat := &Matrix4{
		1.0, 3.0, 2.0, -7.0,
		5.0, -6.0, 3.0, -2.0,
		-4.0, 3.0, 2.0, 3.0,
		9.0, 1.0, 2.0, 4.0,
	}
	for i := 0; i < b.N; i++ {
		var u, s, v Matrix4
		mat.SVD(&u, &s, &v)
	}
}
