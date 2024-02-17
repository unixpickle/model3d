package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func Teeth() (model3d.Solid, toolbox3d.CoordColorFunc) {
	tooth := &model3d.Capsule{
		P1:     model3d.YZ(0.69, 1.15),
		P2:     model3d.YZ(0.69, 0.98),
		Radius: 0.055,
	}
	teeth := model3d.JoinedSolid{
		model3d.TranslateSolid(tooth, model3d.X(-0.055)),
		model3d.TranslateSolid(tooth, model3d.X(0.055)),
	}
	return teeth, toolbox3d.ConstantCoordColorFunc(render3d.NewColor(1.0))
}
