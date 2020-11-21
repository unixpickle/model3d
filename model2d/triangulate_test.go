package model2d

import (
	"math"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestTriangulate(t *testing.T) {
	poly := []Coord{
		{0, 0},
		{1, 2},
		{3, 3},
		{1, 5},
		{-1, 3},
		{2, 3},
		{0, 2},
	}
	contained := []Coord{
		{0.2, 1},
		{0.5, 2},
		{2.5, 3},
		{1, 4},
		{1, 4.5},
		{-0.9, 3.01},
	}
	notContained := []Coord{
		{1, 0},
		{0, 2.5},
		{1, 1},
		{2.5, 4},
	}
	for i := 0; i < len(poly); i++ {
		p1 := append(append([]Coord{}, poly[i:]...), poly[:i]...)
		p2 := append([]Coord{}, p1...)
		essentials.Reverse(p2)
		for j, p := range [][]Coord{p1, p2} {
			tris := Triangulate(p)
			checkContained := func(c Coord) bool {
				for _, t := range tris {
					if triangle2DContains(t, c) {
						return true
					}
				}
				return false
			}
			for _, c := range contained {
				if !checkContained(c) {
					t.Errorf("mismatched result for contained point %v (case %d %d)", c, i, j)
				}
			}
			for _, c := range notContained {
				if checkContained(c) {
					t.Errorf("mismatched result for uncontained point %v (case %d %d)", c, i, j)
				}
			}
		}
	}
}

func TestTriangulateMeshBasic(t *testing.T) {
	mesh := NewMeshPolar(func(theta float64) float64 {
		return math.Cos(theta) + 1.5
	}, 30)
	tris := triangulateMesh(mesh)

	solid := NewColliderSolid(MeshToCollider(mesh))
	for i := 0; i < 1000; i++ {
		c := NewCoordRandBounds(solid.Min(), solid.Max())
		expected := solid.Contains(c)
		actual := false
		for _, t := range tris {
			if triangle2DContains(t, c) {
				actual = true
			}
		}
		if actual != expected {
			t.Fatalf("point %v: contains=%v but got %v", c, expected, actual)
		}
	}
}

func triangle2DContains(tri [3]Coord, p Coord) bool {
	v1 := tri[0].Sub(tri[1])
	v2 := tri[2].Sub(tri[1])
	mat := (&Matrix2{v1.X, v2.X, v1.Y, v2.Y}).Inverse()
	coords := mat.MulColumn(p.Sub(tri[1]))
	return coords.X >= 0 && coords.Y >= 0 && coords.X+coords.Y <= 1
}
