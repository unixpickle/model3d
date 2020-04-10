package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d"
)

const (
	Size       = 1.0
	FlatLength = 1.0
	PointSize  = 0.3

	ScrewRadius = 0.2
	ScrewGroove = 0.05
	ScrewSlack  = 0.04
	ScrewLength = 0.5

	HandleLength = 0.5
)

func main() {
	solid := &model3d.SubtractedSolid{
		Positive: BodySolid{},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{},
			P2:         model3d.Coord3D{Z: ScrewLength + ScrewRadius},
			Radius:     ScrewRadius,
			GrooveSize: ScrewGroove,
			Pointed:    true,
		},
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	log.Println("Eliminate co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving body mesh...")
	mesh.SaveGroupedSTL("dreidel.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)

	handleSolid := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.Coord3D{Z: -HandleLength},
			P2:     model3d.Coord3D{},
			Radius: ScrewRadius,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{},
			P2:         model3d.Coord3D{Z: ScrewLength - ScrewSlack},
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGroove,
		},
	}
	log.Println("Creating handle...")
	mesh = model3d.MarchingCubesSearch(handleSolid, 0.005, 8)
	mesh.SaveGroupedSTL("handle.stl")
}

type BodySolid struct{}

func (b BodySolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -Size / 2, Y: -Size / 2}
}

func (b BodySolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Size / 2, Y: Size / 2, Z: FlatLength + PointSize}
}

func (b BodySolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	sideSize := Size / 2 * math.Pow((FlatLength+PointSize-c.Z)/PointSize, 0.5)
	return c.Coord2D().Norm() <= sideSize
}
