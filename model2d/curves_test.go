package model2d

import (
	"math"
	"math/rand"
	"testing"
)

func TestBezierEval(t *testing.T) {
	curves := []BezierCurve{
		// 4th order.
		{
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
		},
		// 5th order.
		{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
			Coord{X: 2, Y: 3},
		},
		// 7th order.
		{
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
			Coord{X: 2, Y: -5},
			Coord{X: 7, Y: -2},
			Coord{X: 8, Y: 2},
		},
		// 17th order.
		{
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
			Coord{X: 2, Y: -5},
			Coord{X: 7, Y: -2},
			Coord{X: 8, Y: 2},
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
			Coord{X: 2, Y: -5},
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
			Coord{X: 2, Y: -5},
		},
	}
	for i, c := range curves {
		for j := 0; j < 100; j++ {
			x := rand.Float64()
			v1 := evalSimpleRecursive(c, x)
			v2 := c.Eval(x)
			if v1.Dist(v2) > 1e-5 {
				t.Errorf("curve %d: time %f: expected %v but got %v", i, x, v1, v2)
			}
		}
	}
}

func evalSimpleRecursive(b BezierCurve, t float64) Coord {
	if len(b) < 2 {
		panic("need at least two points")
	}
	if len(b) == 2 {
		return b[0].Scale(1 - t).Add(b[1].Scale(t))
	}
	term1 := evalSimpleRecursive(b[:len(b)-1], t).Scale(1 - t)
	term2 := evalSimpleRecursive(b[1:], t).Scale(t)
	return term1.Add(term2)
}

func TestBezierInverseX(t *testing.T) {
	curves := []BezierCurve{
		{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
		},
		{
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
		},
	}

	for _, curve := range curves {
		testX := func(x float64) {
			tValue := curve.InverseX(x)
			if math.IsNaN(tValue) {
				t.Errorf("unexpected NaN for x: %f", x)
			} else if math.Abs(x-curve.Eval(tValue).X) > 1e-8 {
				t.Errorf("invalid inverse for x: %f", x)
			}
		}
		testX(1)
		testX(3)
		for i := 0; i < 10; i++ {
			testX(rand.Float64()*1.9 + 1.05)
		}
		if !math.IsNaN(curve.InverseX(0.5)) {
			t.Error("not NaN before bounds")
		}
		if !math.IsNaN(curve.InverseX(3.5)) {
			t.Error("not NaN after bounds")
		}
	}
}

func TestBezierSplit(t *testing.T) {
	c := BezierCurve{
		Coord{X: 1, Y: 3},
		Coord{X: 2, Y: 2},
		Coord{X: 2, Y: 3},
		Coord{X: 3, Y: -2},
	}
	for _, t0 := range []float64{0.01, 0.5, 0.99} {
		expected1 := func(x float64) Coord {
			return c.Eval(x * t0)
		}
		expected2 := func(x float64) Coord {
			return c.Eval(t0 + x*(1-t0))
		}
		actual1, actual2 := c.Split(t0)
		for x := 0.0; x < 1.0; x += 0.001 {
			e1, e2 := expected1(x), expected2(x)
			a1, a2 := actual1.Eval(x), actual2.Eval(x)
			for i, a := range []Coord{a1, a2} {
				e := []Coord{e1, e2}[i]
				if a.Dist(e) > 1e-6 {
					t.Errorf("disagreement at split %f (half %d): x=%f actual %f expected %f",
						t0, i, x, a, e)
				}
			}
		}
	}
}

func TestBezierPolynomials(t *testing.T) {
	curve := BezierCurve{
		Coord{X: 3, Y: 3},
		Coord{X: 2, Y: 2},
		Coord{X: 2, Y: 3},
		Coord{X: 1, Y: -2},
		Coord{X: 2, Y: -5},
		Coord{X: 7, Y: -2},
		Coord{X: 8, Y: 2},
	}
	ps := curve.Polynomials()
	for i := 0; i < 1000; i++ {
		x := rand.Float64()
		expected := curve.Eval(x)
		actual := XY(ps[0].Eval(x), ps[1].Eval(x))
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("at time %f point should be %v but got %v", x, expected, actual)
		}
	}
}

func TestBezierLength(t *testing.T) {
	testCurve := func(t *testing.T, c BezierCurve) {
		expected := 0.0
		n := 100000
		for i := 0; i < n; i++ {
			t1 := float64(i) / float64(n)
			t2 := float64(i+1) / float64(n)
			expected += c.Eval(t1).Dist(c.Eval(t2))
		}
		actual := c.Length(1e-8, 0)
		if math.Abs(actual-expected) > 1e-7 {
			t.Errorf("expected length %f but got %f", expected, actual)
		}
	}
	t.Run("Quad", func(t *testing.T) {
		testCurve(t, BezierCurve{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 3, Y: 1},
		})
	})
	t.Run("Cubic", func(t *testing.T) {
		testCurve(t, BezierCurve{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
		})

		// This curve seems to push close to the error bound
		// for cubics.
		testCurve(t, BezierCurve{
			{-0.15410584557718454, -0.49385517630417675},
			{-0.15154246403697602, -0.5155318553225691},
			{0.8225331751788417, 1.5556997675590987},
			{1.5195600606132607, 0.8079498332492203},
		})

		for i := 0; i < 100; i++ {
			c := BezierCurve{
				NewCoordRandNorm(),
				NewCoordRandNorm(),
				NewCoordRandNorm(),
				NewCoordRandNorm(),
			}
			expected := c.length(1e-6, 16)
			actual := c.Length(9e-5, 0)
			if math.Abs(actual-expected) > 1e-4 {
				t.Errorf("incorrect length for %v: got %f expected %f (error %f)",
					c, actual, expected, math.Abs(actual-expected))
			}
		}
	})
	t.Run("Quartic", func(t *testing.T) {
		testCurve(t, BezierCurve{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
			Coord{X: 4, Y: 2},
		})
	})
}

