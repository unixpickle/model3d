package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const Thickness = 0.1

func main() {
	log.Println("Creating objects...")
	mug := CreateMug()
	coffee := CreateMugContents()

	log.Println("Creating color func...")
	onlyContents := &model3d.SubtractedSolid{Positive: coffee, Negative: mug}
	contentsSDF := model3d.MeshToSDF(model3d.MarchingCubesSearch(onlyContents, 0.02, 8))
	colorFunc := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if contentsSDF.SDF(c) > -0.01 {
			return render3d.NewColorRGB(0.29, 0.15, 0.02)
		} else {
			return render3d.NewColor(1.0)
		}
	})

	log.Println("Creating mesh...")
	combined := model3d.JoinedSolid{mug, coffee}
	mesh := model3d.MarchingCubesSearch(combined, 0.008, 8)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)

	log.Println("Saving...")
	mesh.SaveMaterialOBJ("impossible_mug.zip", colorFunc.Cached().TriangleColor)
}
