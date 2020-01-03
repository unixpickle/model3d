package main

import (
	"log"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d"
)

const (
	Size       = 1.0
	FlatLength = 0.8
	PointSize  = 0.5

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
	mesh := model3d.SolidToMesh(solid, 0.005, 0, -1, 10)
	log.Println("Eliminate co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving body mesh...")
	mesh.SaveGroupedSTL("dreidel.stl")
	log.Println("Rendering...")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	handleSolid := model3d.JoinedSolid{
		&model3d.CylinderSolid{
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
	mesh = model3d.SolidToMesh(handleSolid, 0.005, 0, -1, 10)
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
	if !model3d.InSolidBounds(b, c) {
		return false
	}
	sideSize := Size / 2 * (FlatLength + PointSize - c.Z) / PointSize
	return c.Coord2D().Norm() <= sideSize
}
