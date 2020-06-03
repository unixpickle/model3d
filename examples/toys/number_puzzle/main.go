package main

import (
	"log"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	placements := SearchPlacement(AllDigits(), 5)

	log.Println("Creating board...")
	boardSolid := BoardSolid(placements, 5)
	board := model3d.MarchingCubesSearch(boardSolid, 0.01, 8)
	board = board.EliminateCoplanar(1e-5)
	board.SaveGroupedSTL("board.stl")

	saveMesh := model3d.NewMesh()
	saveY := 0.0
	renderModel := render3d.JoinedObject{render3d.Objectify(board, nil)}
	for i, d := range placements {
		log.Println("Creating digit", i+1, "...")
		solid := DigitSolid(d)
		mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
		mesh = mesh.EliminateCoplanar(1e-5)

		// Scale down the mesh a tiny bit so that it fits in nicely
		// with corners of other digits.
		mid := mesh.Min().Mid(mesh.Max())
		mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
			return c.Sub(mid).Scale(0.99).Add(mid)
		})

		saveMesh.AddMesh(mesh.MapCoords(model3d.Coord3D{Y: saveY}.Sub(mesh.Min()).Add))
		saveY += mesh.Max().Y - mesh.Min().Y + 0.04

		color := render3d.NewColorRGB(rand.Float64(), rand.Float64(), rand.Float64())
		object := render3d.Objectify(mesh,
			func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
				return color
			})
		renderModel = append(renderModel, object)
	}

	render3d.SaveRendering("rendering.png", renderModel, model3d.Coord3D{X: 2.5, Y: -3, Z: 6},
		500, 500, nil)
	saveMesh.SaveGroupedSTL("digits.stl")
}
