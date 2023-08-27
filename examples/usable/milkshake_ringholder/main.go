package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	cup, cupColor := CupSolid()
	cream, creamColor := CreamSolid()
	straw, strawColor := StrawSolid()
	cherry, cherryColor := CherrySolid()
	joined := model3d.JoinedSolid{cup, cream, straw, cherry}

	mesh, interior := model3d.MarchingCubesInterior(joined, 0.02, 8)
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
