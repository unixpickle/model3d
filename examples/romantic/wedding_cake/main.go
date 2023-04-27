package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	layers := model3d.JoinedSolid{}
	colors := []toolbox3d.CoordColorFunc{}
	curZ := 0.0
	for _, fn := range []func() (model3d.Solid, toolbox3d.CoordColorFunc){
		FlowerLayer, HexRoughLayer, DotsLayer, MarbleHexagonLayer, RoughRoundLayer,
	} {
		solid, color := fn()
		solid = model3d.TranslateSolid(solid, model3d.Z(curZ))
		color = color.Transform(&model3d.Translate{Offset: model3d.Z(curZ)})
		curZ = solid.Max().Z
		layers = append(layers, solid)
		colors = append(colors, color)
	}
	s, c := LetterCircle(layers[0])
	layers = append(layers, s)
	colors = append(colors, c)

	log.Println("Creating mesh and texture...")
	mesh, interior := model3d.MarchingCubesInterior(layers, 0.02, 8)
	var solidsAndColors []any
	for i, x := range layers {
		solidsAndColors = append(solidsAndColors, x, colors[i])
	}
	fullColor := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		solidsAndColors...,
	)

	log.Println("Saving...")
	mesh.SaveQuantizedMaterialOBJ("cake.zip", 16, fullColor.Cached().TriangleColor)

	log.Println("Rendering...")
	render3d.SaveRotatingGIF("panning.gif", mesh, model3d.Z(1), model3d.XZ(1, 0.4).Normalize(),
		300, 50, 10.0, fullColor.RenderColor)
}
