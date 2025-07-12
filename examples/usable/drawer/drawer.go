package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateDrawer() model3d.Solid {
	min := model3d.Coord3D{
		X: DrawerSlack,
		Y: 0,
		Z: FrameThickness + DrawerSlack,
	}
	max := model3d.Coord3D{
		X: DrawerWidth - DrawerSlack,
		Y: DrawerDepth - DrawerSlack,
		Z: FrameThickness + DrawerHeight - DrawerSlack,
	}

	result := model3d.JoinedSolid{
		// Bottom face.
		&model3d.Rect{
			MinVal: min,
			MaxVal: model3d.XYZ(max.X, max.Y, min.Z+DrawerBottom),
		},
	}

	// Side faces.
	for _, x := range []float64{min.X, max.X - DrawerThickness} {
		result = append(result, &model3d.Rect{
			MinVal: model3d.XYZ(x, min.Y, min.Z),
			MaxVal: model3d.XYZ(x+DrawerThickness, max.Y, max.Z),
		})
	}

	// Front/back faces.
	for _, y := range []float64{min.Y, max.Y - DrawerThickness} {
		result = append(result, &model3d.Rect{
			MinVal: model3d.XYZ(min.X, y, min.Z),
			MaxVal: model3d.XYZ(max.X, y+DrawerThickness, max.Z),
		})
	}

	mid := min.Mid(max)

	return model3d.Subtract(
		result,
		model3d.JoinedSolid{
			&RidgeSolid{X1: min.X, X2: min.X + RidgeDepth, Z: mid.Z},
			&RidgeSolid{X1: max.X, X2: max.X - RidgeDepth, Z: mid.Z},
			toolbox3d.Teardrop3D(
				model3d.XYZ(mid.X, min.Y-1e-5, mid.Z),
				model3d.XYZ(mid.X, min.Y+DrawerThickness+1e-5, mid.Z),
				DrawerHoleRadius,
			),
		},
	)
}
