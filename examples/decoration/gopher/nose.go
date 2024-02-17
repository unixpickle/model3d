package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var NoseColor = render3d.NewColorRGB(0xd9/255.0, 0xce/255.0, 0x98/255.0)

func Nose() (model3d.Solid, toolbox3d.CoordColorFunc) {
	mound := &model3d.Torus{
		Center:      model3d.YZ(0.4, 1.0),
		Axis:        model3d.YZ(-0.5, 0.55).Normalize(),
		InnerRadius: 0.15,
		OuterRadius: 0.3,
	}
	dot := model3d.TranslateSolid(
		model3d.VecScaleSolid(
			&model3d.Sphere{Radius: 0.1},
			model3d.XYZ(1.0, 1.0, 0.8),
		),
		model3d.YZ(0.7, 1.35),
	)
	return model3d.JoinedSolid{mound, dot}, func(c model3d.Coord3D) render3d.Color {
		if dot.Contains(c) {
			return render3d.NewColor(0)
		} else {
			return NoseColor
		}
	}
}
