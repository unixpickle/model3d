package main

import (
	"io/ioutil"

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
	}
	mesh := model3d.SolidToMesh(solid, 0.5, 3, 0.8, 7)
	ioutil.WriteFile("output.stl", mesh.EncodeSTL(), 0755)
}
