package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	Width     float64 `default:"9" usage:"interior width"`
	Height    float64 `default:"3.5" usage:"interior height"`
	Depth     float64 `default:"5" usage:"interior depth"`
	Thickness float64 `default:"0.3" usage:"box side thickness"`
	Delta     float64 `default:"0.0523123" usage:"mesh delta"`
	LidSpace  float64 `default:"0.03" usage:"gap for the lid to slide"`
	HoleWidth float64 `default:"4.5" usage:"width of hole in top"`
	HoleDepth float64 `default:"2.5" usage:"depth of hole in top"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)

	boxMin := model3d.XYZ(
		-(args.Width/2 + args.Thickness),
		-(args.Depth/2 + args.Thickness),
		-args.Thickness,
	)
	box := model3d.Subtract(
		model3d.NewRect(
			boxMin, model3d.XYZ(-boxMin.X, -boxMin.Y, args.Height+args.Thickness),
		),
		model3d.JoinedSolid{
			model3d.NewRect(
				boxMin.AddScalar(args.Thickness),
				model3d.XYZ(-boxMin.X-args.Thickness, -boxMin.Y-args.Thickness, args.Height+args.Thickness+1e-5),
			),
			CreateTop(&args, 0),
		},
	)
	lidSolid := CreateTop(&args, args.LidSpace)
	lidMesh := model3d.DualContour(lidSolid, args.Delta, true, false)
	mesh := model3d.DualContour(box, args.Delta, true, false)

	mesh = mesh.EliminateCoplanar(1e-5)
	lidMesh = lidMesh.EliminateCoplanar(1e-5)

	mesh.SaveGroupedSTL("body.stl")
	lidMesh.SaveGroupedSTL("lid.stl")
}

func CreateTop(args *Args, inset float64) model3d.Solid {
	minX := -(args.Width/2 + args.Thickness/2)
	maxX := -minX

	minX += inset
	maxX -= inset

	midZ := args.Height + args.Thickness/2
	minZ := args.Height
	maxZ := args.Height + args.Thickness

	minY := -(args.Depth/2 + args.Thickness) - 1e-5
	maxY := args.Depth / 2

	return model3d.CheckedFuncSolid(
		model3d.XYZ(minX, minY, minZ),
		model3d.XYZ(maxX, maxY, maxZ),
		func(c model3d.Coord3D) bool {
			if math.Abs(c.X) < args.HoleWidth/2 && math.Abs(c.Y) < args.HoleDepth/2 {
				return false
			}
			thickness := math.Min(c.X-minX, maxX-c.X)
			diffZ := math.Abs(c.Z - midZ)
			return diffZ < thickness
		},
	)
}
