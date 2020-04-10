package main

import (
	"log"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	solid := model3d.JoinedSolid{
		GenerateBase(),
		GenerateFlag(),
		GeneratePeople(),
	}

	log.Println("Creating mesh...")
	m := model3d.MarchingCubesSearch(solid, 0.02, 8).Blur(-1, -1, -1, -1, -1)

	log.Println("Saving mesh...")
	m.SaveGroupedSTL("statue.stl")

	log.Println("Saving rendering...")
	render3d.SaveRendering("rendering.png", m, model3d.Coord3D{Y: -10, Z: 5.5}, 900, 900, nil)
}
