package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
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
	mesh := model3d.MarchingCubesSearch(HolderSolid{}, 0.025, 16)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("kindle_holder.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type HolderSolid struct{}

func (h HolderSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (h HolderSolid) Max() model3d.Coord3D {
	return model3d.XYZ(Width, Depth, Thickness+BackHeight)
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
