package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	StrawHeight    = 1.5
	StrawRadius    = 0.1
	StrawInset     = 0.03
	StrawTwistRate = 5.0
)

func StrawSolid() (model3d.Solid, toolbox3d.CoordColorFunc) {
	s1, c1 := SingleStrawSolid()
	xf1 := model3d.JoinedTransform{
		model3d.Rotation(model3d.Y(1), 0.5),
		&model3d.Translate{Offset: model3d.XZ(-1.1, 0.1)},
	}
	s2 := model3d.TransformSolid(xf1, s1)
	c2 := c1.Transform(xf1)
	xf2 := model3d.JoinedTransform{
		model3d.Rotation(model3d.XY(-0.2, 1).Normalize(), -0.2),
		&model3d.Translate{Offset: model3d.XZ(0.3, 0.0)},
	}
	s3 := model3d.TransformSolid(xf2, s1)
	c3 := c1.Transform(xf2)
	return model3d.JoinedSolid{s2, s3}, toolbox3d.JoinedSolidCoordColorFunc(nil, s2, c2, s3, c3)
}

func SingleStrawSolid() (model3d.Solid, toolbox3d.CoordColorFunc) {
	outerSolid := &model3d.Cylinder{
		P1:     model3d.Z(CupHeight * 0.9),
		P2:     model3d.Z(CupHeight + StrawHeight),
		Radius: StrawRadius,
	}
	solid := &model3d.SubtractedSolid{
		Positive: outerSolid,
		Negative: &model3d.Sphere{
			Center: outerSolid.P2,
			Radius: StrawRadius - StrawInset,
		},
	}
	colorFn := func(c model3d.Coord3D) render3d.Color {
		if c.XY().Norm() <= StrawRadius-StrawInset {
			return render3d.NewColor(0.8)
		}
		theta := math.Atan2(c.X, c.Y)
		twist := c.Z * StrawTwistRate
		theta += twist
		for theta < 0 {
			theta += math.Pi * 2
		}
		for theta > math.Pi*2 {
			theta -= math.Pi * 2
		}
		if theta < math.Pi/2 || (theta > math.Pi && theta < math.Pi+math.Pi/2) {
			return render3d.NewColorRGB(1.0, 0.0, 0.0)
		} else {
			return render3d.NewColor(0.8)
		}
	}
	return solid, colorFn
}
