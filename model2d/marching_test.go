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

		mesh.Iterate(func(s *Segment) {
			delta := s.Normal().Scale(0.001)
			inside := s.Mid().Sub(delta)
			outside := s.Mid().Add(delta)
			if !meshSolid.Contains(inside) || meshSolid.Contains(outside) {
				t.Error("invalid normal:", s.Normal())
			}
		})
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
