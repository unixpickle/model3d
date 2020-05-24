package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
)

// DrivenProfile creates a 2D Solid implementing the
// driven disk in a Geneva drive.
func DrivenProfile(s *Spec) model2d.Solid {
	var negative model2d.JoinedSolid

	// Cut out the circles corresponding to the driver.
	for i := 0; i < 4; i++ {
		theta := math.Pi * 2 * float64(i) / 4
		center := model2d.Coord{X: math.Cos(theta), Y: math.Sin(theta)}
		center = center.Scale(s.CenterDistance - s.Slack)
		negative = append(negative, &model2d.Circle{
			Center: center,
			Radius: s.DriveRadius,
		})
	}

	// Cut out lines where the pin will enter the driven disk.
	for i := 0; i < 4; i++ {
		theta := math.Pi*2*float64(i)/4 + math.Pi/4
		direction := model2d.Coord{X: math.Cos(theta), Y: math.Sin(theta)}
		innerPoint := direction.Scale(s.CenterDistance - s.DriveRadius - s.Slack)
		outerPoint := direction.Scale(s.DrivenRadius)
		mesh := model2d.NewMeshSegments([]*model2d.Segment{{innerPoint, outerPoint}})
		collider := model2d.MeshToCollider(mesh)
		negative = append(negative, model2d.NewColliderSolidHollow(collider, s.PinRadius+s.Slack))
	}

	baseProfile := &model2d.SubtractedSolid{
		Positive: &model2d.Circle{Radius: s.DrivenRadius},
		Negative: negative,
	}

	// Cut off the sharp edges to prevent tiny features.
	return model2d.IntersectedSolid{
		baseProfile,
		&model2d.Circle{Radius: MaximumRadius(baseProfile) - s.Slack},
	}
}

// DriveProfile creates a 2D Solid implementing the drive
// disk in a Geneva drive.
//
// Requires the driven profile to cut it out properly.
func DriveProfile(s *Spec, driven model2d.Solid) model2d.Solid {
	return model2d.JoinedSolid{
		&model2d.Circle{
			Center: model2d.Coord{X: s.DriveRadius},
			Radius: s.PinRadius,
		},
		&model2d.SubtractedSolid{
			Positive: &model2d.Circle{
				Radius: s.DriveRadius,
			},
			Negative: &model2d.Circle{
				Center: model2d.Coord{X: s.CenterDistance},
				Radius: MaximumRadius(driven) + s.Slack*2,
			},
		},
	}
}

// MaximumRadius finds the smallest radius of a circle
// centered at the origin that doesn't touch s.
func MaximumRadius(s model2d.Solid) float64 {
	max := s.Max().Dist(s.Min())
	min := 0.0
	for i := 0; i < 32; i++ {
		r := (max + min) / 2
		if collidesRadius(s, r) {
			min = r
		} else {
			max = r
		}
	}
	return max
}

func collidesRadius(s model2d.Solid, radius float64) bool {
	for theta := 0.0; theta < math.Pi*2; theta += 0.01 {
		d := model2d.Coord{X: math.Cos(theta), Y: math.Sin(theta)}
		if s.Contains(d.Scale(radius)) {
			return true
		}
	}
	return false
}
