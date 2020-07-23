package model2d

import "testing"

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
