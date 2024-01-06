package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	BodyCornerRadius    = 0.15
	BodySideLength      = 2.2
	BodyRadius          = 5.0
	BodyPowerHoleRadius = 0.03
	BodyPowerHoleWidth  = 0.15
)

func InetBody() model3d.Solid {
	sphere := model3d.Sphere{
		Center: model3d.Z(-BodyRadius + 1.0),
		Radius: BodyRadius,
	}
	baseRect := model2d.NewRect(
		model2d.Ones(-BodySideLength/2+BodyCornerRadius),
		model2d.Ones(BodySideLength/2-BodyCornerRadius),
	)
	holes := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.XYZ(-0.25, 0.9, 0.8),
			P2:     model3d.XYZ(-0.25, 1.0, 1.0),
			Radius: 0.02,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(0.0, 0.9, 0.8),
			P2:     model3d.XYZ(0.0, 1.0, 1.0),
			Radius: 0.02,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(0.25, 0.9, 0.8),
			P2:     model3d.XYZ(0.25, 1.0, 1.0),
			Radius: 0.02,
		},
	}
	powerHole2d := model2d.NewColliderSolidInset(
		model2d.NewRect(
			model2d.XY(0.8-BodyPowerHoleRadius-BodyPowerHoleWidth, 0.15+BodyPowerHoleRadius),
			model2d.XY(0.8-BodyPowerHoleRadius, 0.15+BodyPowerHoleRadius+1e-5),
		),
		-BodyPowerHoleRadius,
	)
	powerHole := model3d.RotateSolid(
		model3d.ProfileSolid(powerHole2d, BodySideLength/2-0.2, BodySideLength/2+0.05),
		model3d.X(1),
		math.Pi/2,
	)
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-BodySideLength/2, -BodySideLength/2, 0),
		model3d.XYZ(BodySideLength/2, BodySideLength/2, 1.0),
		func(c model3d.Coord3D) bool {
			if !sphere.Contains(c) {
				return false
			}
			if holes.Contains(c) || powerHole.Contains(c) {
				return false
			}
			return baseRect.SDF(c.XY()) > -BodyCornerRadius
		},
	)
}
