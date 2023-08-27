package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	cup, cupColor := CupSolid()
	cream, creamColor := CreamSolid()
	joined := model3d.JoinedSolid{cup, cream}

	mesh, interior := model3d.MarchingCubesInterior(joined, 0.02, 8)
	colorFunc := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		cup, cupColor,
		cream, creamColor,
	)

	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
}
