package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	solid := model3d.JoinedSolid{
		NewPeel(),
		NewWedge(-0.9999, -0.05),
		NewWedge(0.05, 0.9999),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	log.Println("Rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(0.0, -2.0, 2.0), 500, 500, ColorFunc())
	log.Println("Saving...")
	mesh.SaveGroupedSTL("peel.stl")
}

func ColorFunc() render3d.ColorFunc {
	peelMesh := PeelMesh(PeelStops / 2)
	peelSDF := model3d.MeshToSDF(peelMesh)
	return func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
		if peelSDF.SDF(c) > -(1e-3 + PeelRounding) {
			// Make the peel slightly brighter.
			return render3d.NewColorRGB(1.0, 0.63*1.1, 0.0)
		}
		return render3d.NewColorRGB(0.95, 0.60, 0.0)
	}
}
