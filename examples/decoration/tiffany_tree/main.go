package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	solid := model3d.JoinedSolid{}
	for _, cone := range Cones() {
		solid = append(solid, cone)
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	color := ColorFunc()
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, color.RenderColor)
}

func Cones() []*model3d.Cone {
	radii := []float64{1.0 * 0.7, 0.8 * 0.7, 0.6 * 0.7}
	z := 0.0
	cones := []*model3d.Cone{}
	for _, radius := range radii {
		cone := &model3d.Cone{
			Tip:    model3d.Z(z + radius*2),
			Base:   model3d.Z(z),
			Radius: radius,
		}
		cones = append(cones, cone)
		z += radius
	}
	return cones
}

func ColorFunc() toolbox3d.CoordColorFunc {
	return func(c model3d.Coord3D) render3d.Color {
		loopCounts := []int{8, 6, 5}
		theta := math.Atan2(c.Y, c.X) + math.Pi
		for i, cone := range Cones() {
			z := cone.Base.Z + 0.25
			zOffset := -0.1 * math.Sqrt(math.Abs(math.Cos(theta*float64(loopCounts[i])/2)))
			thickness := 0.05
			if math.Abs(z+zOffset-c.Z) < thickness/2 || math.Abs(cone.Base.Z-c.Z) < thickness {
				return render3d.NewColor(1.0)
			}
		}
		return render3d.NewColorRGB(0x66/255.0, 0xf3/255.0, 0xed/255.0)
	}
}
