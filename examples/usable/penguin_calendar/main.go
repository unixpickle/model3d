package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	log.Println("Creating blocks...")
	CreateBlocks()

	log.Println("Creating body...")
	body, bodyColorFuncs := PenguinBody()
	base := model3d.NewRect(
		model3d.XYZ(-1.5, -1.5, -0.1),
		model3d.XYZ(1.5, 1.5, 0.01),
	)
	hole := model3d.NewRect(
		model3d.XYZ(-0.9, -1.3, 0.9),
		model3d.XYZ(-0.9+1.8, -1.3+1.3, 0.9+1.25),
	)
	fullSolid := model3d.JoinedSolid{
		base,
		&model3d.SubtractedSolid{Positive: body, Negative: hole},
	}
	bodyColorFuncs = append(bodyColorFuncs,
		base, toolbox3d.ConstantCoordColorFunc(render3d.NewColor(1.0)))

	log.Println("Creating mesh...")
	mesh, interior := model3d.MarchingCubesInterior(fullSolid, 0.02, 8)
	log.Println("Creating color func...")
	colorFunc := toolbox3d.JoinedSolidCoordColorFunc(interior, bodyColorFuncs...)
	log.Println("Rendering...")
	render3d.SaveRendering("rendering_penguin.png", mesh, model3d.XYZ(1.5, -8.0, 3.5), 512, 512,
		colorFunc.RenderColor)
	log.Println("Saving...")
	mesh.SaveMaterialOBJ("penguin.zip", colorFunc.Cached().TriangleColor)
}
