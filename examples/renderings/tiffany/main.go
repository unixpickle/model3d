package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	scene := render3d.JoinedObject{
		NewGlobe(),
		NewWalls(),
		CreateTable(),
		CreateLamp(),
		CreateWallLights(),
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -13, Z: 2},
			model3d.Coord3D{Y: 0, Z: 2}, math.Pi/3.6),

		MaxDepth: 5,

		NumSamples:           200,
		MinSamples:           200,
		MaxStddev:            0.05,
		OversaturatedStddevs: 3,

		Antialias: 1.0,
		Cutoff:    1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	ceilingLights := NewWalls().Lights
	for _, l := range ceilingLights {
		point := l.Cylinder.P1
		renderer.FocusPoints = append(renderer.FocusPoints, &render3d.PhongFocusPoint{
			Alpha:  50,
			Target: point,
		})
		renderer.FocusPointProbs = append(renderer.FocusPointProbs,
			0.3/float64(len(ceilingLights)))
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
				SpecularColor: render3d.NewColor(0.2),
				DiffuseColor:  render3d.NewColor(0.4),
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
