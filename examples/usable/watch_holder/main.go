package main

import (
	"flag"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var width float64
	var depth float64
	var mountHeight float64
	var edgeHeight float64
	var thickness float64
	var slitWidth float64
	flag.Float64Var(&width, "width", 3.0, "width of the holder")
	flag.Float64Var(&depth, "depth", 1.5, "depth of the space for the watch")
	flag.Float64Var(&mountHeight, "mount-height", 1.0, "height of part that mounts on wall")
	flag.Float64Var(&edgeHeight, "edge-height", 0.75, "height of the part that holds the watch in place")
	flag.Float64Var(&thickness, "thickness", 0.1, "thickness of sides")
	flag.Float64Var(&slitWidth, "slit-width", 0.2, "maximum size of the power cord")
	flag.Parse()

	rs := toolbox3d.NewRectSet()

	// Base
	rs.Add(&model3d.Rect{
		MinVal: model3d.XYZ(0, 0, -thickness),
		MaxVal: model3d.XYZ(width, depth, 0),
	})
	// Mount
	rs.Add(&model3d.Rect{
		MinVal: model3d.XYZ(0, depth, -thickness),
		MaxVal: model3d.XYZ(width, depth+thickness, mountHeight),
	})
	// Edge
	rs.Add(&model3d.Rect{
		MinVal: model3d.XYZ(0, -thickness, -thickness),
		MaxVal: model3d.XYZ(width/2-slitWidth/2, 0, edgeHeight),
	})
	rs.Add(&model3d.Rect{
		MinVal: model3d.XYZ(width/2+slitWidth/2, -thickness, -thickness),
		MaxVal: model3d.XYZ(width, 0, edgeHeight),
	})

	mesh := rs.Mesh()
	mesh.SaveGroupedSTL("watch_holder.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}
