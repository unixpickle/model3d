package main

import (
	"log"

	"github.com/unixpickle/model3d"
)

func main() {
	log.Println("Creating fork...")
	fork := model3d.MarchingCubesSearch(NewForkSolid(), 0.01, 8)
	fork.SaveGroupedSTL("fork.stl")

	log.Println("Creating wine glass...")
	wineGlass := model3d.MarchingCubesSearch(CreateWineGlass(), 0.02, 8)
	wineGlass.SaveGroupedSTL("wine_glass.stl")
}
