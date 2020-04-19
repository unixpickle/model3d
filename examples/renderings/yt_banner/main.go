package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

var LightPoint = model3d.Coord3D{X: 0, Y: -40, Z: 30}

const (
	LettersWidth     = 6.0
	LettersThickness = 0.3
	LettersRounding  = 0.05

	LightRadius   = 5.0
	LightEmission = 170.0

	SkyRadius = 50.0

	CameraY = -9.0
	CameraZ = 2.0
)

func main() {
	scene := render3d.JoinedObject{
		CreateFloor(),
		CreateSky(),
		CreateLetters(),
		CreateLight(),
		CreateCorgi(math.Pi/4, 3.4, -0.5),
		CreateCorgi(math.Pi*3.0/4.0, -3.4, -0.5),
	}
	RenderScene(scene)
}

func RenderScene(scene render3d.Object) {
	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: CameraY, Z: CameraZ},
			model3d.Coord3D{Y: 10 + CameraY, Z: CameraZ}, math.Pi/3.6),

		FocusPoints: []render3d.FocusPoint{
			&render3d.SphereFocusPoint{
				Center: LightPoint,
				Radius: LightRadius,
			},
		},
		FocusPointProbs: []float64{0.3},

		MaxDepth:   10,
		NumSamples: 200,
		Antialias:  1.0,
		Cutoff:     1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	fmt.Println("Variance:", renderer.RayVariance(scene, 200, 133, 2))

	img := render3d.NewImage(480, 320)
	renderer.Render(img, scene)
	fmt.Println()
	img.Save("output.png")
}

func CreateSky() render3d.Object {
	return &render3d.ColliderObject{
		Collider: &model3d.Sphere{
			Radius: SkyRadius,
		},
		Material: &render3d.LambertMaterial{
			AmbientColor:  render3d.NewColorRGB(0.52, 0.8, 0.92),
			EmissionColor: render3d.NewColor(0.1),
		},
	}
}

func CreateLight() render3d.Object {
	return &render3d.ColliderObject{
		Collider: &model3d.Sphere{
			Center: LightPoint,
			Radius: LightRadius,
		},
		Material: &render3d.LambertMaterial{
			EmissionColor: render3d.NewColor(LightEmission),
		},
	}
}
