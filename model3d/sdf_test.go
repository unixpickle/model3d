package model3d

import (
	"math"
	"testing"
)

func TestMeshSDFConsistency(t *testing.T) {
	solid := &TorusSolid{
		Center:      Coord3D{},
		Axis:        Coord3D{X: 1, Y: 2, Z: -0.5}.Normalize(),
		InnerRadius: 0.2,
		OuterRadius: 0.7,
	}
	mesh := MarchingCubesSearch(solid, 0.02, 8)

	approxSDF := ColliderToSDF(MeshToCollider(mesh), 64)
	exactSDF := MeshToSDF(mesh)

	for i := 0; i < 100; i++ {
		c := NewCoord3DRandNorm()
		approx := approxSDF.SDF(c)
		exact := exactSDF.SDF(c)
		if math.Abs(approx-exact) > 1e-5 {
			t.Fatalf("bad SDF value: approx=%f but exact=%f", approx, exact)
		}
	}
}

func TestMeshPointSDF(t *testing.T) {
	solid := &TorusSolid{
		Center:      Coord3D{},
		Axis:        Coord3D{X: 1, Y: 2, Z: -0.5}.Normalize(),
		InnerRadius: 0.2,
		OuterRadius: 0.7,
	}
	mesh := MarchingCubesSearch(solid, 0.02, 8)
	sdf := MeshToSDF(mesh)

	for i := 0; i < 100; i++ {
		c := NewCoord3DRandNorm()
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

func BenchmarkMeshSDFs(b *testing.B) {
	solid := &TorusSolid{
		Center:      Coord3D{},
		Axis:        Coord3D{X: 1, Y: 2, Z: -0.5}.Normalize(),
		InnerRadius: 0.2,
		OuterRadius: 0.7,
	}
	mesh := MarchingCubesSearch(solid, 0.02, 8)

	approxSDF := ColliderToSDF(MeshToCollider(mesh), 64)
	exactSDF := MeshToSDF(mesh)

	runTests := func(b *testing.B, c Coord3D) {
		b.Run("Approx", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				approxSDF.SDF(c)
			}
		})

		b.Run("Exact", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				exactSDF.SDF(c)
			}
		})

		b.Run("ExactPoint", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				exactSDF.PointSDF(c)
			}
		})
	}

	b.Run("Center", func(b *testing.B) {
		runTests(b, Coord3D{})
	})
	b.Run("Edge", func(b *testing.B) {
		runTests(b, Coord3D{X: 0.9, Y: 0.9})
	})
}
