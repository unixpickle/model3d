package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateCradle(a *Args) model3d.Solid {
	solid := model3d.JoinedSolid{
		// Base.
		&model3d.Rect{
			MinVal: model3d.Coord3D{X: -CradleThickness, Y: -(CradleThickness + CradleBaseSide),
				Z: -CradleBottomThickness},
			MaxVal: model3d.Coord3D{X: a.PhoneWidth + CradleThickness,
				Y: a.PhoneDepth + CradleThickness + CradleBaseSide},
		},
		VerticalHolder{Args: a},
		&model3d.SubtractedSolid{
			Positive: &model3d.Rect{
				MinVal: model3d.Coord3D{X: -CradleThickness, Y: -CradleThickness,
					Z: -CradleThickness},
				MaxVal: model3d.Coord3D{X: a.PhoneWidth + CradleThickness,
					Y: CradleThickness + a.PhoneDepth, Z: CradleSideHeight},
			},
			Negative: &model3d.Rect{
				MinVal: model3d.Coord3D{X: 0, Y: 0},
				MaxVal: model3d.Coord3D{X: a.PhoneWidth + CradleThickness + 1e-5, Y: a.PhoneDepth,
					Z: CradleSideHeight + 1e-5},
			},
		},
	}
	return &model3d.SubtractedSolid{
		Positive: solid,
		Negative: &toolbox3d.ScrewSolid{
			P1: model3d.Coord3D{X: a.PhoneWidth / 2, Y: a.PhoneDepth / 2,
				Z: -(CradleBottomThickness + 1e-5)},
			P2:         model3d.Coord3D{X: a.PhoneWidth / 2, Y: a.PhoneDepth / 2, Z: 1e-5},
			Radius:     ScrewRadius + ScrewSlack,
			GrooveSize: ScrewGroove,
		},
	}
}

type VerticalHolder struct {
	*Args
}

func (v VerticalHolder) Min() model3d.Coord3D {
	return model3d.Coord3D{X: v.PhoneWidth/2 - VerticalHolderWidth/2, Y: -CradleThickness * 2}
}

func (v VerticalHolder) Max() model3d.Coord3D {
	return model3d.Coord3D{
		X: v.PhoneWidth/2 + VerticalHolderWidth/2,
		Y: v.PhoneDepth + CradleThickness*2,
		Z: v.PhoneHeight + CradleHeightSlack + v.PhoneDepth/2 + CradleThickness,
	}
}

func (v VerticalHolder) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(v, c) {
		return false
	}
	archZ := v.PhoneHeight + CradleHeightSlack
	if c.Z < archZ {
		return c.Y < 0 || c.Y > v.PhoneDepth
	}
	inset := c.Z - archZ
	return c.Y < inset || c.Y > v.PhoneDepth-inset
}
