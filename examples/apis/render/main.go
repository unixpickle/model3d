package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	// Join all the objects into a mega-object.
	object := render3d.JoinedObject{
		// Red ball.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(model3d.NewMeshPolar(
				func(g model3d.GeoCoord) float64 {
					return 1
				}, 100).MapCoords(model3d.Coord3D{X: 2, Y: 6, Z: 0}.Add)),
			Material: &render3d.PhongMaterial{
				Alpha:         5.0,
				SpecularColor: render3d.Color{X: 0.8, Y: 0.8, Z: 0.8},
				DiffuseColor:  render3d.Color{X: 0.7, Y: 0.1, Z: 0.1},
				AmbienceColor: render3d.Color{X: 0.3},
			},
		},
		// Blue ball.
		// &render3d.ColliderObject{
		// 	Collider: model3d.MeshToCollider(model3d.NewMeshPolar(
		// 		func(g model3d.GeoCoord) float64 {
		// 			return 1
		// 		}, 100).MapCoords(model3d.Coord3D{X: 1.5, Y: -5}.Add)),
		// 	Material: &render3d.PhongMaterial{
		// 		Alpha:         5.0,
		// 		SpecularColor: render3d.Color{X: 0.8, Y: 0.8, Z: 0.8},
		// 		DiffuseColor:  render3d.Color{X: 0.1, Y: 0.1, Z: 0.7},
		// 		AmbienceColor: render3d.Color{X: 0.3},
		// 	},
		// },
		// Room walls.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.SolidToMesh(
					&model3d.RectSolid{
						MinVal: model3d.Coord3D{X: -5, Y: 0, Z: -2},
						MaxVal: model3d.Coord3D{X: 5, Y: 10, Z: 7},
					},
					0.05, 0, 0, 0,
				).EliminateCoplanar(1e-8).MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: 1}.Mul),
			),
			Material: &render3d.LambertMaterial{
				DiffuseColor:  render3d.Color{X: 0.8, Y: 0.8, Z: 0.8},
				AmbienceColor: render3d.Color{X: 0.3, Y: 0.3, Z: 0.3},
			},
		},
		// Ceiling light.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.SolidToMesh(
					&model3d.RectSolid{
						MinVal: model3d.Coord3D{X: -2, Y: 5, Z: 6.8},
						MaxVal: model3d.Coord3D{X: 2, Y: 7, Z: 7},
					},
					0.05, 0, 0, 0,
				).EliminateCoplanar(1e-8),
			),
			Material: &render3d.LambertMaterial{
				// Make it really bright so it lights the scene
				// adequately.
				LuminanceColor: render3d.Color{X: 1, Y: 1, Z: 1}.Scale(20),
			},
		},
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Z: 3}, model3d.Coord3D{Y: 10, Z: 3}, math.Pi/2),
		Lights: []*render3d.PointLight{
			// &render3d.PointLight{
			// 	Color:  render3d.Color{X: 1, Y: 1, Z: 1},
			// 	Origin: model3d.Coord3D{X: 0, Y: 1, Z: 6},
			// },
		},
		MaxDepth:   2,
		NumSamples: 500,
	}
	img := render3d.NewImage(500, 500)
	renderer.Render(img, object)
	img.Save("output.png")
}
