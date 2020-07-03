package model3d

import (
	"math/rand"
	"testing"
)

func TestLoopSubdivision(t *testing.T) {
	base := NewMeshRect(X(-1), XYZ(1, 1, 1))
	mesh := LoopSubdivision(base, 4)

	if mesh.NeedsRepair() {
		t.Error("mesh needs repair")
	}
	if len(mesh.SingularVertices()) > 0 {
		t.Error("mesh has singular vertices")
	}
	if _, n := mesh.RepairNormals(1e-8); n > 0 {
		t.Error("mesh has bad normals")
	}
}

func TestSubdivider(t *testing.T) {
	subdiv := NewSubdivider()
	mesh := NewMeshRect(X(-1), XYZ(1, 1, 1))
	subdiv.AddFiltered(mesh, func(p1, p2 Coord3D) bool {
		return rand.Intn(2) == 0
	})
	subdiv.Subdivide(mesh, func(p1, p2 Coord3D) Coord3D {
		x, y := p2.Sub(p1).OrthoBasis()
		return p1.Mid(p2).Add(x.Scale(1e-5 * rand.Float64())).Add(y.Scale(1e-5 * rand.Float64()))
	})

	if mesh.NeedsRepair() {
		t.Error("mesh needs repair")
	}
	if len(mesh.SingularVertices()) > 0 {
		t.Error("mesh has singular vertices")
	}
	if _, n := mesh.RepairNormals(1e-8); n > 0 {
		t.Error("mesh has bad normals")
	}
}
