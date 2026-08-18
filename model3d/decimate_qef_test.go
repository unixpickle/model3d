package model3d

import (
	"math/rand"
	"testing"
)

func TestDecimateQEFMinimal(t *testing.T) {
	m := NewMesh()
	m.Add(&Triangle{
		Coord3D{0, 0, 1},
		Coord3D{1, 0, 0},
		Coord3D{0, 1, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{1, 0, 0},
		Coord3D{0, 1, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{0, 0, 1},
		Coord3D{0, 1, 0},
	})
	m.Add(&Triangle{
		Coord3D{0, 0, 0},
		Coord3D{1, 0, 0},
		Coord3D{0, 0, 1},
	})
	if m.NeedsRepair() {
		t.Fatal("invalid initial mesh")
	}

	m = QEFDecimate(m, 0, nil)
	if len(m.faces) != 4 {
		t.Error("invalid reduction")
	}
}

func testDecimateQEFMesh(t *testing.T, startMesh *Mesh) {
	for j := 0; j < 30; j++ {
		m := QEFDecimate(startMesh, 100, nil)
		if m.NumTriangles() > 100 {
			t.Fatalf("expected <=100 triangles but got %d", m.NumTriangles())
		}
		if m.NeedsRepair() {
			t.Error("needs repair")
		}
		if len(m.SingularVertices()) != 0 {
			t.Error("has singular vertices")
		}
		if len(m.TriangleSlice()) == 0 {
			t.Error("no triangles")
		}
		if !m.Orientable() {
			t.Error("bad normals")
		}
	}
}

func TestDecimateQEFSphere(t *testing.T) {
	testDecimateQEFMesh(t, NewMeshPolar(func(g GeoCoord) float64 {
		return 1.0
	}, 50))
}

func TestDecimateQEFCube(t *testing.T) {
	testDecimateQEFMesh(t, MarchingCubesSearch(NewRect(Origin, Ones(1)), 0.1, 4))
}

func TestDecimateQEFRandom(t *testing.T) {
	m := MarchingCubesSearch(&randomSolid{rng: rand.New(rand.NewSource(0))}, 0.05, 4)
	m = QEFDecimate(m, 1, nil)
	if m.NeedsRepair() {
		t.Error("needs repair")
	}
	if len(m.SingularVertices()) != 0 {
		t.Error("has singular vertices")
	}
	if len(m.TriangleSlice()) == 0 {
		t.Error("no triangles")
	}
}

func BenchmarkDecimateQEF(b *testing.B) {
	m := NewMeshPolar(func(g GeoCoord) float64 {
		return 1.0
	}, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QEFDecimate(m, 100, nil)
	}
}
