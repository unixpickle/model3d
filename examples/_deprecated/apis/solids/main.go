package main

import (
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	solid := model3d.JoinedSolid{
		&model3d.Sphere{Radius: 0.5},
		&model3d.Cylinder{
			P1:     model3d.XYZ(0, 0.2, 0),
			P2:     model3d.XYZ(-0.5, 0.5, 0),
			Radius: 0.2,
		},
		&model3d.Torus{
			Axis:        model3d.XYZ(1, 1, 1),
			OuterRadius: 0.7,
			InnerRadius: 0.1,
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	mesh.SaveGroupedSTL("output.stl")
}
