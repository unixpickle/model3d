package model2d

import "testing"

func TestFindPolyline(t *testing.T) {
	meshes := []*Mesh{
		MarchingSquaresSearch(&Circle{Radius: 1}, 0.1, 8),
		MarchingSquaresSearch(JoinedSolid{
			&Circle{Radius: 1},
			&Circle{Radius: 1, Center: XY(2, 2)},
		}, 0.1, 8),
		NewMeshSegments([]*Segment{
			{XY(0, 0), XY(0, 1)},
			{XY(0, 1), XY(0, 2)},
			{XY(0, 2), XY(1, 3)},

			{XY(2, 0), XY(2, 1)},
			{XY(2, 1), XY(2, 2)},
			{XY(2, 2), XY(3, 3)},

			{XY(0, 4), XY(0, 5)},
			{XY(0, 5), XY(0, 6)},
			{XY(0, 6), XY(1, 7)},
		}),
	}
	numPaths := []int{1, 2, 3}
	for i, mesh := range meshes {
		expectedNumPaths := numPaths[i]
		actualNumPaths := 0
		recreatedMesh := NewMesh()
		findPolylines(mesh, func(points []Coord) {
			actualNumPaths++
			for i := 1; i < len(points); i++ {
				recreatedMesh.Add(&Segment{points[i-1], points[i]})
			}
		})
		if actualNumPaths != expectedNumPaths {
			t.Errorf("case %d: expected %d paths but got %d", i, expectedNumPaths, actualNumPaths)
		}
		if !meshesEqual(mesh, recreatedMesh) {
			t.Errorf("case %d: mesh is not equivalent to paths", i)
		}
	}
}
