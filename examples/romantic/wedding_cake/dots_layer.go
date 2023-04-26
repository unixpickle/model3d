package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	DotsLayerThickness     = 1.0
	DotsLayerRadius        = 1.5
	DotsLayerRepeats       = 10.0
	DotsLayerDotsPerRepeat = 8.0
	DotsLayerDotRadius     = 0.03
)

func DotsLayer() (model3d.Solid, toolbox3d.CoordColorFunc) {
	var points []model3d.Coord3D
	dotSpace := math.Pi * 2 * DotsLayerRadius / (DotsLayerDotsPerRepeat * DotsLayerRepeats)
	for i := 0; i < DotsLayerDotsPerRepeat*DotsLayerRepeats; i++ {
		theta := math.Pi * 2 * float64(i) / float64(DotsLayerDotsPerRepeat*DotsLayerRepeats)
		frac := float64(i%DotsLayerDotsPerRepeat) / DotsLayerDotsPerRepeat
		down := 0.1 * math.Sin(frac*math.Pi)
		x := math.Cos(theta) * DotsLayerRadius
		y := math.Sin(theta) * DotsLayerRadius
		points = append(
			points,
			model3d.XYZ(x, y, DotsLayerThickness*0.9-down),
			model3d.XYZ(x, y, DotsLayerThickness*0.9-down-dotSpace),
		)
		if frac == 0 {
			topZ := points[len(points)-1].Z
			for z := topZ; z > dotSpace; z -= dotSpace {
				points = append(points, model3d.XYZ(x, y, z))
			}
		}
	}
	tree := model3d.NewCoordTree(points)
	r := DotsLayerRadius + DotsLayerDotRadius
	return model3d.CheckedFuncSolid(
			model3d.XYZ(-r, -r, 0),
			model3d.XYZ(r, r, DotsLayerThickness),
			func(c model3d.Coord3D) bool {
				return c.XY().Norm() < DotsLayerRadius || tree.Dist(c) < DotsLayerDotRadius
			},
		), func(c model3d.Coord3D) render3d.Color {
			if tree.Dist(c) < DotsLayerDotRadius {
				return GoldDripColor
			} else {
				return render3d.NewColor(1)
			}
		}
}
