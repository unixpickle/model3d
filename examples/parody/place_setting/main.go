package main

import (
	"log"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
)

func main() {
	log.Println("Creating fork...")
	fork := model3d.MarchingCubesSearch(NewForkSolid(), 0.01, 8)
	fork.SaveGroupedSTL("fork.stl")
	render3d.SaveRandomGrid("rendering_fork.png", fork, 3, 3, 300, nil)

	log.Println("Creating wine glass...")
	wineGlass := model3d.MarchingCubesSearch(CreateWineGlass(), 0.02, 8)
	wineGlass.SaveGroupedSTL("wine_glass.stl")
	render3d.SaveRandomGrid("rendering_wine_glass.png", wineGlass, 3, 3, 300, nil)
}
