package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func Eyes() (model3d.Solid, toolbox3d.CoordColorFunc) {
	ball := model3d.TranslateSolid(
		model3d.RotateSolid(
			model3d.VecScaleSolid(
				&model3d.Sphere{Radius: 0.3},
				model3d.XYZ(1.0, 0.3, 1.0),
			),
			model3d.X(1),
			0.2,
		),
		model3d.YZ(0.6, 1.7),
	)
	combined := model3d.JoinedSolid{
		model3d.TranslateSolid(
			model3d.RotateSolid(ball, model3d.Z(1), 0.3),
			model3d.X(-0.2),
		),
		model3d.TranslateSolid(
			model3d.RotateSolid(ball, model3d.Z(1), -0.3),
			model3d.X(0.2),
		),
	}
	return combined, func(c model3d.Coord3D) render3d.Color {
		c.X = math.Min(math.Abs(c.X-0.5), math.Abs(c.X+0.3))
		dist := c.XZ().Dist(model2d.Y(1.7))
		if dist < 0.1 {
			return render3d.NewColor(0)
		} else {
			return render3d.NewColor(1.0)
		}
	}
}
