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

	QEFDecimate(m, 0, nil)
	if len(m.faces) != 4 {
		t.Error("invalid reduction")
	}
}

func TestDecimateQEFSphere(t *testing.T) {
	for j := 0; j < 30; j++ {
		m := NewMeshPolar(func(g GeoCoord) float64 {
			return 1.0
		}, 50)
		QEFDecimate(m, 100, nil)
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
		if m.SelfIntersections() == 0 {
			if _, n := m.RepairNormals(1e-8); n != 0 {
				t.Error("bad normals")
			}
		}
	}
}

func TestDecimateQEFRandom(t *testing.T) {
	m := MarchingCubesSearch(&randomSolid{rng: rand.New(rand.NewSource(0))}, 0.05, 4)
	QEFDecimate(m, 1, nil)
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
