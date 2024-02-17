package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var BodyColor = render3d.NewColorRGB(0x73/255.0, 0xce/255.0, 0xdd/255.0)

func Body() (model3d.Solid, toolbox3d.CoordColorFunc) {
	c1 := &model3d.Sphere{
		Center: model3d.Z(0.5),
		Radius: 0.15,
	}
	c2 := &model3d.Sphere{
		Center: model3d.Z(1.5),
		Radius: 0.1,
	}
	solid := model3d.MetaballSolid(nil, 0.82, c1, c2)
	squishedSolid := model3d.TransformSolid(
		&model3d.VecScale{Scale: model3d.XYZ(1.0, 0.7, 1.0)},
		solid,
	)
	return squishedSolid, toolbox3d.ConstantCoordColorFunc(BodyColor)
}
