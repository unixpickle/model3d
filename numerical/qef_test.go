package numerical

import (
	"math"
	"testing"
)

func TestQEFDist(t *testing.T) {
	v1 := Vec3{3, 1, 4}
	v2 := Vec3{2, 7, 2}
	dMat := NewQEF4Dist(v1).Add(NewQEF4Dist(v2))
	v3 := Vec3{9, 10, 11}
	actual := dMat.Eval(v3)
	expected := v3.DistSquared(v1) + v3.DistSquared(v2)
	if math.Abs(actual-expected) > 1e-5 {
		t.Fatalf("expected %f but got %f", expected, actual)
	}
}

func TestQEFOuter(t *testing.T) {
	v := Vec4{3, 1, 4, 1}
	dMat := NewQEF4Outer(v)

	v1 := Vec3{2, 7, 1}
	actual := dMat.Eval(v1)
	expected := math.Pow(Vec4{v1[0], v1[1], v1[2], 1}.Dot(v), 2)
	if math.Abs(actual-expected) > 1e-5 {
		t.Fatalf("expected %f but got %f", expected, actual)
	}
}

func TestQEFSolve(t *testing.T) {
	v1 := Vec3{3, 1, 4}
	v2 := Vec3{2, 7, 2}
	dMat := NewQEF4Dist(v1).Add(NewQEF4Dist(v2))
	solution, det := dMat.Minimize()
	if math.Abs(det) < 1e-5 {
		t.Fatal("solution was singular")
	}
	expected := v1.Add(v2).Scale(0.5)
	if solution.Dist(expected) > 1e-8 {
		t.Fatalf("expected midpoint %#v but got %#v", expected, solution)
	}
}
