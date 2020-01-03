package toolbox3d

import (
	"math"

	"github.com/unixpickle/model3d"
)

// A ScrewSolid is a model3d.Solid implementation of
// screws. It can also be used for screw holes, by
// combining it with model3d.SubtractedSolid.
//
// Screws are similar to cylinders, so many of the fields
// are analogous to model3d.CylinderSolid.
type ScrewSolid struct {
	// P1 is the center of the start of the screw.
	P1 model3d.Coord3D

	// P2 is the center of the end of the screw.
	P2 model3d.Coord3D

	// Radius is the maximum radius of the screw,
	// including grooves.
	Radius float64

	// GrooveSize is the size of the grooves.
	// This may not exceed Radius.
	GrooveSize float64

	// Pointed can be set to true to indicate that the tip
	// at the P2 end should be cut off at a 45 degree
	// angle (in the shape of a cone).
	// Can be used for internal screw holes to avoid
	// support.
	Pointed bool
}

func (s *ScrewSolid) Min() model3d.Coord3D {
	return s.boundingCylinder().Min()
}

func (s *ScrewSolid) Max() model3d.Coord3D {
	return s.boundingCylinder().Max()
}

func (s *ScrewSolid) Contains(c model3d.Coord3D) bool {
	diff := s.P2.Sub(s.P1)
	height := diff.Norm()
	axis := diff.Normalize()
	b1, b2 := axis.OrthoBasis()

	// Make sure basis obeys right-hand rule.
	if b1.Cross(b2).Dot(axis) < 0 {
		b2, b1 = b1, b2
	}

	offset := c.Sub(s.P1)
	offset = model3d.Coord3D{
		X: offset.Dot(b1),
		Y: offset.Dot(b2),
		Z: offset.Dot(axis),
	}

	if s.Pointed {
		constrainedRadius := (height - offset.Z)
		if offset.Coord2D().Norm() > constrainedRadius {
			return false
		}
	}

	if offset.Z < 0 || offset.Z > height {
		return false
	}

	maxDistance := s.Radius - offset.Coord2D().Norm()
	if maxDistance < 0 {
		return false
	} else if maxDistance > s.GrooveSize {
		return true
	}

	zOffset := math.Atan2(offset.Y, offset.X) * s.GrooveSize / math.Pi
	offZ := offset.Z - zOffset
	roundedZ := math.Round(offZ/(s.GrooveSize*2)) * s.GrooveSize * 2
	if math.Abs(roundedZ-offZ) <= maxDistance {
		return true
	}

	return false
}

func (s *ScrewSolid) boundingCylinder() *model3d.CylinderSolid {
	return &model3d.CylinderSolid{
		P1:     s.P1,
		P2:     s.P2,
		Radius: s.Radius,
	}
}
