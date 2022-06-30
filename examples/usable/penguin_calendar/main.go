package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
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

	mesh, interior := model3d.MarchingCubesInterior(fullSolid, 0.02, 8)
	colorFunc := toolbox3d.JoinedSolidCoordColorFunc(interior, bodyColorFuncs...)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
	mesh.SaveMaterialOBJ("penguin.zip", colorFunc.TriangleColor)
}
