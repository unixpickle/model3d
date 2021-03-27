package model2d

import (
	"math"
	"testing"
)

func TestVertexSlice(t *testing.T) {
	s1 := &Segment{
		XY(0, 1),
		XY(1, 0),
	}
	mesh := NewMesh()
	mesh.Add(s1)
	if len(mesh.VertexSlice()) != 2 {
		t.Error("unexpected number of vertices")
	}
	mesh.Remove(s1)
	if len(mesh.VertexSlice()) != 0 {
		t.Error("unexpected number of vertices")
	}
}

func TestNewMeshPolar(t *testing.T) {
	mesh := NewMeshPolar(func(theta float64) float64 {
		return math.Cos(theta) + 2
	}, 100)
	MustValidateMesh(t, mesh)
}

func TestNewMeshRect(t *testing.T) {
	mesh := NewMeshRect(XY(0.2, 0.3), XY(0.25, 0.5))
	MustValidateMesh(t, mesh)
}
