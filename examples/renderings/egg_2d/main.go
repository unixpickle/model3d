package main

import (
	"image/color"
	"math"

	"github.com/unixpickle/model3d/model2d"
)

func main() {
	// Create a wiggly egg.
	egg := model2d.NewMeshPolar(func(theta float64) float64 {
		return 0.9 + 0.1*math.Sin(theta*5)
	}, 300)
	circle := &model2d.Circle{Radius: 0.3}
	objs := []any{
		egg.Solid(),
		circle,
	}
	model2d.RasterizeColor("egg.png", objs, []color.Color{
		color.Gray{Y: 0xff},
		color.RGBA{R: 0xff, G: 0xff, A: 0xff},
	}, 200.0)
}
