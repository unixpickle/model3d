package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const Production = true

func main() {
	cup, cupColor := CupSolid()
	cream, creamColor := CreamSolid()
	straw, strawColor := StrawSolid()
	cherry, cherryColor := CherrySolid()
	joined := model3d.JoinedSolid{cup, cream, straw, cherry}

	delta := 0.02
	if Production {
		delta = 0.01
	}
	mesh, interior := model3d.DualContourInterior(joined, delta, true, false)
	mesh = model3d.MeshToHierarchy(mesh)[0].Mesh
	colorFunc := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		cup, cupColor,
		cream, creamColor,
		straw, strawColor,
		cherry, cherryColor,
	)

	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
	mesh.SaveMaterialOBJ("milkshake.zip", colorFunc.TriangleColor)
}
