package main

import (
	"github.com/unixpickle/model3d"
)

func main() {
	solid := model3d.JoinedSolid{
		&model3d.SphereSolid{Radius: 0.5},
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{X: 0, Y: 0.2, Z: 0},
			P2:     model3d.Coord3D{X: -0.5, Y: 0.5, Z: 0},
			Radius: 0.2,
		},
		&model3d.TorusSolid{
			Axis:        model3d.Coord3D{X: 1, Y: 1, Z: 1},
			OuterRadius: 0.7,
			InnerRadius: 0.1,
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	mesh.SaveGroupedSTL("output.stl")
}
