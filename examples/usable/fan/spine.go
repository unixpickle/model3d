package main

import (
	"github.com/unixpickle/model3d"
)

func SpineMesh() *model3d.Mesh {
	solid := SpineSolid()
	mesh := model3d.SolidToMesh(solid, 0.015, 0, -1, 5)
	return mesh
}

func SpineSolid() model3d.Solid {
	center1 := model3d.Coord3D{X: SpineWidth / 2, Y: SpineWidth / 2}
	center2 := center1.Add(model3d.Coord3D{Y: GearDistance})
	thickVec := model3d.Coord3D{Z: SpineThickness + SpineWasherSize}
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.RectSolid{
				MaxVal: model3d.Coord3D{X: SpineWidth, Y: SpineLength, Z: SpineThickness},
			},
			&model3d.CylinderSolid{
				P1:     center1,
				P2:     center1.Add(thickVec),
				Radius: SpineWasherRadius,
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.CylinderSolid{
				P1:     center1,
				P2:     center1.Add(thickVec),
				Radius: HoleRadius,
			},
			&model3d.CylinderSolid{
				P1:     center2,
				P2:     center2.Add(thickVec),
				Radius: HoleRadius,
			},
		},
	}
}
