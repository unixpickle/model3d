package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	FlagHeight     = 5.0
	FlagWidth      = 1.5
	FlagThickness  = 0.1
	FlagPoleRadius = 0.1

	FlagRippleRate  = math.Pi * 2
	FlagRippleDepth = 0.1
)

func GenerateFlag() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Cylinder{
			P2:     model3d.Z(FlagHeight),
			Radius: FlagPoleRadius,
		},
		FlagFabric(),
	}
}

func FlagFabric() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(0, -FlagThickness/2-FlagRippleDepth*2, FlagHeight-FlagWidth),
		model3d.XYZ(FlagWidth, FlagThickness/2+FlagRippleDepth*2, FlagHeight),
		func(c model3d.Coord3D) bool {
			if math.Abs(c.Y-FlagRippleDepth*math.Sin(c.X*FlagRippleRate)) > FlagThickness {
				return false
			}
			return c.X < c.Z-(FlagHeight-FlagWidth)
		},
	)
}
