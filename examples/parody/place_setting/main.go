package main

import (
	"log"

	"github.com/unixpickle/model3d"
)

func main() {
	log.Println("Creating wine glass...")
	wineGlass := model3d.MarchingCubesSearch(CreateWineGlass(), 0.01, 8)
	wineGlass.SaveGroupedSTL("wine_glass.stl")
}
