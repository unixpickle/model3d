package main

import (
	"io/ioutil"
	"math"

	"github.com/unixpickle/model3d"
)

func main() {
	screw := model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{},
			P2:     model3d.Coord3D{Z: 0.2},
			Radius: 0.2,
		},
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: 0.2},
			P2:     model3d.Coord3D{Z: 1.0},
			Radius: 0.09,
		},
		&ScrewSolid{
			Center:    model3d.Coord3D{Z: 0.2},
			Height:    0.8,
			Thickness: 0.05,
			Spacing:   0.12,
			Radius:    0.09,
		},
	}
	mesh := model3d.SolidToMesh(screw, 0.015, 2, 0.8, 5)
	ioutil.WriteFile("screw.stl", mesh.EncodeSTL(), 0755)

	hole := model3d.JoinedSolid{
		&model3d.SubtractedSolid{
			Positive: &model3d.CylinderSolid{
				P1:     model3d.Coord3D{},
				P2:     model3d.Coord3D{Z: 1.0},
				Radius: 0.4,
			},
			Negative: &model3d.CylinderSolid{
				P1:     model3d.Coord3D{},
				P2:     model3d.Coord3D{Z: 1.0},
				Radius: 0.15,
			},
		},
		&ScrewSolid{
			Inwards:   true,
			Center:    model3d.Coord3D{},
			Height:    1.0,
			Thickness: 0.05,
			Spacing:   0.12,
			Radius:    0.11,
		},
	}
	mesh = model3d.SolidToMesh(hole, 0.02, 2, 0.8, 5)
	ioutil.WriteFile("hole.stl", mesh.EncodeSTL(), 0755)
}

type ScrewSolid struct {
	Inwards bool

	// Center of the bottom face.
	Center model3d.Coord3D

	Height    float64
	Spacing   float64
	Thickness float64
	Radius    float64
}

func (s *ScrewSolid) Min() model3d.Coord3D {
	return s.Center.Sub(model3d.Coord3D{X: s.Radius + s.Thickness, Y: s.Radius + s.Thickness})
}

func (s *ScrewSolid) Max() model3d.Coord3D {
	return s.Center.Add(model3d.Coord3D{
		X: s.Radius + s.Thickness,
		Y: s.Radius + s.Thickness,
		Z: s.Height,
	})
}

func (s *ScrewSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(s.Min()) != s.Min() || c.Max(s.Max()) != s.Max() {
		return false
	}
	offset := c.Sub(s.Center)

	maxDistance := (model3d.Coord2D{X: offset.X, Y: offset.Y}).Norm() - s.Radius
	if !s.Inwards {
		maxDistance = s.Thickness - maxDistance
	}
	if maxDistance < 0 || maxDistance > s.Thickness {
		return false
	}

	angle := math.Atan2(offset.Y, offset.X)
	thetaToZ := s.Spacing / (math.Pi * 2)
	for z := (-math.Pi*2 + angle) * thetaToZ; z <= s.Height+s.Spacing; z += s.Spacing {
		if math.Abs(z-offset.Z) <= maxDistance {
			return true
		}
	}

	return false
}
