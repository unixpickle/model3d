package main

import (
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	// Create two balls
	balls := model3d.NewMesh()
	balls.AddMesh(model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return 1
	}, 100).MapCoords(model3d.Coord3D{X: 1.5}.Add))
	balls.AddMesh(model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return 1
	}, 100).MapCoords(model3d.Coord3D{X: -1.5}.Add))

	// Create a floor.
	floor := model3d.SolidToMesh(&model3d.RectSolid{
		MinVal: model3d.Coord3D{X: -5, Y: -2, Z: -2},
		MaxVal: model3d.Coord3D{X: 5, Y: 5, Z: -1.8},
	}, 0.05, 0, 0, 0).EliminateCoplanar(1e-8)

	// Join all the objects into a mega-object.
	object := render3d.JoinedObject{
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(balls),
			Material: &render3d.PhongMaterial{
				Alpha:         5.0,
				SpecularColor: render3d.Color{X: 0.8, Y: 0.8, Z: 0.8},
				DiffuseColor:  render3d.Color{X: 0.7, Y: 0, Z: 0},
				AmbienceColor: render3d.Color{X: 0.3},
			},
		},
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(floor),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.Color{X: 0.3, Y: 1, Z: 0.6},
			},
		},
	}

	renderer := render3d.RayCaster{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -10, Z: 2}, model3d.Coord3D{}, 0),
		Lights: []*render3d.PointLight{
			&render3d.PointLight{
				Color:  render3d.Color{X: 1, Y: 1, Z: 1},
				Origin: model3d.Coord3D{X: 20, Y: -20, Z: 20},
			},
		},
	}
	img := render3d.NewImage(500, 500)
	renderer.Render(img, object)
	img.Save("output.png")
}
