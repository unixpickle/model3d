package model3d

import (
	"math"
	"testing"

	"github.com/unixpickle/model3d/model2d"
)

func TestMeshSDFConsistency(t *testing.T) {
	solid := sdfTestingSolid()
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
	solid := sdfTestingSolid()
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

func TestProfileSDF(t *testing.T) {
	profileSolid := model2d.JoinedSolid{
		&model2d.Circle{
			Center: model2d.Coord{X: 1, Y: 1},
			Radius: 1.3,
		},
		&model2d.Circle{
			Center: model2d.Coord{X: -0.2, Y: 0.2},
			Radius: 0.4,
		},
	}
	profileMesh := model2d.MarchingSquaresSearch(profileSolid, 0.01, 8)
	profile := model2d.MeshToSDF(profileMesh)

	for _, minZ := range []float64{-0.3, 0.5} {
		for _, maxZ := range []float64{minZ + 0.1, minZ + 0.3} {
			solid3d := ProfileSolid(profileSolid, minZ, maxZ)
			expected := MeshToSDF(MarchingCubesSearch(solid3d, 0.01, 8))
			actual := ProfileSDF(profile, minZ, maxZ)
			for i := 0; i < 10000; i++ {
				coord := NewCoord3DRandNorm()
				a := actual.SDF(coord)
				x := expected.SDF(coord)
				if math.Abs(a-x) > 0.02 {
					t.Fatalf("unexpected SDF at %v: expected %f but got %f (minZ %f maxZ %f)",
						coord, x, a, minZ, maxZ)
				}
			}
		}
	}
}

func TestProfilePointSDF(t *testing.T) {
	profileSolid := model2d.JoinedSolid{
		&model2d.Circle{
			Center: model2d.Coord{X: 1, Y: 1},
			Radius: 1.3,
		},
		&model2d.Circle{
			Center: model2d.Coord{X: -0.2, Y: 0.2},
			Radius: 0.4,
		},
	}
	profileMesh := model2d.MarchingSquaresSearch(profileSolid, 0.01, 8)
	profile := model2d.MeshToSDF(profileMesh)

	for _, minZ := range []float64{-0.3, 0.5} {
		for _, maxZ := range []float64{minZ + 0.1, minZ + 0.3} {
			pointSDF := ProfilePointSDF(profile, minZ, maxZ)
			for i := 0; i < 10000; i++ {
				coord := NewCoord3DRandNorm()
				closest, a := pointSDF.PointSDF(coord)
				x := pointSDF.SDF(coord)
				if math.Abs(a-x) > 0.02 {
					t.Fatalf("mismatched SDFs at %v", coord)
				}
				if math.Abs(math.Abs(a)-closest.Dist(coord)) > 1e-5 {
					t.Fatalf("closest point is not at the SDF distance: %v %v: expected %f but got %f",
						coord, closest, closest.Dist(coord), a)
				}
			}
		}
	}
}

func BenchmarkMeshSDFs(b *testing.B) {
	solid := sdfTestingSolid()
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

func sdfTestingSolid() Solid {
	return &TorusSolid{
		Center:      Coord3D{},
		Axis:        XYZ(1, 2, -0.5).Normalize(),
		InnerRadius: 0.2,
		OuterRadius: 0.7,
	}
}