func TestSmoothBezier(t *testing.T) {
	actualCurve := SmoothBezier(
		Coord{}, Coord{X: 1, Y: 1}, Coord{X: 2, Y: 1}, Coord{X: 2},
		Coord{X: 3, Y: -1}, Coord{X: 4},
		Coord{X: 5, Y: 1}, Coord{X: 5},
	)
	expectedCurve := JoinedCurve{
		BezierCurve{Coord{}, Coord{X: 1, Y: 1}, Coord{X: 2, Y: 1}, Coord{X: 2}},
		BezierCurve{Coord{X: 2}, Coord{X: 2, Y: -1}, Coord{X: 3, Y: -1}, Coord{X: 4}},
		BezierCurve{Coord{X: 4}, Coord{X: 5, Y: 1}, Coord{X: 5, Y: 1}, Coord{X: 5}},
	}
	for i := 0; i < 1000; i++ {
		ti := rand.Float64()
		actual := actualCurve.Eval(ti)
		expected := expectedCurve.Eval(ti)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	}
}

func TestJoinedCurveEval(t *testing.T) {
	curve := JoinedCurve{
		BezierCurve{
			Coord{X: 3, Y: 3},
			Coord{X: 3, Y: 2},
			Coord{X: 4, Y: 4},
		},
		BezierCurve{
			Coord{X: 4, Y: 4},
			Coord{X: 5, Y: -2},
			Coord{X: 3, Y: 3},
		},
		BezierCurve{
			Coord{X: 3, Y: 3},
			Coord{X: 5, Y: 2},
			Coord{X: -1, Y: 3},
		},
	}
	actualExpected := [][2]Coord{
		{curve.Eval(0.0), curve[0].Eval(0.0)},
		{curve.Eval(0.1), curve[0].Eval(0.1 * 3)},
		{curve.Eval(0.4), curve[1].Eval((0.4 - 1.0/3) * 3)},
		{curve.Eval(0.8), curve[2].Eval((0.8 - 2.0/3) * 3)},
		{curve.Eval(1.0), curve[2].Eval(1.0)},
		{curve.Eval(1.1), curve[2].Eval((1.1 - 2.0/3) * 3)},
		{curve.Eval(-0.1), curve[0].Eval(-0.1 * 3)},
		{curve.Eval(-10), curve[0].Eval(-10 * 3)},
	}
	for i, pair := range actualExpected {
		if d := pair[0].Dist(pair[1]); math.IsNaN(d) || d > 1e-5 {
			t.Errorf("case %d: expected %f but got %f", i, pair[1], pair[0])
		}
	}
}

func TestCurveTranspose(t *testing.T) {
	c1 := BezierCurve{
		Coord{X: 4, Y: 4},
		Coord{X: 5, Y: -2},
		Coord{X: 3, Y: 3},
	}
	c2 := CurveTranspose(c1)
	c1 = c1.Transpose()

	for i := 0; i < 1000; i++ {
		ti := rand.Float64()
		expected := c1.Eval(ti)
		actual := c2.Eval(ti)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	}
}

func BenchmarkBezierEval(b *testing.B) {
	b.Run("Order5", func(b *testing.B) {
		curve := BezierCurve{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
			Coord{X: 2, Y: 3},
		}
		for i := 0; i < b.N; i++ {
			curve.Eval(0.3)
		}
	})
	b.Run("Order7", func(b *testing.B) {
		curve := BezierCurve{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: 4},
			Coord{X: 0, Y: -3},
		}
		for i := 0; i < b.N; i++ {
			curve.Eval(0.3)
		}
	})
	b.Run("Order10", func(b *testing.B) {
		curve := BezierCurve{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: 4},
			Coord{X: 0, Y: -3},
			Coord{X: 3, Y: -7},
			Coord{X: 4, Y: -8},
			Coord{X: 5, Y: -9},
		}
		for i := 0; i < b.N; i++ {
			curve.Eval(0.3)
		}
	})
}

func BenchmarkBezierLength(b *testing.B) {
	c := BezierCurve{
		Coord{X: 1, Y: 3},
		Coord{X: 2, Y: 2},
		Coord{X: 2, Y: 3},
		Coord{X: 3, Y: -2},
	}
	for i := 0; i < b.N; i++ {
		c.Length(1e-6, 0)
	}
}
