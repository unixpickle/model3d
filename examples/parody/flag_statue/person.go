package main

import (
	"math"

	"github.com/unixpickle/model3d"
)

const (
	PersonRadius = 1.0
	PersonHeight = 3.0

	PersonTorsoHeight    = 0.9
	PersonTorsoWidth     = 0.5
	PersonTorsoThickness = 0.3

	PersonLegHeight   = 1.1
	PersonLegSpace    = 0.1
	PersonWaistInset  = 0.05
	PersonWaistHeight = 0.1

	PersonArmHeight   = 0.8
	PersonArmRadius   = 0.1
	PersonArmTipSpace = 0.1
	PersonShoulder    = 0.1

	PersonHeadRadius = 0.2
	PersonHeadOffset = PersonHeadRadius / 2
)

func GeneratePeople() model3d.Solid {
	return model3d.JoinedSolid{
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: 0.8, Z: 0.8},
			model3d.Coord3D{X: -0.2, Z: 1},
		)),
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: 2, Z: 0.6},
			model3d.Coord3D{X: -0.38, Z: 1},
		)),
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: -1, Z: 0.8},
			model3d.Coord3D{X: 0.33, Z: 1},
		)),
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: -2, Z: 0.6},
			model3d.Coord3D{X: 0.4, Z: 1},
		)),
	}
}

type PersonSolid struct {
	P1 model3d.Coord3D
	P2 model3d.Coord3D
}

func NewPersonSolid(footCenter, axis model3d.Coord3D) *PersonSolid {
	return &PersonSolid{
		P1: footCenter,
		P2: footCenter.Add(axis.Normalize().Scale(PersonHeight)),
	}
}

func (p *PersonSolid) Min() model3d.Coord3D {
	return p.boundingCylinder().Min()
}

func (p *PersonSolid) Max() model3d.Coord3D {
	return p.boundingCylinder().Max()
}

func (p *PersonSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}

	// Change of basis.
	zAxis := p.P2.Sub(p.P1).Normalize()
	xAxis := model3d.Coord3D{Y: 1}
	yAxis := xAxis.Cross(zAxis)
	c = c.Sub(p.P1)
	c = model3d.Coord3D{X: c.Dot(xAxis), Y: c.Dot(yAxis), Z: c.Dot(zAxis)}
	c2 := c.Coord2D()

	if c.Z < 0 {
		return false
	}
	if c.Z < PersonLegHeight+PersonTorsoHeight {
		var inset float64
		if c.Z < PersonLegHeight {
			inset = math.Max(0, math.Min(PersonWaistHeight, PersonLegHeight-c.Z)) *
				PersonWaistInset / PersonWaistHeight
		}
		cylinderParams := model3d.Coord2D{
			X: 1 / (PersonTorsoWidth - inset),
			Y: 1 / (PersonTorsoThickness - inset),
		}
		if c2.Mul(cylinderParams).Norm() >= 1 {
			return false
		}
		if c.Z < PersonLegHeight {
			gap := math.Min(PersonLegSpace, PersonLegHeight-c.Z)
			return math.Abs(c.X) >= gap
		}
		return true
	}

	topSolid := model3d.JoinedSolid{
		&model3d.SphereSolid{
			Center: model3d.Coord3D{
				Z: PersonLegHeight + PersonTorsoHeight + PersonHeadOffset,
			},
			Radius: PersonHeadRadius,
		},
		&model3d.CylinderSolid{
			P1: model3d.Coord3D{
				X: -PersonTorsoWidth + PersonShoulder,
				Z: PersonLegHeight + PersonTorsoHeight - PersonShoulder,
			},
			P2: model3d.Coord3D{
				X: -PersonArmTipSpace,
				Z: PersonLegHeight + PersonTorsoHeight + PersonArmHeight,
			},
			Radius: PersonArmRadius,
		},
		&model3d.CylinderSolid{
			P1: model3d.Coord3D{
				X: PersonTorsoWidth - PersonShoulder,
				Z: PersonLegHeight + PersonTorsoHeight - PersonShoulder,
			},
			P2: model3d.Coord3D{
				X: PersonArmTipSpace,
				Z: PersonLegHeight + PersonTorsoHeight + PersonArmHeight,
			},
			Radius: PersonArmRadius,
		},
	}

	return topSolid.Contains(c)
}

func (p *PersonSolid) boundingCylinder() *model3d.CylinderSolid {
	return &model3d.CylinderSolid{
		P1:     p.P1,
		P2:     p.P2,
		Radius: PersonRadius,
	}
}
