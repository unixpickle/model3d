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
		for i := 0; i < 10; i++ {
			xValue := rand.Float64()*1.9 + 1.05
			tValue := curve.InverseX(xValue)
			if math.Abs(xValue-curve.Eval(tValue).X) > 1e-8 {
				t.Errorf("invalid inverse for x: %f", xValue)
			}
		}
		if !math.IsNaN(curve.InverseX(0.5)) {
			t.Error("not NaN before bounds")
		}
		if !math.IsNaN(curve.InverseX(3.5)) {
			t.Error("not NaN after bounds")
		}
	}
}
