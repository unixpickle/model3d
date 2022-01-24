package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	solid := model3d.JoinedSolid{
		FlowerPot(),
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 512, nil)
	mesh.SaveGroupedSTL("arrangement.stl")
}
