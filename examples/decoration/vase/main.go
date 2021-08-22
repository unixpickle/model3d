package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	MaxRadius = 2.0
	Height    = 7.0

	RidgeFrequency = 10.0
	RidgeDepth     = 0.2
	RidgeSpinRate  = 0.5

	MinThickness  = 0.1
	BaseThickness = 0.2
)

func main() {
	log.Println("Creating mesh...")

	// Looks fine with lower Z-axis resolution.
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   0,
		Max:   Height,
		Ratio: 0.5,
	}

	mesh := model3d.MarchingCubesConj(VaseSolid(), 0.015, 8, ax)

	log.Println("Flattening base...")
	mesh = mesh.FlattenBase(0)

	log.Println("Saving STL...")
	mesh.SaveGroupedSTL("vase.stl")

	log.Println("Saving render...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func VaseSolid() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XY(-MaxRadius, -MaxRadius),
		model3d.XYZ(MaxRadius, MaxRadius, Height),
		func(c model3d.Coord3D) bool {
			maxRadius := RadiusFunc(c.Z)

			c2d := c.XY()
			theta := math.Atan2(c2d.Y, c2d.X) + c.Z*RidgeSpinRate

			ridgeInset := RidgeDepth * math.Pow(math.Sin(RidgeFrequency*theta), 2)
			radius := maxRadius - ridgeInset

			if c.Z < BaseThickness {
				return c2d.Norm() < radius
			}

			return c2d.Norm() < radius && c2d.Norm() > maxRadius-RidgeDepth-MinThickness
		},
	)
}

func RadiusFunc(x float64) float64 {
	xMin := -1.0
	xMax := 3.7
	yMax := 2.2
	return MaxRadius * UnscaledRadiusFunc(x/Height*(xMax-xMin)+xMin) / yMax
}

func UnscaledRadiusFunc(x float64) float64 {
	return 0.1*math.Pow(x, 3) - 0.45*math.Pow(x, 2) + 2.2
}
