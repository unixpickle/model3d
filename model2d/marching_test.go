package model2d

import (
	"math/rand"
	"testing"
)

func TestMarchingSquares(t *testing.T) {
	solid := BitmapToSolid(testingBitmap())

	testMesh := func(mesh *Mesh) {
		MustValidateMesh(t, mesh)

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

func TestMarchingSquaresASCII(t *testing.T) {
	expected :=
		`                                                                ` + "\n" +
			`                 /\                          /\                 ` + "\n" +
			`        ________/  \________        ________/  \________        ` + "\n" +
			`       /                    \      /                    \       ` + "\n" +
			`      /                      \    /                      \      ` + "\n" +
			`     /                        \  /                        \     ` + "\n" +
			`    /                          \/                          \    ` + "\n" +
			`   /                                                        \   ` + "\n" +
			`  /                                                          \  ` + "\n" +
			` /                                                            \ ` + "\n" +
			` \                                                            / ` + "\n" +
			`  \                                                          /  ` + "\n" +
			`   \                                                        /   ` + "\n" +
			`    \                          /\                          /    ` + "\n" +
			`     \                        /  \                        /     ` + "\n" +
			`      \                      /    \                      /      ` + "\n" +
			`       \________    ________/      \________    ________/       ` + "\n" +
			`                \  /                        \  /                ` + "\n" +
			`                 \/                          \/                 ` + "\n" +
			`                                                                `
	solid := JoinedSolid{
		&Circle{Radius: 8.0},
		&Circle{Center: X(14), Radius: 8.0},
	}
	ascii := MarchingSquaresASCII(solid, 1.0)
	if ascii != expected {
		t.Errorf("expected:\n----\n%s\n----\nbut got:\n----\n%s\n----\n", expected, ascii)
	}
}
