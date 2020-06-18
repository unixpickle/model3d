package model2d

import (
	"math/rand"
	"testing"
)

func TestMarchingSquares(t *testing.T) {
	solid := BitmapToSolid(testingBitmap())

	testMesh := func(mesh *Mesh) {
		if !mesh.Manifold() {
			t.Error("mesh is non-manifold")
		}

		meshSolid := NewColliderSolid(MeshToCollider(mesh))

		for i := 0; i < 1000; i++ {
			point := Coord{
				X: float64(rand.Intn(int(solid.Max().X) + 2)),
				Y: float64(rand.Intn(int(solid.Max().Y) + 2)),
			}
			if solid.Contains(point) != meshSolid.Contains(point) {
				t.Error("containment mismatch at:", point)
			}
		}

		if _, n := mesh.RepairNormals(0.001); n != 0 {
			t.Error("invalid normals")
		}
	}

	t.Run("Plain", func(t *testing.T) {
		mesh := MarchingSquares(solid, 1.0)
		testMesh(mesh)
	})

	t.Run("Search", func(t *testing.T) {
		mesh := MarchingSquaresSearch(solid, 1.0, 8)
		testMesh(mesh)
	})
}
