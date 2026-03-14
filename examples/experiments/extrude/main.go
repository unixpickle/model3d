package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	shape := model2d.NewRect(model2d.Origin, model2d.XY(10, 5))
	extruded := toolbox3d.Extrude(shape, 0, 5, toolbox3d.InsetFuncSum(
		&toolbox3d.FilletInsetFunc{
			TopRadius:    0,
			BottomRadius: 2,
			Outwards:     false,
		},
		&toolbox3d.ChamferInsetFunc{TopRadius: 1, BottomRadius: 0, Outwards: false},
	))
	mesh := model3d.DualContour(extruded, 0.1, false, false)
	mesh.SaveGroupedSTL("out.stl")
}
