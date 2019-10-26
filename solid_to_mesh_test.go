package model3d

import (
	"math"
	"testing"
)

func TestSolidToMeshSingularEdges(t *testing.T) {
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
