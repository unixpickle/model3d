package model2d

import (
	"math"
	"math/rand"
	"testing"
)

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
