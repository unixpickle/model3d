package model2d

import "testing"

func TestRasterizeCollider(t *testing.T) {
	shape := &Circle{Radius: 40}
	mesh := MarchingSquaresSearch(shape, 0.1, 8)
	collider := MeshToCollider(mesh)

	rast := &Rasterizer{LineWidth: 1, Scale: 2}
	render1 := rast.RasterizeCollider(collider)
	render2 := rast.RasterizeSolid(NewColliderSolidHollow(collider, 0.25))

	if render1.Bounds() != render2.Bounds() {
		t.Fatal("differing bounds", render1.Bounds(), render2.Bounds())
	}

	for y := 0; y < render1.Bounds().Dy(); y++ {
		for x := 0; x < render1.Bounds().Dx(); x++ {
			px1 := render1.GrayAt(x, y)
			px2 := render2.GrayAt(x, y)
			if px1 != px2 {
				t.Errorf("different color at %d,%d: got %v but expected %v", x, y, px1, px2)
			}
		}
	}
}

func TestRasterizeColliderSolid(t *testing.T) {
	shape := &Circle{Radius: 40}
	mesh := MarchingSquaresSearch(shape, 0.1, 8)
	collider := MeshToCollider(mesh)

	rast := &Rasterizer{Scale: 2}
	render1 := rast.RasterizeColliderSolid(collider)
	render2 := rast.RasterizeSolid(NewColliderSolid(collider))

	if render1.Bounds() != render2.Bounds() {
		t.Fatal("differing bounds", render1.Bounds(), render2.Bounds())
	}

	for y := 0; y < render1.Bounds().Dy(); y++ {
		for x := 0; x < render1.Bounds().Dx(); x++ {
			px1 := render1.GrayAt(x, y)
			px2 := render2.GrayAt(x, y)
			if px1 != px2 {
				t.Errorf("different color at %d,%d: got %v but expected %v", x, y, px1, px2)
			}
		}
	}
}
