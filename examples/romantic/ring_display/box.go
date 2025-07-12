package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func RingBox() (model3d.Solid, toolbox3d.CoordColorFunc) {
	box := model3d.NewRect(model3d.XYZ(-0.9, -0.9, -0.5), model3d.XYZ(0.9, 0.9, 0.5))
	outer := model3d.NewColliderSolidInset(box, -0.2)
	inner := model3d.NewColliderSolidInset(box, -0.1)
	fullBox := model3d.Subtract(outer, inner)

	boxBottom := toolbox3d.ClampAxis(fullBox, toolbox3d.AxisZ, math.Inf(-1), 0.0)
	boxTop := toolbox3d.ClampAxis(fullBox, toolbox3d.AxisZ, 0.0, math.Inf(1))

	rotateOrigin := model3d.Y(-1.0)
	boxTop = model3d.TranslateSolid(boxTop, rotateOrigin.Scale(-1))
	boxTop = model3d.RotateSolid(boxTop, model3d.X(1), math.Pi/2)
	boxTop = model3d.TranslateSolid(boxTop, rotateOrigin)

	hinge := &model3d.Cylinder{
		P1:     model3d.XYZ(-0.5, -1.175, -0.175),
		P2:     model3d.XYZ(0.5, -1.175, -0.175),
		Radius: 0.15,
	}

	feltBottom := model3d.JoinedSolid{
		toolbox3d.ClampAxis(inner, toolbox3d.AxisZ, math.Inf(-1), -0.1),
		model3d.NewColliderSolidInset(
			model3d.NewRect(model3d.XYZ(-0.9, -0.9, -0.5), model3d.XYZ(0.9, 0.9, -0.1)),
			-0.1,
		),
	}

	boxBody := model3d.JoinedSolid{
		boxBottom,
		feltBottom,
		boxTop,
		hinge,
	}
	colorFn := func(c model3d.Coord3D) render3d.Color {
		if hinge.SDF(c) > -0.005 {
			return render3d.NewColorRGB(0.5, 0.4, 0.4)
		} else {
			return render3d.NewColorRGB(0x81/255.0, 0xD8/255.0, 0xD0/255.0)
		}
	}

	return boxBody, colorFn
}
