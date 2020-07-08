package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	PrintSize      = 4
	PrintThickness = 0.3
	TextThickness  = 0.1
)

func main() {
	img := model2d.MustReadBitmap("text.png", nil).FlipY()
	bmpSolid := model2d.BitmapToSolid(img)
	bmpSolid = model2d.ScaleSolid(bmpSolid, PrintSize/float64(img.Width))
	fullSolid := model3d.JoinedSolid{
		&TextSolid{
			Text: bmpSolid,
		},
		HersheyKissSolid{Center: model3d.Coord3D{X: 0.8, Y: PrintSize - 0.9,
			Z: PrintThickness}},
		HersheyKissSolid{Center: model3d.Coord3D{X: PrintSize / 2, Y: PrintSize - 0.7,
			Z: PrintThickness}},
		HersheyKissSolid{Center: model3d.Coord3D{X: PrintSize - 0.8, Y: PrintSize - 0.9,
			Z: PrintThickness}},
	}
	m := model3d.MarchingCubesSearch(fullSolid, 0.01, 8).Blur(-1, -1, -1, -1, -1)
	m = m.FlattenBase(0)
	m.SaveGroupedSTL("kiss.stl")
	render3d.SaveRandomGrid("rendering.png", m, 3, 3, 300, nil)
}

type TextSolid struct {
	Text model2d.Solid
}

func (t *TextSolid) Min() model3d.Coord3D {
	return model3d.XYZ(0, 0, 0)
}

func (t *TextSolid) Max() model3d.Coord3D {
	return model3d.XYZ(PrintSize, PrintSize, TextThickness+PrintThickness)
}

func (t *TextSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(t, c) {
		return false
	}
	if !t.Text.Contains(c.XY()) {
		return c.Z < PrintThickness
	} else {
		return true
	}
}
