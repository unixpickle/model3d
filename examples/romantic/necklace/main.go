package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
)

const (
	LinkWidth     = 0.2
	LinkHeight    = 0.4
	LinkThickness = 0.04
	LinkOddShift  = LinkWidth * 0.4

	TotalLength = 20

	StartRadius = 2.0
	SpiralRate  = 0.4 / (math.Pi * 2)
	MoveRate    = 0.6 * LinkHeight
)

func main() {
	log.Println("Creating link mesh...")
	solid := LinkSolid{}
	link := model3d.SolidToMesh(solid, 0.005, 0, -1, 5)
	for i := 0; i < 10; i++ {
		link = link.LassoSolid(solid, 0.005, 3, 200, 0.2)
	}
	link = link.FlattenBase(0)
	link = link.EliminateCoplanar(1e-5)
	linkOdd := link.MapCoords((model3d.Coord3D{X: LinkOddShift / 2}).Add)
	link = link.MapCoords((model3d.Coord3D{X: -LinkOddShift / 2}).Add)

	log.Println("Creating full mesh...")
	m := model3d.NewMesh()
	manifold := NewSpiralManifold(StartRadius, SpiralRate)
	for i := 0; i < int(TotalLength/LinkHeight); i++ {
		if i%2 == 0 {
			m.AddMesh(link.MapCoords(manifold.Convert))
		} else {
			m.AddMesh(linkOdd.MapCoords(manifold.Convert))
		}
		manifold.Move(MoveRate)
	}
	log.Println("Verifying mesh...")
	if m.SelfIntersections() > 0 {
		panic("self intersections detected")
	}
	if _, n := m.RepairNormals(1e-5); n != 0 {
		panic("incorrect normals")
	}
	log.Println("Saving results...")
	m.SaveGroupedSTL("necklace.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(m), 3, 3, 300, 300)
}

type LinkSolid struct{}

func (l LinkSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -LinkWidth / 2, Y: -LinkHeight / 2, Z: 0}
}

func (l LinkSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: LinkWidth / 2, Y: LinkHeight / 2, Z: LinkWidth/2 + LinkThickness}
}

func (l LinkSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(l, c) {
		return false
	}
	if c.Z < LinkThickness &&
		(c.X < -LinkWidth/2+LinkThickness || c.X > LinkWidth/2-LinkThickness) {
		return true
	}
	if c.Y > -LinkHeight/2+LinkThickness && c.Y < LinkHeight/2-LinkThickness {
		return false
	}

	height := LinkWidth/2 - math.Abs(c.X)
	return c.Z >= height && c.Z <= height+LinkThickness
}
