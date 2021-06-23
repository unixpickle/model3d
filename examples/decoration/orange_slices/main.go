package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	solid := model3d.JoinedSolid{
		NewPeel(),
		NewWedge(-0.99, -0.18),
		NewWedge(0.18, 0.99),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, ColorFunc())
	log.Println("Saving...")
	mesh.SaveGroupedSTL("peel.stl")
}

func ColorFunc() render3d.ColorFunc {
	peel := NewPeel()
	peelMesh := model3d.MarchingCubesSearch(peel, 0.01, 8)
	peelSDF := model3d.MeshToSDF(peelMesh)
	return func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
		if peelSDF.SDF(c) > -0.01 {
			// Make the peel slightly brighter.
			return render3d.NewColorRGB(1.0, 0.63*1.1, 0.0)
		}
		return render3d.NewColorRGB(1.0, 0.63, 0.0)
	}
}
