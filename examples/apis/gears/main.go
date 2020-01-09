package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const Module = 0.1 / math.Pi

func main() {
	solid := &toolbox3d.SpurGear{
		P1:      model3d.Coord3D{},
		P2:      model3d.Coord3D{Z: 0.2},
		Profile: toolbox3d.InvoluteGearProfile(20*math.Pi/180, 0.1, 30),
	}
	mesh := model3d.SolidToMesh(solid, 0.003, 0, -1, 5)
	mesh.SaveGroupedSTL("gear.stl")
}
