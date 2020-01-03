package model3d

import (
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
		return []*Triangle{&Triangle{polygon[0], polygon[1], polygon[2]}}
	} else if len(polygon) < 3 {
		panic("polygon needs at least three points")
	}

	basis1 := polygon[1].Sub(polygon[0])
	basis1 = basis1.Normalize()
	basis2 := polygon[2].Sub(polygon[0])
	basis2 = basis2.Add(basis1.Scale(-basis1.Dot(basis2)))
	basis2 = basis2.Normalize()

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
