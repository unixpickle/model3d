package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	// Join all the objects into a mega-object.
	object := render3d.JoinedObject{
		// Mirror ball.
		&render3d.Sphere{
			Center: model3d.Coord3D{X: 2, Y: 7, Z: 0},
			Radius: 2,
			Material: &render3d.PhongMaterial{
				Alpha:         20.0,
				SpecularColor: render3d.Color{X: 1, Y: 1, Z: 1},
			},
		},

		// Red ball.
		&render3d.Sphere{
			Center: model3d.Coord3D{X: -2, Y: 5.5, Z: -1},
			Radius: 1,
			Material: &render3d.PhongMaterial{
				Alpha:         10.0,
				SpecularColor: render3d.Color{X: 0.3, Y: 0.3, Z: 0.3},
				DiffuseColor:  render3d.Color{X: 0.7, Y: 0.1, Z: 0.1},
			},
		},

		// Room walls.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.NewMeshRect(
					model3d.Coord3D{X: -5, Y: -7, Z: -2},
					model3d.Coord3D{X: 5, Y: 10, Z: 7},
				).MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: 1}.Mul),
			),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.Color{X: 0.8, Y: 0.8, Z: 0.8},
			},
		},

		// Ceiling light.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.NewMeshRect(
					model3d.Coord3D{X: -2, Y: 5, Z: 6.8},
					model3d.Coord3D{X: 2, Y: 7, Z: 7},
				),
			),
			Material: &render3d.LambertMaterial{
				// Make it really bright so it lights the scene
				// adequately.
				EmissionColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(50),
			},
		},
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -6, Z: 2.5},
			model3d.Coord3D{Y: 10, Z: 2.5}, math.Pi/3.6),

		// Focus reflections towards the light source
		// to lower variance (i.e. grain).
		FocusPoints: []render3d.FocusPoint{
			&render3d.PhongFocusPoint{
				Target: model3d.Coord3D{X: 0, Y: 6, Z: 6.9},
				Alpha:  10.0,
			},
		},
		FocusPointProbs: []float64{0.5},

		MaxDepth:   5,
		NumSamples: 400,
	}

	img := render3d.NewImage(200, 200)
	renderer.Render(img, object)
	img.Save("output.png")
}
