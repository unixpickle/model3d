package main

import (
	"log"
	"math"
	"math/rand"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	log.Println("Creating colors...")
	centers := SortedCenterCoords()
	colors := make([]render3d.Color, len(centers))
	for i := range colors {
		colors[i] = render3d.NewColorRGB(rand.Float64(), rand.Float64(), rand.Float64())
	}

	log.Println("Creating base mesh...")
	baseMesh := CreateGolfBall()
	baseCollider := model3d.MeshToCollider(baseMesh)
	log.Println("Creating full object...")
	fullObject := render3d.JoinedObject{}
	for i, center := range centers {
		obj := &render3d.ColliderObject{
			Collider: baseCollider,
			Material: &render3d.PhongMaterial{
				Alpha:         10.0,
				SpecularColor: render3d.NewColor(0.1),
				DiffuseColor:  colors[i].Scale(0.9),
			},
		}
		fullObject = append(fullObject, render3d.Translate(obj, center))
	}

	log.Println("Rendering...")
	render3d.SaveRandomGrid("test.png", fullObject, 3, 3, 300, nil)
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
	return model3d.MarchingCubesSearch(subtracted, 0.02, 8)
}

func SortedCenterCoords() []model3d.Coord3D {
	result := model3d.NewMeshIcosphere(model3d.Coord3D{}, 4.0, 2).VertexSlice()
	essentials.VoodooSort(result, func(i, j int) bool {
		return model3d.NewSegment(result[i], result[j])[0] == result[i]
	})
	return result
}
