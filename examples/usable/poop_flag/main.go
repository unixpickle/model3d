package main

import (
	"math"

	"github.com/unixpickle/model3d"
)

const (
	Radius     = 1.2
	Height     = 1.2
	RippleRate = math.Pi * 7
	RippleSize = 0.02
)

func main() {
	mesh := model3d.SolidToMesh(PoopSolid{}, 0.02, 1, -1, 5)
	mesh.SaveGroupedSTL("poop.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}

type PoopSolid struct{}

func (p PoopSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -Radius, Y: -Radius}
}

func (p PoopSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Radius, Y: Radius, Z: Height}
}

func (p PoopSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(p.Min()) != p.Min() || c.Max(p.Max()) != p.Max() {
		return false
	}
	radius := (model3d.Coord2D{X: c.X, Y: c.Y}).Norm()
	ripple := RippleSize * math.Sin(radius*RippleRate)
	maxHeight := Height - (radius*Height/Radius + ripple)
	return c.Z < maxHeight
}
