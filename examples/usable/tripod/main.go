package main

import (
	"flag"
	"log"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	ScrewRadius = 0.2
	ScrewGroove = 0.05
	ScrewSlack  = 0.02

	CradleThickness       = 0.15
	CradleBottomThickness = 0.3
	CradleSideHeight      = 0.5
	CradleHeightSlack     = 0.5
	CradleBaseSide        = 0.4

	VerticalHolderWidth = 0.5

	TripodHeight        = 5.5
	TripodHeadRadius    = 0.45
	TripodHeadHeight    = 1.0
	TripodHeadZ         = TripodHeight - 0.5
	TripodLegRadius     = 0.4
	TripodLegSpanRadius = 3.0
	TripodFootOutset    = 0.3
	TripodFootRadius    = 0.5
	TripodFootHeight    = 0.7
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
	mesh = mesh.Transform(ax.Inverse())
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving cradle mesh...")
	mesh.SaveGroupedSTL("cradle.stl")
	log.Println("Rendering cradle...")
	render3d.SaveRandomGrid("rendering_cradle.png", mesh, 3, 3, 300, nil)

	ax.Min = TripodFootHeight
	ax.Max = TripodHeadZ - 0.1
	ax.Ratio = 0.2
	tripod := CreateTripod()
	log.Println("Creating tripod mesh...")
	mesh = model3d.MarchingCubesConj(tripod, 0.01, 8, ax)
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving tripod mesh...")
	mesh.SaveGroupedSTL("tripod.stl")
	log.Println("Rendering tripod...")
	render3d.SaveRandomGrid("rendering_tripod.png", mesh, 3, 3, 300, nil)

	log.Println("Creating foot mesh...")
	mesh = model3d.MarchingCubesSearch(CreateFoot(), 0.01, 8)
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving foot mesh...")
	mesh.SaveGroupedSTL("foot.stl")
	log.Println("Rendering foot...")
	render3d.SaveRandomGrid("rendering_foot.png", mesh, 3, 3, 300, nil)
}
