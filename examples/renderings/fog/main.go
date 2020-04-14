package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	bounds := &model3d.Rect{
		MinVal: model3d.Coord3D{X: -20, Y: -20, Z: -3},
		MaxVal: model3d.Coord3D{X: 20, Y: 30, Z: 10},
	}
	scene := render3d.JoinedObject{
		&render3d.ParticipatingMedium{
			Collider: bounds,
			Material: &render3d.HGMaterial{
				G:            0.5,
				ScatterColor: render3d.NewColor(0.9),
			},
			Lambda: 0.01,
		},

		// A room to surround the scene.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.NewMeshRect(
					bounds.MinVal,
					bounds.MaxVal,
				).MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: 1}.Mul),
			),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.NewColor(0.45),
			},
		},

		// A ceiling light.
		&render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: model3d.Coord3D{Z: bounds.MaxVal.Z + 1.0, Y: 7},
				Radius: 2.0,
			},
			Material: &render3d.LambertMaterial{
				EmissionColor: render3d.NewColor(100.0),
			},
		},
	}

	// Some spheres to look at.
	for i := 0; i < 20; i++ {
		scene = append(scene, &render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: model3d.NewCoord3DRandNorm().Mul(
					model3d.Coord3D{X: 5, Y: 10, Z: 1},
				).Add(model3d.Coord3D{Y: 5}),
				Radius: 1,
			},
			Material: &render3d.PhongMaterial{
				Alpha:         10.0,
				SpecularColor: render3d.NewColor(0.1),
				DiffuseColor: render3d.NewColorRGB(rand.Float64(), rand.Float64(),
					rand.Float64()).Scale(0.45),
			},
		})
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -13, Z: 2},
			model3d.Coord3D{Y: 0, Z: 2}, math.Pi/3.6),

		FocusPoints: []render3d.FocusPoint{
			&render3d.PhongFocusPoint{
				Target: model3d.Coord3D{Z: bounds.MaxVal.Z, Y: 7},
				Alpha:  50.0,
			},
		},
		FocusPointProbs: []float64{0.3},

		MaxDepth:   10,
		NumSamples: 1000,
		Antialias:  1.0,
		Cutoff:     1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	img := render3d.NewImage(200, 200)
	renderer.Render(img, scene)
	fmt.Println()
	img.Save("output.png")
}
