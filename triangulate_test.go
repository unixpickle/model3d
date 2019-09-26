package model3d

import (
	"testing"

	"github.com/unixpickle/essentials"
)

func TestTriangulate(t *testing.T) {
	poly := []Coord2D{
		{0, 0},
		{1, 2},
		{3, 3},
		{1, 5},
		{-1, 3},
		{2, 3},
		{0, 2},
	}
	contained := []Coord2D{
		{0.2, 1},
		{0.5, 2},
		{2.5, 3},
		{1, 4},
		{1, 4.5},
		{-0.9, 3.01},
	}
	notContained := []Coord2D{
		{1, 0},
		{0, 2.5},
		{1, 1},
		{2.5, 4},
	}
	for i := 0; i < len(poly); i++ {
		p1 := append(append([]Coord2D{}, poly[i:]...), poly[:i]...)
		p2 := append([]Coord2D{}, p1...)
		essentials.Reverse(p2)
		for j, p := range [][]Coord2D{p1, p2} {
			tris := Triangulate(p)
			checkContained := func(c Coord2D) bool {
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

func triangle2DContains(tri [3]Coord2D, p Coord2D) bool {
	v1 := tri[0].Add(tri[1].Scale(-1))
	v2 := tri[2].Add(tri[1].Scale(-1))
	mat := (&Matrix2{v1.X, v2.X, v1.Y, v2.Y}).Inverse()
	coords := mat.MulColumn(p.Add(tri[1].Scale(-1)))
	return coords.X >= 0 && coords.Y >= 0 && coords.X+coords.Y <= 1
}
