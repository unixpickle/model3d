package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	cup, colorFn := CupSolid()
	mesh := model3d.MarchingCubesSearch(cup, 0.02, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFn.RenderColor)
}
