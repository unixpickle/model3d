package numerical

import (
	"math"
	"testing"
)

func TestRecursiveLineSearch(t *testing.T) {
	rs := &RecursiveLineSearch[Vec3]{LineSearch: LineSearch{Stops: 10, Recursions: 4}}
	solution, value := rs.Minimize(Vec3{-3, -3, -3}, Vec3{3, 3, 3}, func(v Vec3) float64 {
		return math.Pow(v[0]-1, 2) + math.Pow(v[1]+1, 2) + math.Pow(v[2]-2, 4) + 1
	})
	expected := Vec3{1, -1, 2}
	if math.Abs(value-1) > 1e-3 {
		t.Errorf("expected value %f but got %f", 1.0, value)
	}
	if solution.Dist(expected) > 1e-3 {
		t.Errorf("expected %v but got %v", expected, solution)
	}
}
