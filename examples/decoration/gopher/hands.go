package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var HandColor = NoseColor

func Hands() (model3d.Solid, toolbox3d.CoordColorFunc) {
	c1 := &model3d.Capsule{
		P1:     model3d.XYZ(0, 0.1, 0.8),
		P2:     model3d.XYZ(0.1, 0.12, 0.76),
		Radius: 0.03,
	}
	c2 := &model3d.Capsule{
		P1:     model3d.XYZ(0, 0.1, 0.7),
		P2:     model3d.XYZ(0.1, 0.12, 0.66),
		Radius: 0.03,
	}
	singleSolid := model3d.MetaballSolid(nil, 0.07, c1, c2)
	solid1 := model3d.TranslateSolid(singleSolid, model3d.XZ(0.95, 0.05))
	solid2 := model3d.VecScaleSolid(solid1, model3d.XYZ(-1, 1, 1))
	solid := model3d.JoinedSolid{solid1, solid2}
	return solid, toolbox3d.ConstantCoordColorFunc(HandColor)
}
