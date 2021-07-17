package numerical

import (
	"math"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestPolynomialDivideRoot(t *testing.T) {
	p1 := Polynomial{-10, -3, 1}
	p2 := p1.divideRoot(-2)
	if len(p2) != 2 || math.Abs(p2[0]+5) > 1e-5 || math.Abs(p2[1]-1) > 1e-5 {
		t.Errorf("unexpected quotient: %v", p2)
	} else if math.IsNaN(p2[0]) || math.IsNaN(p2[1]) || math.IsInf(p2[0], 0) || math.IsInf(p2[1], 0) {
		t.Errorf("invalid numeric in result: %v", p2)
	}

	// (2x^2+3x+4)(x-5) = 2x^3-7x^2-11x-20
	p1 = Polynomial{-20, -11, -7, 2}
	actual := p1.divideRoot(5)
	expected := Polynomial{4, 3, 2}
	if len(actual) != len(expected) {
		t.Errorf("bad length in result: %v", actual)
	} else {
		for i, x := range expected {
			a := actual[i]
			if math.IsNaN(a) || math.IsInf(a, 0) {
				t.Errorf("invalid numeric in result: %v", actual)
				break
			} else if math.Abs(a-x) > 1e-5 {
				t.Errorf("unexpected quotient: %v", actual)
			}
		}
	}
}

func TestPolynomialSearchRoot(t *testing.T) {
	// (2x^2+3x+4)(x-5) = 2x^3-7x^2-11x-20
	p := Polynomial{-20, -11, -7, 2}
	root := p.searchRoot(4.9, 5.1)
	if math.IsNaN(root) || math.IsInf(root, 0) || math.Abs(root-5) > 1e-5 {
		t.Errorf("unexpected root: %v", root)
	}
}

func TestPolynomialRealRoots(t *testing.T) {
	checkRoots := func(t *testing.T, actual []float64, expected []float64) {
		if len(actual) != len(expected) {
			t.Fatalf("bad roots: %v (expected %v)", actual, expected)
		}
		leftover := append([]float64{}, expected...)
		for _, a := range actual {
			if math.IsNaN(a) || math.IsInf(a, 0) {
				t.Fatalf("bad roots: %v (expected %v)", actual, expected)
			}
			found := false
			for i, x := range leftover {
				if math.Abs(a-x) < 1e-5 {
					found = true
					essentials.UnorderedDelete(&leftover, i)
					break
				}
			}
			if !found {
				t.Fatalf("bad roots: %v (expected %v)", actual, expected)
			}
		}
	}

	t.Run("Quad", func(t *testing.T) {
		// 2(x-2)(x-3) = 2x^2-10x+12
		p := Polynomial{12, -10, 2}
		roots := p.RealRoots()
		checkRoots(t, roots, []float64{2, 3})
	})

	t.Run("CubicSingle", func(t *testing.T) {
		// 2x^3-7x^2-11x-20 has only one real root, 5.
		p := Polynomial{-20, -11, -7, 2}
		roots := p.RealRoots()
		checkRoots(t, roots, []float64{5})
	})

	t.Run("Quartic", func(t *testing.T) {
		// (x+2)(x-3)(x^2+2x+7) = x^4+x^3-x^2-19x-42
		p := Polynomial{-42, -19, -1, 1, 1}
		roots := p.RealRoots()
		checkRoots(t, roots, []float64{-2, 3})
	})

	t.Run("Large", func(t *testing.T) {
		// (x+5)(x-7)(x+3)(x+2)(x-3)(x^2+2x+7)^2
		p := Polynomial{30870, 34839, 17738, 4526, -570, -744, -234, -30, 4, 1}
		roots := p.RealRoots()
		checkRoots(t, roots, []float64{-5, 7, -3, -2, 3})
	})

	t.Run("LargeRepeated", func(t *testing.T) {
		// (x+5)^2(x-7)(x+3)(x+2)(x-3)(x^2+2x+7)^2
		p := Polynomial{154350, 205065, 123529, 40368, 1676, -4290, -1914, -384, -10, 9, 1}
		roots := p.RealRoots()
		checkRoots(t, roots, []float64{-5, -5, 7, -3, -2, 3})
	})
}
