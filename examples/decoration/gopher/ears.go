package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func Ears() (model3d.Solid, toolbox3d.CoordColorFunc) {
	const tiltFactor = 0.99
	earSolid := &model3d.Cylinder{
		P1:     model3d.XYZ(-0.8, 0.1, 2.0),
		P2:     model3d.XYZ(-0.8*tiltFactor, 0.2, 2.0*tiltFactor),
		Radius: 0.2,
	}
	insideSolid := &model3d.Cylinder{
		P1:     model3d.XYZ(-0.8, 0.1, 2.0),
		P2:     model3d.XYZ(-0.8*tiltFactor, 0.2, 2.0*tiltFactor),
		Radius: 0.1,
	}
	insideSolid.P1.Y -= 0.1
	insideSolid.P2.Y += 0.1

	otherEar := *earSolid
	otherEar.P1.X *= -1
	otherEar.P2.X *= -1
	bothEars := model3d.JoinedSolid{
		model3d.NewColliderSolidInset(earSolid, -0.03),
		model3d.NewColliderSolidInset(&otherEar, -0.03),
	}

	otherInside := *insideSolid
	otherInside.P1.X *= -1
	otherInside.P2.X *= -1
	bothInside := model3d.JoinedSolid{
		insideSolid, &otherInside,
	}

	return bothEars, func(c model3d.Coord3D) render3d.Color {
		if bothInside.Contains(c) {
			return render3d.NewColor(0)
		} else {
			return BodyColor
		}
	}
}
