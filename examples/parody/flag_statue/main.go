package main

import (
	"image"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
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
	img := image.NewGray(image.Rect(0, 0, 900, 900))
	model3d.RenderRayCast(model3d.MeshToCollider(m), img,
		model3d.Coord3D{Y: -10, Z: 5.5},
		model3d.Coord3D{X: 1},
		model3d.Coord3D{Z: -1, Y: -0.3},
		model3d.Coord3D{Z: -0.3, Y: 1},
		math.Pi/5)
	f, err := os.Create("rendering.png")
	essentials.Must(err)
	defer f.Close()
	png.Encode(f, img)
}
