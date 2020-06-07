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
		FlagFabric{},
	}
}

type FlagFabric struct{}

func (f FlagFabric) Min() model3d.Coord3D {
	return model3d.Coord3D{X: 0, Y: -FlagThickness/2 - FlagRippleDepth*2,
		Z: FlagHeight - FlagWidth}
}

func (f FlagFabric) Max() model3d.Coord3D {
	return model3d.Coord3D{X: FlagWidth, Y: FlagThickness/2 + FlagRippleDepth*2,
		Z: FlagHeight}
}

func (f FlagFabric) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(f, c) {
		return false
	}

	if math.Abs(c.Y-FlagRippleDepth*math.Sin(c.X*FlagRippleRate)) > FlagThickness {
		return false
	}

	return c.X < c.Z-(FlagHeight-FlagWidth)
}
