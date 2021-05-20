package model3d

import (
	"math/rand"
	"testing"
)

func TestLoopSubdivision(t *testing.T) {
	base := NewMeshRect(X(-1), XYZ(1, 1, 1))
	mesh := LoopSubdivision(base, 4)

	MustValidateMesh(t, mesh, false)
}

func TestSubdivideEdges(t *testing.T) {
	base := NewMeshTorus(XYZ(0.2, 0.3, 0.4), XY(0.5, 1.0).Normalize(), 0.2, 1.0, 5, 5)
	for i := 1; i < 6; i++ {
		mesh := SubdivideEdges(base, i)
		expectedN := len(base.TriangleSlice()) * i * i
		actualN := len(mesh.TriangleSlice())
		if actualN != expectedN {
			t.Errorf("expected %d triangles but got %d", expectedN, actualN)
		}
		MustValidateMesh(t, mesh, true)
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

	MustValidateMesh(t, mesh, false)
}
