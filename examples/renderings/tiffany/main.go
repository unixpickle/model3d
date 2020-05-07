package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

// ErrorMargin controls the amount of allowed noise in an
// HD rendering of the scene.
const ErrorMargin = 0.01

func main() {
	ceilingLights := CreateCeilingLights()
	lamp := CreateLamp()
	wallLights := CreateWallLights()

	scene := render3d.JoinedObject{
		NewGlobe(),
		NewWalls(ceilingLights),
		CreateTable(),
		lamp,
		wallLights,
	}

	allLights := make([]render3d.AreaLight, 0, len(ceilingLights)+2)
	for _, x := range ceilingLights {
		allLights = append(allLights, x)
	}
	allLights = append(allLights, lamp, wallLights)

	renderer := render3d.BidirPathTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -13, Z: 2},
			model3d.Coord3D{Y: 0, Z: 2}, math.Pi/3.6),
		Light: render3d.JoinAreaLights(allLights...),

		MaxDepth: 15,
		MinDepth: 3,

		NumSamples: 200,
		MinSamples: 200,

		// Gamma-aware convergence criterion.
		Convergence: func(mean, stddev render3d.Color) bool {
			stddevs := stddev.Array()
			for i, m := range mean.Array() {
				s := stddevs[i]
				if m-3*s > 1 {
					// Oversaturated, so even if the variance
					// is high, this region is stable.
					continue
				}
				if math.Pow(m+s, 1/2.2)-math.Pow(m, 1/2.2) > ErrorMargin {
					return false
				}
			}
			return true
		},

		RouletteDelta: 0.2,

		Antialias: 1.0,
		Cutoff:    1e-4,

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
