package main

import (
	"github.com/unixpickle/model3d"
)

const (
	Width       = 4
	Depth       = 3
	HolderSize  = 0.5
	Thickness   = 0.2
	BackHeight  = 2.0
	FrontHeight = 0.5
	Slope       = 0.3
)

func main() {
	mesh := model3d.SolidToMesh(HolderSolid{}, 0.025, 0, -1, 10)
	for i := 0; i < 3; i++ {
		mesh = mesh.LassoSolid(HolderSolid{}, 0.01, 3, 300, 0.3)
	}
	mesh = mesh.EliminateCoplanar(1e-8)
	mesh.SaveGroupedSTL("kindle_holder.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

type HolderSolid struct{}

func (h HolderSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (h HolderSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Width, Y: Depth, Z: Thickness + BackHeight}
}

func (h HolderSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(h.Min()) != h.Min() || c.Max(h.Max()) != h.Max() {
		return false
	}
	if c.Z < Thickness {
		return true
	}
	offset := -(c.Z - Thickness) * Slope
	frontY := Depth + offset
	backY := Depth + offset - HolderSize - Thickness
	if c.Y < frontY && c.Y > frontY-Thickness && c.Z < FrontHeight+Thickness {
		return true
	}
	if c.Y < backY && c.Y > backY-Thickness {
		return true
	}
	return false
}