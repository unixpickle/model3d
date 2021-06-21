package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	solid := model3d.JoinedSolid{
		NewPeel(),
		NewWedge(-0.99, -0.18),
		NewWedge(0.18, 0.99),
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
	mesh.SaveGroupedSTL("peel.stl")
}
