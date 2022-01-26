package model3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
)

// Triangulate turns any simple polygon into a set of
// equivalent triangles.
//
// The polygon is passed as a series of points, in order.
// The first point is re-used as the ending point, so no
// ending should be explicitly specified.
func Triangulate(polygon []Coord2D) [][3]Coord2D {
	return model2d.Triangulate(polygon)
}

// TriangulateFace turns any simple polygon face into a
// set of triangles.
//
// If the points are not coplanar, the result is
// approximated.
func TriangulateFace(polygon []Coord3D) []*Triangle {
	if len(polygon) == 3 {
		return []*Triangle{{polygon[0], polygon[1], polygon[2]}}
	} else if len(polygon) < 3 {
		panic("polygon needs at least three points")
	}

	basis1 := polygon[1].Sub(polygon[0])
	basis1 = basis1.Normalize()

	// Find a point that is not co-linear with the first two.
	minDot := math.Inf(1)
	basis2, _ := basis1.OrthoBasis()
	for _, p := range polygon[2:] {
		v := p.Sub(polygon[0]).ProjectOut(basis1).Normalize()
		dot := math.Abs(v.Dot(basis1))
		if !math.IsNaN(dot) && !math.IsInf(dot, 0) && dot < minDot {
			minDot = dot
			basis2 = v
		}
	}

	coords2D := make([]Coord2D, len(polygon))
	for i, p := range polygon {
		p1 := p.Sub(polygon[0])
		coords2D[i] = Coord2D{X: basis1.Dot(p1), Y: basis2.Dot(p1)}
	}
	triangles2D := Triangulate(coords2D)

	triangles := make([]*Triangle, len(triangles2D))
	for i, tri := range triangles2D {
		triangles[i] = &Triangle{}
		for j, p := range tri {
			triangles[i][j] = basis1.Scale(p.X).Add(basis2.Scale(p.Y)).Add(polygon[0])
		}
	}
	return triangles
}
