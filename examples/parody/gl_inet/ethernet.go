package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func EthernetJackSolid() model3d.Solid {
	profile := model2d.NewColliderSolid(
		model2d.MeshToCollider(
			model2d.CurveMesh(model2d.JoinedCurve{
				&model2d.BezierCurve{
					model2d.XY(0.15, 0.15),
					model2d.XY(0.85, 0.15),
				},
				&model2d.BezierCurve{
					model2d.XY(0.85, 0.15),
					model2d.XY(0.85, 0.7),
				},
				&model2d.BezierCurve{
					model2d.XY(0.85, 0.7),
					model2d.XY(0.65, 0.7),
				},
				&model2d.BezierCurve{
					model2d.XY(0.65, 0.7),
					model2d.XY(0.65, 0.85),
				},
				&model2d.BezierCurve{
					model2d.XY(0.65, 0.85),
					model2d.XY(0.35, 0.85),
				},
				&model2d.BezierCurve{
					model2d.XY(0.35, 0.85),
					model2d.XY(0.35, 0.7),
				},
				&model2d.BezierCurve{
					model2d.XY(0.35, 0.7),
					model2d.XY(0.15, 0.7),
				},
				&model2d.BezierCurve{
					model2d.XY(0.1, 0.7),
					model2d.XY(0.1, 0.1),
				},
			}, 300).MapCoords(model2d.XY(1.0, -1.0).Mul).Translate(model2d.Y(1.0)),
		),
	)
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-0.9, -BodySideLength/2+0.02, 0.1),
		model3d.XYZ(0.3, -BodySideLength/2+0.3, 0.6),
		func(c model3d.Coord3D) bool {
			x := (c.X + 0.9) / (0.9 + 0.3)
			if x > 0.5 {
				x -= 0.5
			}
			x *= 2.0
			y := (c.Z - 0.1) / 0.5
			return !profile.Contains(model2d.XY(x, y))
		},
	)
}
