package main

import (
	"fmt"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	// Join all the objects into a mega-object.
	object := render3d.JoinedObject{
		// Mirror ball.
		&render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: model3d.XYZ(2, 7, 0),
				Radius: 2,
			},
			Material: &render3d.PhongMaterial{
				Alpha:         400.0,
				SpecularColor: render3d.NewColor(1),
			},
		},

		// Red ball.
		&render3d.ColliderObject{
			Collider: &model3d.Sphere{
				Center: model3d.XYZ(-2, 5.5, -1),
				Radius: 1,
			},
			Material: &render3d.PhongMaterial{
				Alpha:         10.0,
				SpecularColor: render3d.NewColor(0.1),
				DiffuseColor:  render3d.NewColorRGB(0.95, 0.2, 0.2).Scale(0.5),
			},
		},

		// Glass diamond.
		&render3d.ColliderObject{
			Collider: LoadDiamond(),
			Material: &render3d.JoinedMaterial{
				Materials: []render3d.Material{
					&render3d.RefractMaterial{
						IndexOfRefraction: 1.3,
						RefractColor:      render3d.NewColor(0.9),
					},
					&render3d.PhongMaterial{
						Alpha:         50.0,
						SpecularColor: render3d.NewColor(0.1),
					},
				},
				Probs: []float64{0.9, 0.1},
			},
		},

		// Room walls.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.NewMeshRect(
					model3d.XYZ(-5, -10, -2),
					model3d.XYZ(5, 10, 7),
				).MapCoords(model3d.XYZ(-1, 1, 1).Mul),
			),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.NewColor(0.4),
			},
		},

		// Ceiling light.
		&render3d.ColliderObject{
			Collider: model3d.MeshToCollider(
				model3d.NewMeshRect(
					model3d.XYZ(-2, 5, 6.8),
					model3d.XYZ(2, 7, 7),
				),
			),
			Material: &render3d.LambertMaterial{
				// Make it really bright so it lights the scene
				// adequately.
				EmissionColor: render3d.NewColor(25),
			},
		},
	}

	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -7, Z: 2.5},
			model3d.Coord3D{Y: 10, Z: 2.5}, math.Pi/3.6),

		// Focus reflections towards the light source
		// to lower variance (i.e. grain).
		FocusPoints: []render3d.FocusPoint{
			&render3d.PhongFocusPoint{
				Target: model3d.XYZ(0, 6, 6.9),
				Alpha:  40.0,
				MaterialFilter: func(m render3d.Material) bool {
					if _, ok := m.(*render3d.LambertMaterial); ok {
						return true
					} else if phong, ok := m.(*render3d.PhongMaterial); ok {
						return phong.DiffuseColor.Sum() > 0
					} else {
						// Don't focus sharp materials like refraction
						// and specular-only phong materials.
						return false
					}
				},
			},
		},
		FocusPointProbs: []float64{0.3},

		MaxDepth:   5,
		NumSamples: 400,
		Antialias:  1.0,
		Cutoff:     1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	fmt.Println("Ray variance:", renderer.RayVariance(object, 200, 200, 5))

	img := render3d.NewImage(200, 200)
	renderer.Render(img, object)
	fmt.Println()
	img.Save("output.png")
}

func LoadDiamond() model3d.Collider {
	r, err := os.Open("diamond.stl")
	essentials.Must(err)
	triangles, err := model3d.ReadSTL(r)
	essentials.Must(err)

	mesh := model3d.NewMeshTriangles(triangles)

	// Put the diamond on its side.
	mesh = mesh.Rotate(model3d.Y(1), 0.5*math.Pi+math.Atan(1/1.2))

	translate := model3d.Translate{
		Offset: model3d.Coord3D{Z: -(2 + mesh.Min().Z), Y: 4},
	}
	mesh = mesh.MapCoords(translate.Apply)

	return model3d.MeshToCollider(mesh)
}
