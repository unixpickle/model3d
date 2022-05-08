package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	mesh := CreateGolfBall()
	render3d.SaveRandomGrid("test.png", mesh, 3, 3, 300, nil)
}

func CreateGolfBall() *model3d.Mesh {
	sphere := &model3d.Sphere{Radius: 1.0}

	icosphere := model3d.NewMeshIcosphere(sphere.Center, sphere.Radius, 5)
	dimples := model3d.JoinedSolid{}
	for _, c := range icosphere.VertexSlice() {
		dimples = append(dimples, &model3d.Sphere{Center: c.Scale(1.08), Radius: 0.15})
	}
	subtracted := &model3d.SubtractedSolid{
		Positive: sphere,
		Negative: dimples.Optimize(),
	}
	fineMesh := model3d.MarchingCubesSearch(subtracted, 0.01, 8)
	smoothSolid := model3d.NewColliderSolidInset(model3d.MeshToCollider(fineMesh), -0.05)
	return model3d.MarchingCubesSearch(smoothSolid, 0.02, 8)
}
