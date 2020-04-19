package model3d

import (
	"testing"
)

func TestTriangulateFace(t *testing.T) {
	corner1 := Coord3D{1, 2, 3}
	side1 := Coord3D{-1, 1, 2}
	side2 := Coord3D{0, -2, 1}
	rect := []Coord3D{
		corner1,
		corner1.Add(side1),
		corner1.Add(side1).Add(side2),
		corner1.Add(side2),
	}
	triangles := TriangulateFace(rect)
	if len(triangles) != 2 {
		t.Fatalf("unexpected triangle count: %d", len(triangles))
	}
	expected := []*Triangle{
		&Triangle{rect[1], rect[2], rect[3]},
		&Triangle{rect[3], rect[0], rect[1]},
	}
	for i, act := range triangles {
		exp := expected[i]
		for j, p1 := range act {
			p2 := exp[j]
			if p1.Dist(p2) > 1e-8 {
				t.Fatalf("triangle %d: expected triangle %v but got %v", i, *exp, *act)
			}
		}
	}
}
