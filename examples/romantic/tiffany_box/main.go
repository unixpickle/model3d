package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	Thickness = 0.1
	Slack     = 0.04

	RibbonThickness = 0.05
	RibbonWidth     = 0.2
)

func main() {
	bottomSides := &model3d.SubtractedSolid{
		Positive: model3d.NewRect(
			model3d.Ones(-1.0),
			model3d.XYZ(1.0, 1.0, 1.0-Thickness),
		),
		Negative: model3d.NewRect(
			model3d.Ones(-1.0+Thickness),
			model3d.XYZ(1.0-Thickness, 1.0-Thickness, 1.0-Thickness+1e-5),
		),
	}
	bottomRibbon := RotateFourTimes(
		model3d.JoinedSolid{
			// On the bottom of the box.
			model3d.NewRect(
				model3d.XYZ(-(1.0+RibbonThickness), -RibbonWidth/2, -(1.0+RibbonThickness)),
				model3d.XYZ(0, RibbonWidth/2, -1.0),
			),
			// Going up the side of the box.
			model3d.NewRect(
				model3d.XYZ(-(1.0+RibbonThickness), -RibbonWidth/2, -(1.0+RibbonThickness)),
				model3d.XYZ(-1.0, RibbonWidth/2, 0.5),
			),
			// Going across below the lip of the lid.
			model3d.NewRect(
				model3d.XYZ(-(1.0+Thickness), -RibbonWidth/2, 0.5-RibbonThickness),
				model3d.XYZ(-1.0, RibbonWidth/2, 0.5),
			),
		},
	)
	bottom := model3d.JoinedSolid{bottomSides, bottomRibbon}

	topSides := &model3d.SubtractedSolid{
		Positive: model3d.NewRect(
			model3d.XYZ(-(1.0+Thickness+Slack), -(1.0+Thickness+Slack), 0.5),
			model3d.XYZ(1.0+Thickness+Slack, 1.0+Thickness+Slack, 1.0),
		),
		Negative: model3d.NewRect(
			model3d.XYZ(-(1.0+Slack), -(1.0+Slack), 0.5-1e-5),
			model3d.XYZ(1.0+Slack, 1.0+Slack, 1.0-Thickness),
		),
	}
	topRibbon := RotateFourTimes(model3d.JoinedSolid{
		// On side of lid
		model3d.NewRect(
			model3d.XYZ(
				-(1.0+Thickness+Slack+RibbonThickness),
				-RibbonWidth/2,
				0.5-RibbonThickness,
			),
			model3d.XYZ(-(1.0+Thickness+Slack), RibbonWidth/2, 1.0+RibbonThickness),
		),
		// On top of lid
		model3d.NewRect(
			model3d.XYZ(-(1.0+Thickness+Slack), -RibbonWidth/2, 1.0),
			model3d.XYZ(0, RibbonWidth/2, 1.0+RibbonThickness),
		),
	})
	top := model3d.JoinedSolid{topSides, topRibbon}

	bottomMesh := model3d.MarchingCubesSearch(bottom, 0.01, 8)
	bottomMesh.SaveGroupedSTL("bottom.stl")

	topMesh := model3d.MarchingCubesSearch(top, 0.01, 8)
	topMesh.SaveGroupedSTL("top.stl")

	bottomMesh.AddMesh(topMesh)
	bottomMesh.SaveGroupedSTL("both.stl")
}

func RotateFourTimes(s model3d.Solid) model3d.Solid {
	res := model3d.JoinedSolid{}
	for i := 0; i < 4; i++ {
		angle := math.Pi / 2.0 * float64(i)
		res = append(res, model3d.RotateSolid(s, model3d.Z(1), angle))
	}
	return res
}
