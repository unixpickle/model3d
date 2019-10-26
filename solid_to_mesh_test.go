package model3d

import (
	"math"
	"testing"
)

func TestSolidToMeshSingularEdgesSimple(t *testing.T) {
	mesh := SolidToMesh(simpleSingular{}, 0.5, 0, 0, 0)
	if mesh.NeedsRepair() {
		t.Fatal("mesh needs repair")
	}
}

type simpleSingular struct{}

func (s simpleSingular) Min() Coord3D {
	return Coord3D{}
}

func (s simpleSingular) Max() Coord3D {
	return Coord3D{X: 1, Y: 1, Z: 1}
}

func (s simpleSingular) Contains(c Coord3D) bool {
	if c.Min(s.Min()) != s.Min() || c.Max(s.Max()) != s.Max() {
		return false
	}
	return (c.X < 0.1 && c.Y < 0.1 && c.Z < 0.1) || (c.X > 0.9 && c.Y > 0.9 && c.Z < 0.1)
}

func TestSolidToMeshSingularEdgesComplex(t *testing.T) {
	// Create an adversarial pumpkin mesh.
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat)
	}, 30)
	mesh.Iterate(func(t *Triangle) {
		t1 := *t
		for i, c := range t1 {
			t1[i] = c.Scale(0.9)
		}
		t1[0], t1[1] = t1[1], t1[0]
		mesh.Add(&t1)
	})
	collider := MeshToCollider(mesh)
	solid := NewColliderSolid(collider)
	mesh = SolidToMesh(solid, 0.1, 0, 0, 0)
	if mesh.NeedsRepair() {
		t.Fatal("mesh needs repair")
	}
}
