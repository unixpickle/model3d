package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	EggHeight = 1.0
	Period    = 0.05
	Thickness = 0.02
)

func main() {
	solid := model3d.JoinedSolid{
		model3d.IntersectedSolid{
			CreateEgg(),
			PlaneSolid(),
		},
		// Stand
		&model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Z(0.3),
			Radius: 0.2,
		},
	}
	log.Println("Creating mesh...")

	// Fix artifacts on the base
	squeeze := &toolbox3d.AxisPinch{
		Axis:  toolbox3d.AxisZ,
		Min:   -0.01,
		Max:   0.01,
		Power: 0.25,
	}

	mesh := model3d.MarchingCubesConj(solid, 0.005, 8, squeeze)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("egg.stl")
	log.Println("Rendering mesh...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func CreateEgg() model3d.Solid {
	radius := model2d.BezierCurve{
		model2d.Coord{},
		model2d.Y(0.6),
		model2d.XY(1.0, 0.3),
		model2d.X(1.0),
	}
	maxRadius := 0.0
	for t := 0.0; t < 1.0; t += 0.001 {
		maxRadius = math.Max(maxRadius, radius.EvalX(t))
	}
	return model3d.CheckedFuncSolid(
		model3d.XY(-maxRadius, -maxRadius),
		model3d.XYZ(maxRadius, maxRadius, EggHeight),
		func(c model3d.Coord3D) bool {
			rad := radius.EvalX(c.Z / EggHeight)
			return c.XY().Norm() < rad
		},
	)
}

func PlaneSolid() model3d.Solid {
	axes := []model3d.Coord3D{
		model3d.XYZ(1, 1, 1),
		model3d.XYZ(-2, 1, 1).Normalize(),
	}
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1, -1, -1),
		model3d.XYZ(1, 1, 1),
		func(c model3d.Coord3D) bool {
			for _, axis := range axes {
				dot := axis.Dot(c)
				dot = math.Mod(dot+1000.0, Period)
				if dot < Thickness/2 || dot > Period-Thickness/2 {
					return true
				}
			}
			return false
		},
	)
}
