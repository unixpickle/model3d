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

	// Set to non-zero to put a hole in the front part of
	// the holder to allow the bottom of the screen to be
	// seen and touched.
	MiddleHoleSize = 0.0
)

func main() {
	solid := model3d.CheckedFuncSolid(
		model3d.Coord3D{},
		model3d.XYZ(Width, Depth, Thickness+BackHeight),
		func(c model3d.Coord3D) bool {
			if c.Z < Thickness {
				return true
			}
			offset := -(c.Z - Thickness) * Slope
			frontY := Depth + offset
			backY := Depth + offset - HolderSize - Thickness
			if c.Y < frontY && c.Y > frontY-Thickness && c.Z < FrontHeight+Thickness {
				return c.X <= Width/2-MiddleHoleSize/2 || c.X >= Width/2+MiddleHoleSize/2
			}
			if c.Y < backY && c.Y > backY-Thickness {
				return true
			}
			return false
		},
	)

	mesh := model3d.MarchingCubesSearch(solid, 0.025, 16)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("kindle_holder.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}
