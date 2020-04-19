package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
)

func main() {
	scene := render3d.JoinedObject{
		NewFloorObject(),
		NewDomeObject(),
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
			&render3d.PhongFocusPoint{
				Target: LightDirection.Scale(RoomRadius),
				Alpha:  20.0,
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
		NumSamples: 100,
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
