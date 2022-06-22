package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateMugContents() model3d.Solid {
	radiusFunc := model2d.BezierCurve{
		model2d.XY(1.1, 1.0),
		model2d.XY(-0.5, 1.0),
		model2d.XY(-1.0, 1.0),
		model2d.XY(-1.0, 0.0),
	}
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1.0, -1.0, -0.91),
		model3d.XYZ(1.0, 1.0, 0.3),
		func(c model3d.Coord3D) bool {
			return c.XY().Norm() < radiusFunc.EvalX(c.Z)-(Thickness/2-0.01)
		},
	)
}

func CreateMug() model3d.Solid {
	return model3d.JoinedSolid{createMugBody(), createHandle()}
}

func createMugBody() model3d.Solid {
	side := createHoleSide()
	radiusFunc := model2d.BezierCurve{
		model2d.XY(1.1, 1.0),
		model2d.XY(-0.5, 1.0),
		model2d.XY(-1.0, 1.0),
		model2d.XY(-1.0, 0.0),
	}
	polys := radiusFunc.Polynomials()

	// These derivatives can tell us the slope of the cup
	// surface, which tells us how much to stretch the thickness.
	dx := polys[0].Derivative()
	dy := polys[1].Derivative()
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1.2, -1.2, -1.0),
		model3d.XYZ(1.2, 1.2, 1.1),
		func(c model3d.Coord3D) bool {
			r := c.XY().Norm()
			if c.Z < -0.95 {
				return r < 0.6
			} else if r < 1e-5 {
				return c.Z < -0.9
			}
			t := radiusFunc.InverseX(c.Z)
			actualRad := radiusFunc.Eval(t).Y
			dydx := math.Abs(dy.Eval(t) / math.Max(math.Abs(dx.Eval(t)), 0.0001))
			theta := math.Atan2(c.Y, c.X)
			thicknessScale := math.Sqrt(1 + dydx*dydx)
			return side.Contains(model3d.XYZ(theta, (actualRad-r)/thicknessScale, c.Z))
		},
	)
}

func createHandle() model3d.Solid {
	torus := &model3d.Torus{
		Center:      model3d.XYZ(-1.1, 0.0, 0.2),
		Axis:        model3d.Y(1),
		InnerRadius: 0.1,
		OuterRadius: 0.4,
	}
	return model3d.TransformSolid(
		&model3d.Matrix3Transform{
			Matrix: &model3d.Matrix3{1.0, 0.0, 0.0, 0.0, 2.0, 0.0, 0.0, 0.0, 1.0},
		},
		toolbox3d.ClampAxisMax(torus, toolbox3d.AxisX, -0.95),
	)
}

func createHoleSide() model3d.Solid {
	baseObj := model3d.NewRect(
		model3d.XYZ(-math.Pi, -Thickness/2, -1.0),
		model3d.XYZ(math.Pi, Thickness/2, 1.0),
	)
	holeBounds := *baseObj
	holeBounds.MinVal.Z = -0.7
	holeBounds.MinVal.X += 0.3
	holeBounds.MaxVal.X -= 0.3
	axis := model3d.Y(1)

	topLipBezier := model2d.BezierCurve{
		model2d.XY(0.9, Thickness/2),
		model2d.XY(1.0, Thickness/2),
		model2d.XY(1.05, Thickness*1.5),
		model2d.XY(1.1, Thickness*0.25),
		model2d.XY(1.1, 0),
	}
	topLip := model3d.CheckedFuncSolid(
		model3d.XYZ(-math.Pi, -Thickness*1.5, 0.9),
		model3d.XYZ(math.Pi, Thickness*1.5, 1.1),
		func(c model3d.Coord3D) bool {
			return math.Abs(c.Y) < topLipBezier.EvalX(c.Z)
		},
	)

	return Holeify(model3d.JoinedSolid{baseObj, topLip}, &holeBounds, axis)
}
