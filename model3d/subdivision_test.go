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
