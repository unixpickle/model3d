package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreatePlant() (model3d.Solid, toolbox3d.CoordColorFunc) {
	base := toolbox3d.ClampAxisMin(&model3d.Cone{
		Tip:    model3d.XZ(-2.2, -5.0),
		Base:   model3d.XZ(-2.2, 1.6),
		Radius: 0.8,
	}, toolbox3d.AxisZ, 0)
	biggerBase := toolbox3d.ClampAxisMin(&model3d.Cone{
		Tip:    model3d.XZ(-2.2, -5.0),
		Base:   model3d.XZ(-2.2, 1.6),
		Radius: 0.9,
	}, toolbox3d.AxisZ, 1.3)
	vase := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{base, biggerBase},
		Negative: toolbox3d.ClampAxisMin(&model3d.Cone{
			Tip:    model3d.XZ(-2.2, -5.0),
			Base:   model3d.XZ(-2.2, 1.6001),
			Radius: 0.8,
		}, toolbox3d.AxisZ, 1.4),
	}
	branch := CreateBranch()
	leaves := model3d.TranslateSolid(model3d.JoinedSolid{
		branch,
		model3d.TranslateSolid(model3d.RotateSolid(branch, model3d.Y(1), -0.4), model3d.XYZ(-0.3, 0.15, -0.1)),
		model3d.TranslateSolid(model3d.RotateSolid(branch, model3d.Y(1), 0.4), model3d.XYZ(0.3, 0.15, -0.05)),
		model3d.TranslateSolid(model3d.RotateSolid(branch, model3d.Y(1), -0.2), model3d.XYZ(-0.2, -0.15, -0.3)),
		model3d.TranslateSolid(model3d.RotateSolid(branch, model3d.Y(1), 0.2), model3d.XYZ(0.2, -0.15, -0.3)),
	}, model3d.XZ(-2.2, 1.0))
	leafSDF := model3d.MeshToSDF(model3d.MarchingCubesSearch(leaves, 0.02, 8))
	colorFn := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if leafSDF.SDF(c) > -0.01 {
			return render3d.NewColorRGB(0.0, 1.0, 0.0)
		}
		return render3d.NewColorRGB(0.85, 0.46, 0.24)
	})
	return model3d.JoinedSolid{vase, leaves}, colorFn
}

func CreateBranch() model3d.Solid {
	solid2d := model2d.CheckedFuncSolid(
		model2d.XY(-0.5, 0),
		model2d.XY(0.5, 2.0),
		func(c model2d.Coord) bool {
			radius := 0.3*math.Pow(math.Sin(math.Pi*c.Y), 2) +
				0.1*math.Pow(math.Sin(math.Pi*5*c.Y), 2)
			return math.Abs(c.X) < radius
		},
	)
	mesh2d := model2d.MarchingSquaresSearch(solid2d, 0.01, 8)
	profile := model3d.ProfileCollider(model2d.MeshToCollider(mesh2d), -0.01, 0.01)
	solid3d := model3d.RotateSolid(
		model3d.NewColliderSolidHollow(profile, 0.1),
		model3d.X(1),
		math.Pi/2,
	)
	stem := toolbox3d.RadialCurve(1000, false, func(t float64) (model3d.Coord3D, float64) {
		c := model3d.Z(t * 2.9)
		if t < 0.8 {
			return c, 0.15
		}
		return c, 0.15 * math.Cos(0.5*math.Pi*(t-0.8)/0.2)
	})
	return model3d.JoinedSolid{
		model3d.TranslateSolid(solid3d, model3d.Z(1.0)),
		stem,
	}
}
