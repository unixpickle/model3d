package main

import (
	"github.com/unixpickle/model3d"
)

const (
	PersonRadius = 1.0
	PersonHeight = 3.0

	PersonTorsoHeight    = 1.0
	PersonTorsoWidth     = 0.5
	PersonTorsoThickness = 0.3

	PersonLegHeight = 1.0
	PersonLegRadius = 0.18
	PersonLegSpace  = 0.23

	PersonArmHeight   = 0.8
	PersonArmRadius   = 0.1
	PersonArmTipSpace = 0.1
	PersonShoulder    = 0.1

	PersonHeadRadius = 0.2
)

func GeneratePeople() model3d.Solid {
	return model3d.JoinedSolid{
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: 0.8, Z: 0.8},
			model3d.Coord3D{X: -0.2, Z: 1},
		)),
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: 2, Z: 0.5},
			model3d.Coord3D{X: -0.38, Z: 1},
		)),
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: -1, Z: 0.8},
			model3d.Coord3D{X: 0.33, Z: 1},
		)),
		model3d.CacheSolidBounds(NewPersonSolid(
			model3d.Coord3D{X: -2, Z: 0.5},
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
	if !model3d.InSolidBounds(p, c) {
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
	} else if c.Z < PersonLegHeight {
		return c2.Sub(model3d.Coord2D{X: -PersonLegSpace}).Norm() < PersonLegRadius ||
			c2.Sub(model3d.Coord2D{X: PersonLegSpace}).Norm() < PersonLegRadius
	} else if c.Z < PersonLegHeight+PersonTorsoHeight {
		return c2.Mul(model3d.Coord2D{X: 1 / PersonTorsoWidth, Y: 1 / PersonTorsoThickness}).Norm() < 1
	}

	topSolid := model3d.JoinedSolid{
		&model3d.SphereSolid{
			Center: model3d.Coord3D{
				Z: PersonLegHeight + PersonTorsoHeight + PersonHeadRadius/3,
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
