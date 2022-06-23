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
	mug := model3d.ScaleSolid(CreateMug(), 1.5)
	coffee := model3d.ScaleSolid(CreateMugContents(), 1.5)

	log.Println("Creating color func...")
	onlyContents := &model3d.SubtractedSolid{Positive: coffee, Negative: mug}
	contentsColl := model3d.MeshToCollider(model3d.MarchingCubesSearch(onlyContents, 0.015, 8))
	colorFunc := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if onlyContents.Contains(c) || contentsColl.SphereCollision(c, 0.015) {
			return render3d.NewColorRGB(0.29, 0.15, 0.02)
		} else {
			return render3d.NewColor(1.0)
		}
	})

	log.Println("Creating mesh...")
	combined := model3d.JoinedSolid{mug, coffee}
	mesh := model3d.MarchingCubesSearch(combined, 0.0075, 8)

	log.Println("Decimating mesh...")
	colorFunc = colorFunc.Cached()
	d := &model3d.Decimator{
		FeatureAngle:     0.03,
		BoundaryDistance: 1e-5,
		PlaneDistance:    1e-4,
		FilterFunc:       colorFunc.ChangeFilterFunc(mesh, 0.03),
	}
	prev := len(mesh.TriangleSlice())
	mesh = d.Decimate(mesh)
	post := len(mesh.TriangleSlice())
	log.Printf("went from %d -> %d triangles", prev, post)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)

	log.Println("Saving...")
	mesh.SaveMaterialOBJ("impossible_mug.zip", colorFunc.Cached().TriangleColor)
}
