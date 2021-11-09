package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	radii := []float64{1.0 * 0.7, 0.8 * 0.7, 0.6 * 0.7}
	z := 0.0
	solid := model3d.JoinedSolid{}
	for _, radius := range radii {
		cone := &model3d.Cone{
			Tip:    model3d.Z(z + radius*2),
			Base:   model3d.Z(z),
			Radius: radius,
		}
		solid = append(solid, cone)
		z += radius
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}
