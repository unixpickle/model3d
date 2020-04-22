package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const HighRes = false

func main() {
	scene := render3d.JoinedObject{
		NewFloorObject(),
		NewDomeObject(),
		NewLightObject(),
		ReadVase(),
		ReadRose(),
		ReadWineGlass(),
		ReadPumpkin(),
		ReadRocks(),
		ReadCurvyThing(),
	}

	RenderScene(scene)
}

func RenderScene(scene render3d.Object) {
	renderer := render3d.RecursiveRayTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: CameraY, Z: CameraZ},
			model3d.Coord3D{Y: RoomRadius, Z: CameraZ}, math.Pi/3.6),

		FocusPoints: []render3d.FocusPoint{
			&render3d.SphereFocusPoint{
				Center: LightCenter,
				Radius: LightRadius,
				MaterialFilter: func(m render3d.Material) bool {
					switch m.(type) {
					case *render3d.PhongMaterial:
						return true
					case *render3d.LambertMaterial:
						return true
					}
					return false
				},
			},
		},
		FocusPointProbs: []float64{0.3},

		MaxDepth:   10,
		NumSamples: 50,
		Antialias:  1.0,
		Cutoff:     1e-4,

		LogFunc: func(p, samples float64) {
			fmt.Printf("\rRendering %.1f%%...", p*100)
		},
	}

	width := 480
	height := 320
	if HighRes {
		width *= 2
		height *= 2
		renderer.NumSamples = 100000
		renderer.MinSamples = 1000
		renderer.MaxStddev = 0.02
	}

	fmt.Println("Variance:", renderer.RayVariance(scene, 200, 133, 2))

	img := render3d.NewImage(width, height)
	renderer.Render(img, scene)
	fmt.Println()
	img.Save("output.png")
}
