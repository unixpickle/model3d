package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func HeartOutline() *model2d.Mesh {
	curve := model2d.BezierCurve{
		model2d.XY(0.0, 0.2),
		model2d.XY(0.3, 1.2),
		model2d.XY(1.5, 0.8),
		model2d.XY(1.0, -0.5),
		model2d.XY(0.5, -1.0),
		model2d.XY(0.0, -1.0),
	}
	bothCurve := model2d.FuncCurve(func(t float64) model2d.Coord {
		if t < 0.5 {
			return curve.Eval(2 * t)
		} else {
			return curve.Eval(2 * (1 - t)).Mul(model2d.XY(-1, 1))
		}
	})
	min, max := (&toolbox3d.LineSearch{Stops: 50, Recursions: 5}).CurveBounds(0, 1, curve.Eval)
	mesh := model2d.CurveMesh(bothCurve, 300)
	mesh = mesh.Scale(args.HeartWidth / max.Sub(min).X)
	mesh = mesh.Center()
	return mesh
}
