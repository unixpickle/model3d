package model2d

import (
	"math/rand"
	"testing"
)

func TestMeshRepairNormals(t *testing.T) {
	mesh := MarchingSquaresSearch(&Circle{Radius: 0.3}, 0.01, 8)
	if !mesh.Manifold() {
		t.Fatal("bad initial mesh")
	}

	mesh1, numRepairs := mesh.RepairNormals(1e-8)
	if numRepairs > 0 {
		t.Errorf("expected 0 repairs but got: %d", numRepairs)
	}
	if !meshesEqual(mesh, mesh1) {
		t.Error("meshes are not equal")
	}

	flipped := NewMesh()
	expectedFlipped := 0
	mesh.Iterate(func(s *Segment) {
		if rand.Intn(2) == 0 {
			flipped.Add(s)
		} else {
			s1 := *s
			s1[0], s1[1] = s1[1], s1[0]
			flipped.Add(&s1)
			expectedFlipped++
		}
	})
	mesh1, numRepairs = flipped.RepairNormals(1e-8)
	if numRepairs != expectedFlipped {
		t.Errorf("expected %d repairs but got %d", expectedFlipped, numRepairs)
	}
	if !meshesEqual(mesh, mesh1) {
		t.Error("meshes are not equal")
	}
}

func meshesEqual(m1, m2 *Mesh) bool {
	seg1 := meshSegmentValues(m1)
	seg2 := meshSegmentValues(m2)
	if len(seg1) != len(seg2) {
		return false
	}
	for s, c := range seg1 {
		if seg2[s] != c {
			return false
		}
	}
	return true
}

func meshSegmentValues(m *Mesh) map[Segment]int {
	res := map[Segment]int{}
	m.Iterate(func(s *Segment) {
		res[*s]++
	})
	return res
}
