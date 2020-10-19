package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	CornerSmoothing = 0.04
	Thickness       = 0.05
)

func main() {
	heartMesh := model2d.MustReadBitmap("outline.png", nil).Mesh().SmoothSq(50)
	heartMesh = heartMesh.MapCoords(heartMesh.Min().Scale(-1).Add)
	heartMesh = heartMesh.MapCoords(heartMesh.Max().Recip().Mul)
	solid := model2d.JoinedSolid{
		model2d.NewColliderSolid(model2d.MeshToCollider(heartMesh)),
		&model2d.Rect{
			MinVal: model2d.XY(0.45, 0.8),
			MaxVal: model2d.XY(0.55, 2.5),
		},
		&model2d.Rect{
			MinVal: model2d.XY(0.5, 1.65+0.5),
			MaxVal: model2d.XY(0.8, 1.75+0.5),
		},
		&model2d.Rect{
			MinVal: model2d.XY(0.5, 1.85+0.5),
			MaxVal: model2d.XY(0.8, 1.95+0.5),
		},
	}

	model2d.Rasterize("rendering_2d.png", solid, 200.0)

	log.Println("Creating 3D solid...")
	mesh2d := model2d.MarchingSquaresSearch(solid, 0.01, 8)
	collider2d := model2d.MeshToCollider(mesh2d)
	collider3d := model3d.ProfileCollider(collider2d, -Thickness/2, Thickness/2)
	smoothEdges := model3d.NewColliderSolidHollow(collider3d, CornerSmoothing)
	solid3d := model3d.JoinedSolid{
		smoothEdges,
		model3d.ProfileSolid(solid, -Thickness/2, Thickness/2),
	}

	log.Println("Creating 3D mesh...")
	mesh3d := model3d.MarchingCubesSearch(solid3d, 0.01, 8)
	log.Println("Saving mesh...")
	mesh3d.SaveGroupedSTL("key.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_3d.png", mesh3d, 3, 3, 300, nil)
}
