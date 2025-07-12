package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	CherryRadius     = 0.25
	CherryStemLength = 0.15
	CherryStemRadius = 0.05
)

func CherrySolid() (model3d.Solid, toolbox3d.CoordColorFunc) {
	curve := model2d.SmoothBezier(
		model2d.XY(0, -CherryRadius),
		model2d.XY(-CherryRadius*0.7, -CherryRadius),
		model2d.XY(-CherryRadius, -CherryRadius*0.3),
		model2d.XY(-CherryRadius, 0),
		// Point 1
		model2d.XY(-CherryRadius*0.9, CherryRadius*0.5),
		model2d.XY(-CherryRadius*0.6, CherryRadius*0.7),
		// Point 2
		model2d.XY(-CherryRadius*0.4, CherryRadius*0.9),
		model2d.XY(-CherryRadius*0.1, CherryRadius*0.8),
		// Point 3
		model2d.XY(-CherryRadius*0.1, CherryRadius*0.7),
		model2d.XY(0, CherryRadius*0.7),
	)
	profile := model2d.CurveMesh(curve, 100)
	profile.AddMesh(profile.MapCoords(model2d.XY(-1, 1).Mul))
	solid2d := profile.Solid()
	center := CreamHeight + CupHeight
	solid := model3d.CheckedFuncSolid(
		model3d.XYZ(-CherryRadius, -CherryRadius, center-CherryRadius),
		model3d.XYZ(CherryRadius, CherryRadius, center+CherryRadius+0.2),
		func(c model3d.Coord3D) bool {
			x := c.XY().Norm()
			y := c.Z - center
			return solid2d.Contains(model2d.XY(x, y))
		},
	)

	stemCurve := model2d.BezierCurve{
		model2d.XY(0.0, center+CherryRadius-0.03),
		model2d.XY(0.0, center+CherryRadius+CherryStemLength/2),
		model2d.XY(CherryStemLength*0.4, center+CherryRadius+CherryStemLength),
	}
	var segments []model3d.Segment
	eps := 0.01
	for t := 0.0; t < 1.0-eps; t += eps {
		p1 := stemCurve.Eval(t)
		p2 := stemCurve.Eval(t + eps)
		segments = append(
			segments,
			model3d.NewSegment(
				model3d.YZ(p1.X, p1.Y),
				model3d.YZ(p2.X, p2.Y),
			),
		)
	}
	stem := toolbox3d.LineJoin(CherryStemRadius, segments...)

	colorFn := func(c model3d.Coord3D) render3d.Color {
		if stem.Contains(c) {
			return render3d.NewColorRGB(0.2, 0.0, 0.0)
		} else {
			return render3d.NewColorRGB(1.0, 0.0, 0.0)
		}
	}

	return model3d.JoinedSolid{solid, stem}, colorFn
}
