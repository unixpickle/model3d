package main

import (
	"log"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	placements := SearchPlacement(AllDigits(), 5)

	saveMesh := model3d.NewMesh()
	saveY := 0.0

	var renderModel render3d.JoinedObject
	for i, d := range placements {
		log.Println("Creating solid", i+1)
		solid := DigitSolid(d)
		mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
		mesh = mesh.EliminateCoplanar(1e-5)

		saveMesh.AddMesh(mesh.MapCoords(model3d.Coord3D{Y: saveY}.Sub(mesh.Min()).Add))
		saveY += mesh.Max().Y - mesh.Min().Y + 0.04

		color := render3d.NewColorRGB(rand.Float64(), rand.Float64(), rand.Float64())
		object := render3d.Objectify(mesh,
			func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
				return color
			})
		renderModel = append(renderModel, object)
	}

	render3d.SaveRandomGrid("rendering.png", renderModel, 3, 3, 300, nil)
	saveMesh.SaveGroupedSTL("digits.stl")
}
