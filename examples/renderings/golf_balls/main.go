package main

import (
	"math"

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
	coordTree := model3d.NewCoordTree(icosphere.VertexSlice())
	dimples := model3d.JoinedSolid{}
	for _, c := range icosphere.VertexSlice() {
		surfaceRadius := 0.5 * coordTree.KNN(2, c)[1].Dist(c)
		radius := 0.2
		// Find sin(theta) such that cos(theta)*radius = surfaceRadius
		// cos(theta) = surfaceRadius/radius
		// sin(theta) = sqrt(1 - (surfaceRadius/radius)^2)
		offset := radius * math.Sqrt(1-math.Pow(surfaceRadius/radius, 2))
		dimples = append(dimples, &model3d.Sphere{Center: c.Scale(1 + offset), Radius: radius})
	}
	subtracted := &model3d.SubtractedSolid{
		Positive: sphere,
		Negative: dimples.Optimize(),
	}
	return model3d.MarchingCubesSearch(subtracted, 0.01, 8)
}
