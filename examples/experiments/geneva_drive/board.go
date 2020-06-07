package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

// BoardSolid creates the board to attach the gears to.
func BoardSolid(spec *Spec) model3d.Solid {
	return &model3d.SubtractedSolid{
		Positive: &model3d.RectSolid{
			MinVal: model3d.Coord3D{
				Y: -math.Max(spec.DriveRadius(), spec.DrivenRadius),
			},
			MaxVal: model3d.Coord3D{
				X: spec.CenterDistance + spec.DriveRadius() + spec.DrivenRadius,
				Y: math.Max(spec.DriveRadius(), spec.DrivenRadius),
				Z: spec.BoardThickness,
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Coord3D{X: spec.DriveRadius(), Z: -1e-5},
				P2:     model3d.Coord3D{X: spec.DriveRadius(), Z: spec.BoardThickness + 1e-5},
				Radius: spec.ScrewRadius + spec.ScrewSlack,
			},
			&model3d.Cylinder{
				P1: model3d.Coord3D{X: spec.DriveRadius() + spec.CenterDistance, Z: -1e-5},
				P2: model3d.Coord3D{X: spec.DriveRadius() + spec.CenterDistance,
					Z: spec.BoardThickness + 1e-5},
				Radius: spec.ScrewRadius + spec.ScrewSlack,
			},
		},
	}
}

// BoardScrewSolid creates the screw to use to attach the
// two gears to the board.
func BoardScrewSolid(spec *Spec) model3d.Solid {
	return model3d.StackSolids(
		&model3d.Cylinder{
			P2:     model3d.Z(spec.ScrewCapHeight),
			Radius: spec.ScrewCapRadius,
		},
		&model3d.Cylinder{
			P2:     model3d.Coord3D{Z: spec.BoardThickness + spec.ScrewSlack},
			Radius: spec.ScrewRadius,
		},
		&toolbox3d.ScrewSolid{
			P2:         model3d.Coord3D{Z: spec.BottomThickness + spec.Thickness},
			Radius:     spec.ScrewRadius,
			GrooveSize: spec.ScrewGroove,
		},
	)
}
