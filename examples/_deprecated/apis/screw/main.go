package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	screw := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Coord3D{Z: 0.2},
			Radius: 0.2,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: 0.2},
			P2:         model3d.Coord3D{Z: 1.0},
			Radius:     0.14,
			GrooveSize: 0.05,
		},
	}
	mesh := model3d.MarchingCubesSearch(screw, 0.004, 8)
	ioutil.WriteFile("screw.stl", mesh.EncodeSTL(), 0755)

	hole := &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Coord3D{Z: 1.0},
			Radius: 0.4,
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: 0.0},
			P2:         model3d.Coord3D{Z: 1.0},
			Radius:     0.16,
			GrooveSize: 0.05,
		},
	}
	mesh = model3d.MarchingCubesSearch(hole, 0.005, 8)
	ioutil.WriteFile("hole.stl", mesh.EncodeSTL(), 0755)
}
