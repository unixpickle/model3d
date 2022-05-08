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
				Alpha:         50.0,
				SpecularColor: render3d.NewColor(0.1),
				DiffuseColor:  colors[i].Scale(0.9),
			},
		}
		fullObject = append(fullObject, render3d.Translate(obj, center))
	}

	backdrop := &render3d.ColliderObject{
		Collider: model3d.NewRect(model3d.XYZ(-100.0, 8.0, -100.0), model3d.XYZ(100.0, 10.1, 100.0)),
		Material: &render3d.PhongMaterial{DiffuseColor: render3d.NewColor(0.5)},
	}
	fullObject = append(fullObject, backdrop)

	lightObject := &render3d.ColliderObject{
		Collider: model3d.NewRect(model3d.XYZ(-4.0, -20.0, -4.0), model3d.XYZ(4.0, -20+0.1, 4.0)),
		Material: &render3d.PhongMaterial{EmissionColor: render3d.NewColor(20.0)},
	}
	fullObject = append(fullObject, lightObject)

	log.Println("Rendering...")
	renderer := &render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Y(-15), model3d.Coord3D{}, math.Pi/3.6),

		MaxDepth:   5,
		NumSamples: 4096,
		Antialias:  1.0,
		Cutoff:     1e-4,

		FocusPoints: []render3d.FocusPoint{
			&render3d.PhongFocusPoint{
				Target: model3d.XYZ(0, -20, 0),
				Alpha:  4.0,
				MaterialFilter: func(m render3d.Material) bool {
					return true
				},
			},
		},
		FocusPointProbs: []float64{0.3},

		LogFunc: func(p, samples float64) {
			log.Printf("Rendering %.1f%%...", p*100)
		},
	}
	img := render3d.NewImage(512, 512)
	renderer.Render(img, fullObject)
	img.Save("rendering.png")
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
	result := model3d.NewMeshIcosphere(model3d.Coord3D{}, 3.5, 2).VertexSlice()
	essentials.VoodooSort(result, func(i, j int) bool {
		return model3d.NewSegment(result[i], result[j])[0] == result[i]
	})
	return result
}
