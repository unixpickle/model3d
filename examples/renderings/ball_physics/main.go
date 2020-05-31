package main

import (
	"fmt"
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	log.Println("Creating scene...")
	scene := NewScene()

	log.Println("Rendering...")

	renderer := &render3d.BidirPathTracer{
		Camera: render3d.NewCameraAt(model3d.Coord3D{Y: -RoomDepth + 0.1, Z: RoomHeight / 3},
			model3d.Coord3D{Y: 0, Z: RoomHeight / 3}, math.Pi/3.6),

		MaxDepth:      15,
		MinDepth:      3,
		NumSamples:    40,
		RouletteDelta: 0.2,
		Antialias:     1.0,
		Cutoff:        1e-4,
	}

	for i := 0; i < 20; i++ {
		log.Println("Stepping frame", i, "...")
		scene.Step()

		log.Println("Rendering frame", i, "...")
		sceneObj, light := scene.Scene()
		renderer.Light = light
		img := render3d.NewImage(100, 100)
		renderer.Render(img, sceneObj)
		img.Save(fmt.Sprintf("scene_%03d.png", i))
	}
}
