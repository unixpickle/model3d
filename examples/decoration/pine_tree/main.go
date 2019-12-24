package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
)

const (
	BaseOutset    = 0.1
	BaseThickness = 0.2

	Radius = 1.9
	Height = 4.0

	TopInset    = 0.1
	TrunkRadius = 0.3

	Slope     = 1.3
	Thickness = 0.2
)

func main() {
	log.Println("Creating mesh...")
	mesh := model3d.SolidToMesh(TreeSolid{}, 0.015, 0, -1, 30)
	log.Println("Exporting...")
	mesh.SaveGroupedSTL("pine.stl")
	log.Println("Rendering...")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

type TreeSolid struct{}

func (t TreeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -(Radius + BaseOutset), Y: -(Radius + BaseOutset), Z: -BaseThickness}
}

func (t TreeSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Radius + BaseOutset, Y: Radius + BaseOutset, Z: Height}
}

func (t TreeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(t, c) {
		return false
	}

	// Simple square base.
	if c.Z < 0 {
		return true
	}

	rad := c.Coord2D().Norm()
	maxRad := (Radius * (Height - c.Z) / Height)
	if rad >= maxRad {
		return false
	}
	if rad < math.Min(TrunkRadius, maxRad-TopInset) {
		return true
	}
	coneBaseZ := c.Z - rad*Slope

	// Add a large number to avoid skip at 0.
	return (int(coneBaseZ/Thickness+10000))%2 == 0
}
