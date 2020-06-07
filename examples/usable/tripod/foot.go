package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateFoot() model3d.Solid {
	return model3d.StackSolids(
		&model3d.Cylinder{
			P2:     model3d.Z(0.2),
			Radius: TripodFootRadius,
		},
		&toolbox3d.ScrewSolid{
			P2:         model3d.Z(1.0),
			Radius:     ScrewRadius,
			GrooveSize: ScrewGroove,
		},
	)
}
