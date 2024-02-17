package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func Base() (model3d.Solid, toolbox3d.CoordColorFunc) {
	return &model3d.Cylinder{
		P1:     model3d.Z(-0.5),
		P2:     model3d.Z(-0.2),
		Radius: 1.5,
	}, toolbox3d.ConstantCoordColorFunc(BodyColor)
}
