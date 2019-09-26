package model3d

import (
	"math"

	"github.com/unixpickle/essentials"
)

// Triangulate turns any simple polygon into a set of
// equivalent triangles.
//
// The polygon is passed as a series of points, in order.
// The first point is re-used as the ending point, so no
// ending should be explicitly specified.
func Triangulate(polygon []Coord2D) [][3]Coord2D {
	if len(polygon) == 3 {
		return [][3]Coord2D{[3]Coord2D{polygon[0], polygon[1], polygon[2]}}
	} else if len(polygon) < 3 {
		panic("polygon needs at least three points")
	}

	for i := range polygon {
		if isVertexEar(polygon, i) {
			p1 := polygon[(i+len(polygon)-1)%len(polygon)]
			p3 := polygon[(i+1)%len(polygon)]
			newPoly := append([]Coord2D{}, polygon...)
			essentials.OrderedDelete(newPoly, i)
			return append(Triangulate(newPoly), [3]Coord2D{p1, polygon[i], p3})
		}
	}
	panic("no ears detected")
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

	basis1 := polygon[1].Add(polygon[0].Scale(-1))
	basis1 = basis1.Scale(1 / basis1.Norm())
	basis2 := polygon[2].Add(polygon[0].Scale(-1))
	basis2 = basis2.Add(basis1.Scale(-basis1.Dot(basis2)))
	basis2 = basis2.Scale(1 / basis2.Norm())

	coords2D := make([]Coord2D, len(polygon))
	for i, p := range polygon {
		p = p.Add(polygon[0].Scale(-1))
		coords2D[i] = Coord2D{X: basis1.Dot(p), Y: basis2.Dot(p)}
	}
	triangles2D := Triangulate(coords2D)

	triangles := make([]*Triangle, len(coords2D))
	for i, tri := range triangles2D {
		for j, p := range tri {
			triangles[i][j] = basis1.Scale(p.X).Add(basis2.Scale(p.Y)).Add(polygon[0])
		}
	}
	return triangles
}

func isVertexEar(polygon []Coord2D, vertex int) bool {
	idx1 := (vertex + len(polygon) - 1) % len(polygon)
	idx3 := (vertex + 1) % len(polygon)
	p1 := polygon[idx1]
	p2 := polygon[vertex]
	p3 := polygon[idx3]

	psClockwise := isPolygonClockwise([]Coord2D{p1, p2, p3})
	clockwise := isPolygonClockwise(polygon)
	if psClockwise != clockwise {
		p1, p3 = p3, p1
	}

	v1 := p1.Add(p2.Scale(-1))
	v2 := p3.Add(p2.Scale(-1))
	n1 := v1.Scale(1 / v1.Norm())
	n2 := v2.Scale(1 / v2.Norm())

	theta := math.Acos(n1.Dot(n2))
	rotMat := Matrix2{math.Cos(theta), -math.Sin(theta), math.Sin(theta), math.Cos(theta)}
	rotatedV1 := rotMat.MulColumn(v1)
	if rotatedV1.Dot(v2) < 1-1e-8 {
		// This is not an interior corner.
		return false
	}

	inverseMat := (&Matrix2{v1.X, v2.X, v1.Y, v2.Y}).Inverse()

	for i, p := range polygon {
		if i == idx1 || i == vertex || i == idx3 {
			continue
		}
		coords := inverseMat.MulColumn(p.Add(p2.Scale(-1)))
		if coords.X > 0 && coords.Y > 0 || coords.X+coords.Y < 1 {
			// Another point lies inside this triangle.
			return false
		}
	}

	return true
}

// isPolygonClockwise checks if the polygon goes
// clockwise, assuming that the y-axis goes up and the
// x-axis goes to the right.
func isPolygonClockwise(polygon []Coord2D) bool {
	minX := polygon[0].X
	minIdx := 0
	for i, p := range polygon {
		if p.X < minX {
			minX = p.X
			minIdx = i
		}
	}
	nextPoint := polygon[(minIdx+1)%len(polygon)]
	return nextPoint.Y > polygon[minIdx].Y
}
