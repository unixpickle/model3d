package model2d

import (
	"math"
	"testing"
)

func TestMeshSDFConsistency(t *testing.T) {
	solid := sdfTestingSolid()
	mesh := MarchingSquaresSearch(solid, 0.02, 8)

	approxSDF := ColliderToSDF(MeshToCollider(mesh), 64)
	exactSDF := MeshToSDF(mesh)

	for i := 0; i < 100; i++ {
		c := NewCoordRandNorm()
		approx := approxSDF.SDF(c)
		exact := exactSDF.SDF(c)
		if math.Abs(approx-exact) > 1e-5 {
			t.Fatalf("bad SDF value: approx=%f but exact=%f", approx, exact)
		}
	}
}

func TestMeshPointSDF(t *testing.T) {
	solid := sdfTestingSolid()
	mesh := MarchingSquaresSearch(solid, 0.02, 8)
	sdf := MeshToSDF(mesh)

	for i := 0; i < 100; i++ {
		c := NewCoordRandNorm()
		expectedSDF := sdf.SDF(c)
		cp, actualSDF := sdf.PointSDF(c)
		if math.Abs(actualSDF-expectedSDF) > 1e-5 {
			t.Errorf("invalid SDF value: actual=%f expected=%f", actualSDF, expectedSDF)
		}
		expectedDist := math.Abs(actualSDF)
		actualDist := cp.Dist(c)
		if math.Abs(actualDist-expectedDist) > 1e-5 {
			t.Errorf("invalid closest point: expected dist %f but got %f", expectedDist, actualDist)
		}
	}
}

func sdfTestingSolid() Solid {
	return JoinedSolid{
		&Circle{
			Center: Coord{X: 0.3},
			Radius: 0.2,
		},
		&Circle{
			Center: Coord{X: 0.2, Y: 0.1},
			Radius: 0.2,
		},
	}
}
