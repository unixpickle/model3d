package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const HexThickness = 0.8

func MarbleHexagon() (model3d.Solid, toolbox3d.CoordColorFunc) {
	hexMesh := model2d.NewMeshPolar(func(theta float64) float64 {
		return 1.0
	}, 6)
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(hexMesh))
	solid3d := model3d.ProfileSolid(solid2d, 0.0, HexThickness)

	streaks := []*MarbleStreak{
		{
			Theta:    0.0,
			Power:    3.0,
			Darkness: 0.28,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.05),
				model2d.XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.5, 0.3),
				model2d.XY(1.0, -0.3),
			},
		},
		{
			Theta:    math.Pi * 0.7,
			Power:    2.0,
			Darkness: 0.34,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.02),
				model2d.XY(1.0, 0.08),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.25, 0.1),
				model2d.XY(0.6, -0.3),
				model2d.XY(1.0, -0.1),
			},
		},
		{
			Theta:    math.Pi * 1.5,
			Power:    2.0,
			Darkness: 0.25,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.05),
				model2d.XY(0.25, 0.04),
				model2d.XY(1.0, 0.045),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.2, 0.1),
				model2d.XY(0.5, -0.1),
				model2d.XY(0.7, -0.25),
				model2d.XY(1.0, -0.3),
			},
		},
		{
			Theta:    math.Pi * 0.3,
			Power:    3.0,
			Darkness: 0.15,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.08),
				model2d.XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.3, 0.1),
				model2d.XY(0.6, -0.1),
				model2d.XY(1.0, 0.2),
			},
		},
		{
			Theta:    math.Pi * 1.2,
			Power:    3.0,
			Darkness: 0.3,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.08),
				model2d.XY(0.8, 0.05),
				model2d.XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.3, -0.2),
				model2d.XY(0.6, -0.4),
				model2d.XY(1.0, -0.5),
			},
		},
		{
			Theta:    math.Pi * 1.8,
			Power:    2.0,
			Darkness: 0.4,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.08),
				model2d.XY(0.5, 0.09),
				model2d.XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.3, 0.1),
				model2d.XY(0.6, -0.1),
				model2d.XY(1.0, 0.2),
			},
		},
		{
			Theta:    math.Pi * 0.5,
			Power:    2.0,
			Darkness: 0.2,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.1),
				model2d.XY(0.5, 0.11),
				model2d.XY(1.0, 0.08),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.2, 0.1),
				model2d.XY(0.5, -0.2),
				model2d.XY(1.0, -0.3),
			},
		},
		{
			Theta:    math.Pi * 1.2,
			Power:    2.0,
			Darkness: 0.2,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.1),
				model2d.XY(1.0, 0.08),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.5, 0.05),
				model2d.XY(1.0, 0.2),
			},
		},
		{
			Theta:    math.Pi * 0.6,
			Power:    2.5,
			Darkness: 0.2,
			Thickness: model2d.BezierCurve{
				model2d.XY(0.0, 0.05),
				model2d.XY(0.5, 0.06),
				model2d.XY(1.0, 0.06),
			},
			Curve: model2d.BezierCurve{
				model2d.XY(0.0, 0.0),
				model2d.XY(0.5, 0.05),
				model2d.XY(1.0, -0.2),
			},
		},
	}

	return solid3d, func(c model3d.Coord3D) render3d.Color {
		maxStreak := 0.0
		for _, s := range streaks {
			maxStreak = math.Max(maxStreak, s.Evaluate(c))
		}
		return render3d.NewColor(1.0).AddScalar(-maxStreak * 0.5)
	}
}

type MarbleStreak struct {
	Theta     float64
	Power     float64
	Darkness  float64
	Thickness model2d.BezierCurve
	Curve     model2d.BezierCurve
}

func (m *MarbleStreak) Evaluate(c model3d.Coord3D) float64 {
	xAxis := model3d.XYZ(-math.Sin(m.Theta), math.Cos(m.Theta), 0)
	xDist := c.Dot(xAxis)
	zFrac := math.Max(0.0, math.Min(1.0, c.Z/HexThickness))
	offset := m.Curve.EvalX(zFrac)
	thickness := m.Thickness.EvalX(zFrac)
	distFrac := math.Abs(offset-xDist) / thickness
	if distFrac > 1 {
		return 0.0
	}
	return m.Darkness * (1 - math.Pow(distFrac, m.Power))
}
