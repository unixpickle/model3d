// Generated from templates/surface_estimator_test.template

package model2d

import (
	"math"
	"testing"
)

func TestSurfaceEstimator(t *testing.T) {
	solid := &Circle{
		// Randomly chosen.
		Center: XY(0.5, 0.8),
		Radius: 1.0,
	}
	// Randomly chosen.
	dir := XY(-1.0, -3.0).Normalize()

	estimator := SolidSurfaceEstimator{Solid: solid}

	t.Run("Bisect", func(t *testing.T) {
		p1 := solid.Center.Add(dir.Scale(0.3))
		p2 := solid.Center.Add(dir.Scale(1.1))
		actual := estimator.Bisect(p1, p2)
		expected := solid.Center.Add(dir)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected point %v but got %v", expected, actual)
		}
	})

	t.Run("Normal", func(t *testing.T) {
		p := solid.Center.Add(dir)
		expected := dir
		actual := estimator.Normal(p)
		if actual.Dist(expected) > 1e-3 {
			t.Errorf("expected normal %v but got %v", expected, actual)
		}

		for i := 0; i < 1000; i++ {
			expected := NewCoordRandUnit()
			p := solid.Center.Add(expected)
			actual := estimator.Normal(p)
			if actual.Dist(expected) > 1e-3 {
				t.Errorf("expected normal %v but got %v", expected, actual)
			}
		}
	})

	t.Run("NormalES", func(t *testing.T) {
		est := estimator
		est.RandomSearchNormals = true
		est.NormalSamples = 128
		p := solid.Center.Add(dir)
		expected := dir
		actual := est.Normal(p)
		if math.Abs(expected.Norm()-1) > 1e-5 || actual.Dot(expected) < 0.5 {
			t.Errorf("expected normal %v but got %v", expected, actual)
		}
	})
}
