package model2d

import (
	"math"
	"math/rand"
	"testing"
)

func TestBezierEval(t *testing.T) {
	curves := []BezierCurve{
		{
			Coord{X: 1, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 3, Y: -2},
			Coord{X: 2, Y: 3},
		},
		{
			Coord{X: 3, Y: 3},
			Coord{X: 2, Y: 2},
			Coord{X: 2, Y: 3},
			Coord{X: 1, Y: -2},
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
