package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	RootDepth     = 0.05
	PressureAngle = 20 * math.Pi / 180
	Module        = 0.05
	PoleSize      = 0.2
	HoleSize      = 0.22
)

func main() {
	// CreateGear(20, "gear_20.stl")
	// CreateGear(30, "gear_30.stl")
	// CreateGear(40, "gear_40.stl")

	CreateHolder()
}

func CreateGear(teeth int, path string) {
	solid := model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: 0.4},
			P2:     model3d.Coord3D{Z: 0.8},
			Radius: PoleSize,
		},
		&toolbox3d.SpurGear{
			P1:      model3d.Coord3D{},
			P2:      model3d.Coord3D{Z: 0.4},
			Profile: toolbox3d.InvoluteGearProfile(PressureAngle, Module, teeth),
		},
	}
	mesh := model3d.SolidToMesh(solid, 0.006, 0, -1, 5)
	mesh.SaveGroupedSTL(path)
}

func CreateHolder() {
	padding := Module * 10
	c1 := model3d.Coord3D{X: padding + Module*40/2, Y: padding + Module*40/2}
	c2 := model3d.Coord3D{X: padding + Module*40 + Module*30/2, Y: padding + Module*40/2}
	c3 := model3d.Coord3D{X: padding + Module*40/2, Y: padding + Module*40 + Module*20/2}
	thickness := model3d.Coord3D{Z: 0.4}
	solid := &model3d.SubtractedSolid{
		Positive: &model3d.RectSolid{
			MaxVal: model3d.Coord3D{
				X: Module*(40+30) + padding*2,
				Y: Module*(40+20) + padding*2,
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
