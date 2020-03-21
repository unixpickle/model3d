package model3d

import (
	"testing"
)

func TestDecimateMinimal(t *testing.T) {
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

	decimators := []*Decimator{
		// Extremely aggressive decimator.
		&Decimator{
			FeatureAngle:     0.0001,
			PlaneDistance:    2,
			BoundaryDistance: 2,
			EliminateCorners: true,
		},
		// A more normal setup.
		&Decimator{
			FeatureAngle:     0.1,
			PlaneDistance:    0.1,
			BoundaryDistance: 0.1,
		},
	}
	for _, d := range decimators {
		elim := d.Decimate(m)
		if len(elim.triangles) != len(m.triangles) {
			t.Error("invalid reduction")
		}
	}
}

func TestDecimateSphere(t *testing.T) {
	m := NewMeshPolar(func(g GeoCoord) float64 {
		return 1.0
	}, 50)

	decimators := []*Decimator{
		// Extremely aggressive decimator.
		&Decimator{
			FeatureAngle:     0.0001,
			PlaneDistance:    2,
			BoundaryDistance: 2,
			EliminateCorners: true,
		},
		// A more normal setup.
		&Decimator{
			FeatureAngle:     0.1,
			PlaneDistance:    0.05,
			BoundaryDistance: 0.05,
		},
	}
	for i, d := range decimators {
		elim := d.Decimate(m)
		if elim.NeedsRepair() {
			t.Errorf("decimator %d: needs repair", i)
		}
		if len(elim.SingularVertices()) != 0 {
			t.Errorf("decimator %d: has singular vertices", i)
		}
		if len(elim.TriangleSlice()) == 0 {
			t.Errorf("decimator %d: no triangles", i)
		}
	}
}
