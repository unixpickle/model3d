package model3d

import (
	"testing"
)

func TestNewMeshRect(t *testing.T) {
	mesh := NewMeshRect(Coord3D{X: -0.3, Y: -0.4, Z: -0.2}, Coord3D{X: 0.4, Y: 0.35, Z: 0.19})
	if mesh.NeedsRepair() {
		t.Error("mesh needs repair")
	}
	if len(mesh.TriangleSlice()) != 12 {
		t.Errorf("expected exactly 12 triangles, but got %d", len(mesh.TriangleSlice()))
	}
	if _, n := mesh.RepairNormals(1e-8); n != 0 {
		t.Errorf("found %d bad normals", n)
	}
}
