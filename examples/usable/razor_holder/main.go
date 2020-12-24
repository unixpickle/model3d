package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Width        = 2.5
	StickyHeight = 1.2
	HolderHeight = 0.3
	HolderLength = 1.0
	Thickness    = 0.2

	GapWidth = 0.75
)

func main() {
	rs := toolbox3d.NewRectSet()
	rs.Add(&model3d.Rect{
		MaxVal: model3d.XYZ(Width, HolderLength, Thickness),
	})
	rs.Add(&model3d.Rect{
		MaxVal: model3d.XYZ(Width, Thickness, StickyHeight),
	})
	rs.Add(&model3d.Rect{
		MinVal: model3d.XYZ(0, HolderLength, 0),
		MaxVal: model3d.XYZ(Width, HolderLength+Thickness, Thickness+HolderHeight),
	})
	rs.Remove(&model3d.Rect{
		MinVal: model3d.XY((Width-GapWidth)/2, Thickness*2),
		MaxVal: model3d.XYZ((Width+GapWidth)/2, Thickness+HolderLength,
			StickyHeight),
	})
	mesh := rs.Mesh()
	ioutil.WriteFile("razor_holder.stl", mesh.EncodeSTL(), 0755)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200, nil)
}
