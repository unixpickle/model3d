package toolbox3d

import (
	"fmt"
	"math"
	"testing"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/numerical"
)

func TestLineSearch(t *testing.T) {
	polynomials := [][3]float64{
		{1, 2, -3},
		{2, 3, 5},
		{-1, 2, 3},
	}

	polyMinMax := func(maximize bool, poly [3]float64) (float64, float64) {
		extremum := -poly[1] / (2 * poly[0])

		bestVal := math.Inf(1)
		if maximize {
			bestVal = -bestVal
		}
		bestX := 0.0

		for _, x := range []float64{-10, extremum, 10} {
			if x < -10 || x > 10 {
				// Avoid extremum outside of bounds.
				continue
			}
			y := poly[0]*x*x + poly[1]*x + poly[2]
			if (maximize && y > bestVal) || (!maximize && y < bestVal) {
				bestVal = y
				bestX = x
			}
		}

		return bestX, bestVal
	}

	search := LineSearch{Stops: 10, Recursions: 10}
	for _, poly := range polynomials {
		t.Run(fmt.Sprintf("%f-%f-%f", poly[0], poly[1], poly[2]), func(t *testing.T) {
			polyFunc := func(x float64) float64 {
				return poly[0]*x*x + poly[1]*x + poly[2]
			}
			actualX, actualY := search.Minimize(-10, 10, polyFunc)
			expectedX, expectedY := polyMinMax(false, poly)
			if math.Abs(actualX-expectedX) > 1e-5 || math.Abs(actualY-expectedY) > 1e-5 {
				t.Errorf("poly %v should have minimum (%f,%f) but got (%f,%f)",
					poly, expectedX, expectedY, actualX, actualY)
			}
			actualX, actualY = search.Maximize(-10, 10, polyFunc)
			expectedX, expectedY = polyMinMax(true, poly)
			if math.Abs(actualX-expectedX) > 1e-5 || math.Abs(actualY-expectedY) > 1e-5 {
				t.Errorf("poly %v should have maximum (%f,%f) but got (%f,%f)",
					poly, expectedX, expectedY, actualX, actualY)
			}
		})
	}
}

func TestSolidBounds(t *testing.T) {
	rawSolid := model3d.VecScaleSolid(
		&model3d.Sphere{Radius: 1.0, Center: model3d.XYZ(1, 2, 3)},
		model3d.XYZ(2.0, 0.75, 1.5),
	)
	min, max := rawSolid.Min(), rawSolid.Max()
	looseBounds := model3d.ForceSolidBounds(rawSolid, min.AddScalar(-1), max.AddScalar(1))

	search := LineSearch3D{LineSearch: numerical.LineSearch{Stops: 20, Recursions: 5}}
	tightMin, tightMax := search.SolidBounds(looseBounds)
	if tightMin.Dist(min) > 1e-4 {
		t.Errorf("expected min %v but got %v", min, tightMin)
	}
	if tightMax.Dist(max) > 1e-4 {
		t.Errorf("expected max %v but got %v", max, tightMax)
	}
}
