package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Dedendum        = 0.04
	Addendum        = 0.04
	PressureAngle   = 25 * math.Pi / 180
	Module          = 0.0495
	HoleSize        = 0.23
	HoleClearance   = 0.02
	PoleSize        = 0.2
	PoleBaseRadius  = 0.4
	ScrewRadius     = 0.15
	ScrewSlack      = 0.02
	ScrewGrooveSize = 0.05
)

func main() {
	CreateGear(20, "gear_20.stl", false)
	CreateGear(30, "gear_30.stl", false)
	CreateGear(40, "gear_40.stl", true)

	CreatePole()
	CreateHolder()
}

func CreateGear(teeth int, path string, invert bool) {
	var p1, p2 model3d.Coord3D
	p2.Z = 0.4
	theta := 20 * math.Pi / 180
	if invert {
		theta *= -1
	}
	solid := &model3d.SubtractedSolid{
		Positive: &toolbox3d.HelicalGear{
			P1: p1,
			P2: p2,
			Profile: toolbox3d.InvoluteGearProfileSizes(PressureAngle, Module, Dedendum, Addendum,
				teeth),
			Angle: theta,
		},
		Negative: &toolbox3d.ScrewSolid{
			// Fix rounding errors resulting in filled-in sides.
			P1: p1.Sub(model3d.Coord3D{Z: 0.001}),
			P2: p2.Add(model3d.Coord3D{Z: 0.001}),

			Radius:     ScrewRadius + ScrewSlack,
			GrooveSize: ScrewGrooveSize,
		},
	}
	mesh := model3d.SolidToMesh(solid, 0.004, 0, -1, 5)
	mesh.SaveGroupedSTL(path)
}

func CreatePole() {
	solid := model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{},
			P2:     model3d.Coord3D{Z: 0.2},
			Radius: PoleBaseRadius,
		},
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: 0.2},
			P2:     model3d.Coord3D{Z: 0.6},
			Radius: PoleSize,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: 0.6},
			P2:         model3d.Coord3D{Z: 1.0},
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooveSize,
		},
	}
	mesh := model3d.SolidToMesh(solid, 0.004, 0, -1, 5)
	mesh.SaveGroupedSTL("pole.stl")
}

func CreateHolder() {
	padding := Module * 10
	c1 := model3d.Coord3D{X: padding + Module*40/2, Y: padding + Module*40/2}
	c2 := model3d.Coord3D{X: padding + Module*40 + Module*30/2, Y: padding + Module*40/2}
	c3 := model3d.Coord3D{X: padding + Module*40/2, Y: padding + Module*40 + Module*20/2}
	c2.X += HoleClearance
	c3.Y += HoleClearance
	thickness := model3d.Coord3D{Z: 0.4}
	solid := &model3d.SubtractedSolid{
		Positive: &model3d.RectSolid{
			MaxVal: model3d.Coord3D{
				X: Module*(40+30) + padding*2 + HoleClearance,
				Y: Module*(40+20) + padding*2 + HoleClearance,
				Z: 0.4,
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.CylinderSolid{
				P1:     c1,
				P2:     c1.Add(thickness),
				Radius: HoleSize,
			},
			&model3d.CylinderSolid{
				P1:     c2,
				P2:     c2.Add(thickness),
				Radius: HoleSize,
			},
			&model3d.CylinderSolid{
				P1:     c3,
				P2:     c3.Add(thickness),
				Radius: HoleSize,
			},
		},
	}
	mesh := model3d.SolidToMesh(solid, 0.015, 0, -1, 5)
	mesh.SaveGroupedSTL("holder.stl")
}
