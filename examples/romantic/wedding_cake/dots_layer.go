package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	. "github.com/unixpickle/model3d/shorthand"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	DotsLayerThickness     = 1.0
	DotsLayerRadius        = 1.5
	DotsLayerRepeats       = 10.0
	DotsLayerDotsPerRepeat = 8.0
	DotsLayerDotRadius     = 0.03
)

func DotsLayer() (Solid3, toolbox3d.CoordColorFunc) {
	var points []C3
	dotSpace := math.Pi * 2 * DotsLayerRadius / (DotsLayerDotsPerRepeat * DotsLayerRepeats)
	for i := 0; i < DotsLayerDotsPerRepeat*DotsLayerRepeats; i++ {
		theta := math.Pi * 2 * float64(i) / float64(DotsLayerDotsPerRepeat*DotsLayerRepeats)
		frac := float64(i%DotsLayerDotsPerRepeat) / DotsLayerDotsPerRepeat
		down := 0.1 * math.Sin(frac*math.Pi)
		x := math.Cos(theta) * DotsLayerRadius
		y := math.Sin(theta) * DotsLayerRadius
		points = append(
			points,
			XYZ(x, y, DotsLayerThickness*0.9-down),
			XYZ(x, y, DotsLayerThickness*0.9-down-dotSpace),
		)
		if frac == 0 {
			topZ := points[len(points)-1].Z
			for z := topZ; z > dotSpace; z -= dotSpace {
				points = append(points, XYZ(x, y, z))
			}
		}
	}
	tree := model3d.NewCoordTree(points)
	r := DotsLayerRadius + DotsLayerDotRadius
	return MakeSolid3(
			XYZ(-r, -r, 0),
			XYZ(r, r, DotsLayerThickness),
			func(c C3) bool {
				return c.XY().Norm() < DotsLayerRadius || tree.Dist(c) < DotsLayerDotRadius
			},
		), func(c C3) Color {
			if tree.Dist(c) < DotsLayerDotRadius {
				return GoldDripColor
			} else {
				return Gray(1)
			}
		}
}
