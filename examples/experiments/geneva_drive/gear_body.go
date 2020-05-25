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
			P2: model3d.Coord3D{Z: s.BottomThickness},
			// Jut out the bottom of the drive a bit so that the
			// driven gear rests on it.
			Radius: s.DriveRadius() + s.PinRadius*2,
		},
		&profileSolid{
			Solid:     profile,
			Thickness: s.Thickness,
		},
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
			P2:     model3d.Coord3D{Z: s.BottomThickness},
			Radius: s.DrivenSupportRadius,
		},
		&profileSolid{
			Solid:     profile,
			Thickness: s.Thickness,
		},
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

type profileSolid struct {
	Solid     model2d.Solid
	Thickness float64
}

func (p *profileSolid) Min() model3d.Coord3D {
	m2 := p.Solid.Min()
	return model3d.Coord3D{X: m2.X, Y: m2.Y, Z: 0}
}

func (p *profileSolid) Max() model3d.Coord3D {
	m2 := p.Solid.Max()
	return model3d.Coord3D{X: m2.X, Y: m2.Y, Z: p.Thickness}
}

func (p *profileSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(p, c) && p.Solid.Contains(c.Coord2D())
}
