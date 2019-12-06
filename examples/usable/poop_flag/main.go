package main

import (
	"math"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d"
)

const (
	Radius     = 1.2
	Height     = 1.2
	RippleRate = math.Pi * 7
	RippleSize = 0.02

	ScrewRadius     = 0.3
	ScrewGrooveSize = 0.05
	ScrewSlack      = 0.03

	SteakHeight = 3.0
)

func main() {
	screw := &toolbox3d.ScrewSolid{
		P1:         model3d.Coord3D{Z: -SteakHeight},
		P2:         model3d.Coord3D{Z: Height / 2},
		Radius:     ScrewRadius,
		GrooveSize: ScrewGrooveSize,
	}
	pointedScrew := PointedSolid{
		Solid:  screw,
		Center: model3d.Coord3D{Z: Height / 2},
	}
	poop := &model3d.SubtractedSolid{
		Positive: PoopSolid{},
		Negative: pointedScrew,
	}

	mesh := model3d.SolidToMesh(poop, 0.02, 1, -1, 10)
	mesh.SaveGroupedSTL("poop.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)

	screw.P1, screw.P2 = screw.P2, screw.P1
	screw.Radius -= ScrewSlack
	mesh = model3d.SolidToMesh(pointedScrew, 0.01, 0, -1, 10)
	mesh.SaveGroupedSTL("steak.stl")
	model3d.SaveRandomGrid("steak.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
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

type PointedSolid struct {
	model3d.Solid
	Center model3d.Coord3D
}

func (p PointedSolid) Contains(c model3d.Coord3D) bool {
	if !p.Solid.Contains(c) {
		return false
	}
	radius := p.Center.Z - c.Z
	if radius < 0 {
		return false
	}
	c1 := c.Sub(p.Center)
	c1.Z = 0
	return c1.Norm() <= radius
}
