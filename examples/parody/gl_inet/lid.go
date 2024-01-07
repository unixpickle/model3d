package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	LidThickness = 0.08
	LidSlack     = 0.02
	InnerInset   = 0.35
	BodyHoleMinZ = 0.2
)

func LidAndCutout() (model3d.Solid, model3d.Solid) {
	body := InetBody(false)
	lid := &model3d.SubtractedSolid{
		Positive: body,
		Negative: model3d.TranslateSolid(body, model3d.Z(-LidThickness)),
	}
	baseRect := model2d.NewRect(
		model2d.Ones(-BodySideLength/2+BodyCornerRadius+InnerInset),
		model2d.Ones(BodySideLength/2-BodyCornerRadius-InnerInset),
	)
	bodyHole := model3d.ProfileSolid(
		model2d.NewColliderSolidInset(baseRect, -BodyCornerRadius),
		BodyHoleMinZ,
		100.0,
	)
	holes := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.XYZ(-0.25, 0.9, 0.8),
			P2:     model3d.XYZ(-0.25, 0.9, 1.0),
			Radius: 0.02,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(0.0, 0.9, 0.8),
			P2:     model3d.XYZ(0.0, 0.9, 1.0),
			Radius: 0.02,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(0.25, 0.9, 0.8),
			P2:     model3d.XYZ(0.25, 0.9, 1.0),
			Radius: 0.02,
		},
	}
	lidInnerEdge := &model3d.SubtractedSolid{
		Positive: model3d.ProfileSolid(
			model2d.NewColliderSolidInset(baseRect, -BodyCornerRadius+LidSlack),
			0.7,
			0.89,
		),
		Negative: model3d.ProfileSolid(
			model2d.NewColliderSolidInset(baseRect, -BodyCornerRadius+LidSlack+LidThickness),
			0.7,
			0.89,
		),
	}

	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{lid, lidInnerEdge},
		Negative: holes,
	}, model3d.JoinedSolid{bodyHole, lid}
}
