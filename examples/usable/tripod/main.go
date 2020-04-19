package main

import (
	"flag"
	"log"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Thickness       = 0.15
	BottomThickness = 0.3
	SideHeight      = 0.5
	HeightSlack     = 0.5
	BaseSide        = 0.4

	ScrewRadius = 0.2
	ScrewGroove = 0.05
	ScrewSlack  = 0.02

	VerticalHolderWidth = 0.5
)

type Args struct {
	PhoneWidth  float64
	PhoneHeight float64
	PhoneDepth  float64
}

func main() {
	var args Args
	flag.Float64Var(&args.PhoneWidth, "width", 6.7, "cradle width")
	flag.Float64Var(&args.PhoneHeight, "height", 3.2, "cradle height")
	flag.Float64Var(&args.PhoneDepth, "depth", 0.65, "cradle thickness")
	flag.Parse()

	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   0,
		Max:   100,
		Ratio: 1.0 / 3,
	}
	cradle := CreateCradle(&args)
	log.Println("Creating cradle mesh...")
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, cradle), 0.01, 8)
	mesh = mesh.MapCoords(ax.Inverse().Apply)
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving cradle mesh...")
	mesh.SaveGroupedSTL("cradle.stl")
	log.Println("Rendering cradle...")
	render3d.SaveRandomGrid("rendering_cradle.png", mesh, 3, 3, 300, nil)
}
