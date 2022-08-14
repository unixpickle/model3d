package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	AxleRadius         = 0.15
	BaseThickness      = 0.1
	WheelThickness     = 0.15
	WheelRadius        = 0.35
	WheelSlip          = 0.05
	ArchPanelWidth     = 1.0
	ArchPanelHeight    = 3.0
	ArchPanelThickness = 0.1
)

func main() {
	model := model3d.JoinedSolid{BasePlatform(), ArchSides()}
	mesh := model3d.MarchingCubesSearch(model, 0.02, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func BasePlatform() model3d.Solid {
	base := model3d.NewRect(
		model3d.XYZ(-3.5, -1.5, -BaseThickness),
		model3d.XYZ(3.5, 1.5, 1e-5),
	)
	var wheels model3d.JoinedSolid
	firstX := 3.5 - WheelRadius*2
	for _, x := range []float64{-firstX, -firstX + WheelRadius*2.5, firstX - WheelRadius*2.5, firstX} {
		wheelZ := -(BaseThickness + WheelRadius - WheelSlip)
		for _, y := range []float64{-1.3, 1.3} {
			wheel := &model3d.Cylinder{
				P1:     model3d.XYZ(x, y-WheelThickness/2, wheelZ),
				P2:     model3d.XYZ(x, y+WheelThickness/2, wheelZ),
				Radius: WheelRadius,
			}
			wheels = append(wheels, wheel)
		}
		axle := model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.XYZ(x, -1.3, wheelZ),
				P2:     model3d.XYZ(x, 1.3, wheelZ),
				Radius: AxleRadius,
			},
			model3d.NewRect(
				model3d.XYZ(x-AxleRadius, -1.3, wheelZ),
				model3d.XYZ(x+AxleRadius, 1.3, -BaseThickness+1e-5),
			),
		}
		wheels = append(wheels, axle...)
	}
	return append(wheels, base).Optimize()
}

func ArchSides() model3d.Solid {
	var result model3d.JoinedSolid
	for i := 0; i < 7; i++ {
		x := -3.5 + ArchPanelWidth/2 + float64(i)*ArchPanelWidth
		fullHeight := i/3%2 == 0
		for _, y := range []float64{-1.5 + ArchPanelThickness/2, 1.5 - ArchPanelThickness/2} {
			result = append(result, ArchPanel(model3d.XYZ(x, y, ArchPanelHeight/2), fullHeight))
		}
	}

	var uncurvedPanel model3d.JoinedSolid
	for i := 0; i < 3; i++ {
		uncurvedPanel = append(uncurvedPanel, ArchPanel(
			model3d.XYZ(ArchPanelWidth*float64(i), 3.5-ArchPanelWidth, ArchPanelHeight/2),
			false,
		))
	}
	curveOffset := 0.2
	curvedPanel := model3d.CheckedFuncSolid(
		uncurvedPanel.Min(),
		uncurvedPanel.Max().Add(model3d.Y(curveOffset)),
		func(c model3d.Coord3D) bool {
			if c.X < ArchPanelWidth {
				c.Y -= curveOffset * c.X / ArchPanelWidth
			} else if c.X < ArchPanelWidth*2 {
				c.Y -= curveOffset
			} else {
				c.Y -= curveOffset * (ArchPanelWidth*3 - c.X) / ArchPanelWidth
			}
			return uncurvedPanel.Contains(c)
		},
	)

	sidePanel := model3d.RotateSolid(
		model3d.TranslateSolid(curvedPanel, model3d.X(-ArchPanelWidth*1.5)),
		model3d.Z(1),
		math.Pi/2,
	)
	otherSide := model3d.RotateSolid(sidePanel, model3d.Z(1), math.Pi)
	result = append(result, sidePanel, otherSide)
	return result
}

func ArchPanel(center model3d.Coord3D, fullHeight bool) model3d.Solid {
	outer := model3d.NewRect(
		model3d.XYZ(-ArchPanelWidth/2, -ArchPanelThickness/2, -ArchPanelHeight/2),
		model3d.XYZ(ArchPanelWidth/2, ArchPanelThickness/2, ArchPanelHeight/2),
	)
	inset := 0.1
	archRadius := ArchPanelWidth/2 - inset
	arch := &model3d.Cylinder{
		P1:     model3d.XYZ(0, -ArchPanelThickness-1e-5, ArchPanelHeight/2-(inset+archRadius)),
		P2:     model3d.XYZ(0, ArchPanelThickness+1e-5, ArchPanelHeight/2-(inset+archRadius)),
		Radius: archRadius,
	}
	archBottom := model3d.NewRect(
		model3d.XYZ(-archRadius, -ArchPanelThickness-1e-5, -ArchPanelHeight/2+inset),
		model3d.XYZ(-archRadius, ArchPanelThickness+1e-5, ArchPanelHeight/2-(inset+archRadius)+0.1),
	)
	if !fullHeight {
		archBottom.MinVal.Z = 0
	}
	untranslated := &model3d.SubtractedSolid{
		Positive: outer,
		Negative: model3d.JoinedSolid{arch, archBottom},
	}
	return model3d.TranslateSolid(untranslated, center)
}
