package main

import (
	"flag"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Arguments struct {
	Thickness float64 `default:"0.5"`
	Space     float64 `default:"0.02"`
	Delta     float64 `default:"0.02"`
}

func main() {
	var args Arguments
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	longSide := 4 * (args.Thickness + args.Space)
	shortSide := 2 * (args.Thickness + args.Space)
	outlineShape := &model2d.Capsule{
		P1:     model2d.X(-longSide/2 + shortSide/2),
		P2:     model2d.X(longSide/2 - shortSide/2),
		Radius: shortSide / 2,
	}
	outlineMesh2D := model2d.MarchingSquaresSearch(outlineShape, 0.01, 8)
	segments := []model3d.Segment{}
	outlineMesh2D.Iterate(func(seg *model2d.Segment) {
		segments = append(
			segments,
			model3d.NewSegment(model3d.XY(seg[0].X, seg[0].Y), model3d.XY(seg[1].X, seg[1].Y)),
		)
	})
	solid3D := toolbox3d.LineJoin(args.Thickness/2, segments...)
	mesh3D := model3d.DualContour(solid3D, args.Delta, false, false)

	allMeshes := model3d.NewMesh()
	allMeshes.AddMesh(mesh3D)
	allMeshes.AddMesh(mesh3D.Rotate(model3d.Z(1), math.Pi/2).Rotate(model3d.Y(1), math.Pi/2))
	allMeshes.AddMesh(mesh3D.Rotate(model3d.Y(1), math.Pi/2).Rotate(model3d.Z(1), math.Pi/2))

	render3d.SaveRandomGrid("rendering.png", allMeshes, 1, 1, 512, nil)
	allMeshes.SaveGroupedSTL("three_axis_chain.stl")
}
