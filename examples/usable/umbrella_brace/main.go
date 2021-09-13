package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	HoleDepth = 0.75

	// The lip that prevents the part from sliding
	// completely into the base.
	LipRadius       = 1.25
	LipMinThickness = 0.1
	LipMaxThickness = 0.25

	// Radius of part that slides into table or base.
	MinMidRadius = 1.05
	MaxMidRadius = 1.05

	// Radius of inner hole for umbrella to slide into.
	SmallRadius = 0.625
)

func main() {
	solid := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			CreateLip(),
			CreateMidHole(),
		},
		Negative: CreateSmallHole(),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving...")
	mesh.SaveGroupedSTL("umbrella_brace.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func CreateLip() model3d.Solid {
	maxZ := LipMaxThickness - LipMinThickness
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-LipRadius, -LipRadius, -LipMinThickness),
		model3d.XYZ(LipRadius, LipRadius, maxZ),
		func(c model3d.Coord3D) bool {
			r := c.XY().Norm()
			if c.Z <= 0 {
				return r < LipRadius
			}
			fracOfThickness := 1 - c.Z/maxZ
			minRad := LipRadius - fracOfThickness*(LipRadius-MaxMidRadius)
			return r > minRad && r < LipRadius
		},
	)
}

func CreateMidHole() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-MaxMidRadius, -MaxMidRadius, 0),
		model3d.XYZ(MaxMidRadius, MaxMidRadius, HoleDepth),
		func(c model3d.Coord3D) bool {
			frac := c.Z / HoleDepth
			rad := frac*MinMidRadius + (1-frac)*MaxMidRadius
			return c.XY().Norm() < rad
		},
	)
}

func CreateSmallHole() model3d.Solid {
	return &model3d.Cylinder{
		Radius: SmallRadius,
		P1:     model3d.Z(-100),
		P2:     model3d.Z(HoleDepth + 1e-5),
	}
}
