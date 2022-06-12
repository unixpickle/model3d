package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func MetalRing() (model3d.Solid, toolbox3d.CoordColorFunc) {
	r := 0.45
	thickness := 0.1
	minThickness := 0.07
	depth := 0.22
	bevelRing := model3d.CheckedFuncSolid(
		model3d.XYZ(-r-thickness, -r-thickness, 0),
		model3d.XYZ(r+thickness, r+thickness, depth),
		func(c model3d.Coord3D) bool {
			edgeDist := math.Min(depth-c.Z, c.Z)
			th := math.Min(thickness, edgeDist+minThickness)
			return math.Abs(c.XY().Norm()-(r+thickness/2)) < th/2
		},
	)
	colorFn := toolbox3d.ConstantCoordColorFunc(render3d.NewColor(0.8))
	return bevelRing, colorFn
}
