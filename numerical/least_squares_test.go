package numerical

import (
	"math"
	"testing"
)

func TestLeastSquares3(t *testing.T) {
	t.Run("SimpleCase", func(t *testing.T) {
		mat := &Matrix3{
			0.14717068, -0.51542179, 1.23163796, -0.96365947, 0.45653824,
			-0.17872864, -1.13401065, 1.40549596, -1.26707842,
		}
		x := Vec3{1.0, 2.0, 3.0}
		b := mat.MulColumn(x)
		rows := mat.Rows()
		actual := LeastSquares3(rows[:], b[:], 1e-6)

		if actual.Dist(x) > 1e-6 || math.IsNaN(actual.Norm()) {
			t.Errorf("expected %v but got %v", x, actual)
		}
	})
	t.Run("OverdeterminedExact", func(t *testing.T) {
		rows := []Vec3{
			{0.14717068, -0.51542179, 1.23163796},
			{-0.96365947, 0.45653824, -0.17872864},
			{-1.13401065, 1.40549596, -1.26707842},
			{-0.52692109, -0.97020696, -1.16253638},
			{-0.52692109, -0.97020696, -1.16253638},
		}
		x := Vec3{1.0, 2.0, 3.0}
		b := make([]float64, len(rows))
		for i, v := range rows {
			b[i] = v.Dot(x)
		}
		actual := LeastSquares3(rows[:], b[:], 1e-6)

		if actual.Dist(x) > 1e-6 || math.IsNaN(actual.Norm()) {
			t.Errorf("expected %v but got %v", x, actual)
		}
	})
	t.Run("OverdeterminedApprox", func(t *testing.T) {
		rows := []Vec3{
			{0.14717068, -0.51542179, 1.23163796},
			{-0.96365947, 0.45653824, -0.17872864},
			{-1.13401065, 1.40549596, -1.26707842},
			{-0.52692109, -0.97020696, -1.16253638},
			{-0.52692109, -0.97020696, -1.16253638},
		}
		x := Vec3{1.0, 2.0, 3.0}
		b := make([]float64, len(rows))
		for i, v := range rows {
			b[i] = v.Dot(x)
		}
		b[0] += 1.0
		actual := LeastSquares3(rows[:], b[:], 1e-6)

		expected := Vec3{0.42375846, 1.76412247, 3.49721109}
		if actual.Dist(expected) > 1e-6 || math.IsNaN(actual.Norm()) {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	})
	t.Run("Underdetermined", func(t *testing.T) {
		rows := []Vec3{
			{0.14717068, -0.51542179, 1.23163796},
			{-0.96365947, 0.45653824, -0.17872864},
		}
		x := Vec3{1.0, 2.0, 3.0}
		b := make([]float64, len(rows))
		for i, v := range rows {
			b[i] = v.Dot(x)
		}
		actual := LeastSquares3(rows[:], b[:], 1e-6)

		expected := Vec3{-0.09457603, -0.70187553, 2.000099}
		if actual.Dist(expected) > 1e-6 || math.IsNaN(actual.Norm()) {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	})
}
