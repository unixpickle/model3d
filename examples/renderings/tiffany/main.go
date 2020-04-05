package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	scene := render3d.JoinedObject{
		NewGlobe(),
		NewWalls(),
		CreateTable(),
		CreateLight(),
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -13, Z: 2},
			model3d.Coord3D{Y: 0, Z: 2}, math.Pi/3.6),

		// Focus towards the area light.
		FocusPoints: []render3d.FocusPoint{
			&render3d.SphereFocusPoint{
				Center: model3d.Coord3D{Z: 5},
				Radius: 1,
			},
		},
		FocusPointProbs: []float64{0.5},

		MaxDepth:   5,
		NumSamples: 400,
		Antialias:  1.0,
		Cutoff:     1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	fmt.Println("Ray variance:", renderer.RayVariance(scene, 200, 200, 5))

	img := render3d.NewImage(200, 200)
	renderer.Render(img, scene)
	fmt.Println()
	img.Save("output.png")
}

func CreateTable() render3d.Object {
	createPiece := func(min, max model3d.Coord3D) render3d.Object {
		return &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(model3d.NewMeshRect(min, max)),
			Material: &render3d.PhongMaterial{
				Alpha:         5,
				SpecularColor: render3d.Color{X: 0.2, Y: 0.2, Z: 0.2},
				DiffuseColor:  render3d.Color{X: 0.8, Y: 0.8, Z: 0.8},
			},
		}
	}

	return render3d.JoinedObject{
		createPiece(model3d.Coord3D{X: -3, Y: -2, Z: -1.2},
			model3d.Coord3D{X: 3, Y: 2, Z: -1}),
		createPiece(model3d.Coord3D{X: -3, Y: -2, Z: -5},
			model3d.Coord3D{X: -2.8, Y: 2, Z: -1.2}),
		createPiece(model3d.Coord3D{X: 2.8, Y: -2, Z: -5},
			model3d.Coord3D{X: 3, Y: 2, Z: -1.2}),
	}
}

func CreateLight() render3d.Object {
	return &render3d.Sphere{
		Center: model3d.Coord3D{Z: 6},
		Radius: 1.3,
		Material: &render3d.LambertMaterial{
			EmissionColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(200),
		},
	}
}
