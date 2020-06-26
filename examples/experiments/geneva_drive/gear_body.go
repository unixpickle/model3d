package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

// DrivenBody creates a 3D solid implementing the drive
// gear.
func DriveBody(s *Spec, profile model2d.Solid) model3d.Solid {
	stack := model3d.StackSolids(
		&model3d.Cylinder{
			P2: model3d.Z(s.BottomThickness),
			// Jut out the bottom of the drive a bit so that the
			// driven gear rests on it.
			Radius: s.DriveRadius() + s.PinRadius*2,
		},
		model3d.ProfileSolid(profile, 0, s.Thickness),
	)
	return &model3d.SubtractedSolid{
		Positive: stack,
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: -1e-5},
			P2:         model3d.Coord3D{Z: s.BottomThickness + s.Thickness + 1e-5},
			Radius:     s.ScrewRadius + s.ScrewSlack,
			GrooveSize: s.ScrewGroove,
		},
	}
}

// DrivenBody creates a 3D solid implementing the driven
// gear.
func DrivenBody(s *Spec, profile model2d.Solid) model3d.Solid {
	stack := model3d.StackSolids(
		&model3d.Cylinder{
			P2:     model3d.Z(s.BottomThickness),
			Radius: s.DrivenSupportRadius,
		},
		model3d.ProfileSolid(profile, 0, s.Thickness),
	)
	return &model3d.SubtractedSolid{
		Positive: stack,
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: -1e-5},
			P2:         model3d.Coord3D{Z: s.BottomThickness + s.Thickness + 1e-5},
			Radius:     s.ScrewRadius + s.ScrewSlack,
			GrooveSize: s.ScrewGroove,
		},
	}
}
