package main

import (
	"flag"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	. "github.com/unixpickle/model3d/shorthand"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	Thickness           float64 `default:"10" help:"thickness of the entire bearing"`
	OuterRadius         float64 `default:"40" help:"Total radius of bearing"`
	InnerRadius         float64 `default:"20" help:"Radius at which cylinders are placed."`
	RollerLargerRadius  float64 `default:"8"`
	RollerSmallerRadius float64 `default:"4"`
	RollerInOutHeight   float64 `default:"4"`
	RollerSpaceFrac     float64 `default:"0.1"`
	RollerGap           float64 `default:"0.2"`
	Delta               float64 `default:"0.25" help:"Meshification delta"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	rollerCount := int(2 * math.Pi * args.InnerRadius / (2 * math.Max(args.RollerLargerRadius, args.RollerSmallerRadius) * (1 + args.RollerSpaceFrac)))
	rollerCenter := func(i int) C2 {
		spacing := 2 * math.Pi / float64(rollerCount)
		x := math.Cos(spacing * float64(i))
		y := math.Sin(spacing * float64(i))
		return model2d.XY(x, y).Scale(args.InnerRadius)
	}

	outerBody := model3d.Subtract(
		Cylinder(Origin3, XYZ(0, 0, args.Thickness), args.OuterRadius),
		Sub3(
			Stack(
				&model3d.ConeSlice{
					P2: XYZ(0, 0, args.RollerInOutHeight),
					R1: args.InnerRadius + args.RollerLargerRadius + args.RollerGap,
					R2: args.InnerRadius + args.RollerSmallerRadius + args.RollerGap,
				},
				Cylinder(
					Origin3,
					XYZ(0, 0, args.Thickness-args.RollerInOutHeight*2),
					args.InnerRadius+args.RollerSmallerRadius+args.RollerGap,
				),
				&model3d.ConeSlice{
					P2: XYZ(0, 0, args.RollerInOutHeight),
					R1: args.InnerRadius + args.RollerSmallerRadius + args.RollerGap,
					R2: args.InnerRadius + args.RollerLargerRadius + args.RollerGap,
				},
			),
			Stack(
				&model3d.ConeSlice{
					P2: XYZ(0, 0, args.RollerInOutHeight),
					R1: args.InnerRadius - args.RollerLargerRadius - args.RollerGap,
					R2: args.InnerRadius - args.RollerSmallerRadius - args.RollerGap,
				},
				Cylinder(
					Origin3,
					XYZ(0, 0, args.Thickness-args.RollerInOutHeight*2),
					args.InnerRadius-args.RollerSmallerRadius-args.RollerGap,
				),
				&model3d.ConeSlice{
					P2: XYZ(0, 0, args.RollerInOutHeight),
					R1: args.InnerRadius - args.RollerSmallerRadius - args.RollerGap,
					R2: args.InnerRadius - args.RollerLargerRadius - args.RollerGap,
				},
			),
		),
	)

	var rollers model3d.JoinedSolid
	for i := 0; i < rollerCount; i++ {
		center := rollerCenter(i)
		rollers = append(rollers, Stack(
			&model3d.ConeSlice{
				P1: XYZ(center.X, center.Y, 0),
				P2: XYZ(center.X, center.Y, args.RollerInOutHeight),
				R1: args.RollerLargerRadius,
				R2: args.RollerSmallerRadius,
			},
			Cylinder(
				XYZ(center.X, center.Y, 0),
				XYZ(center.X, center.Y, args.Thickness-args.RollerInOutHeight*2),
				args.RollerSmallerRadius,
			),
			&model3d.ConeSlice{
				P1: XYZ(center.X, center.Y, 0),
				P2: XYZ(center.X, center.Y, args.RollerInOutHeight),
				R1: args.RollerSmallerRadius,
				R2: args.RollerLargerRadius,
			},
		))
	}

	rollerMesh := model3d.DualContour(rollers, args.Delta, true, false)
	rollerMesh = rollerMesh.EliminateCoplanar(1e-5)

	bodyMesh := model3d.DualContour(outerBody, args.Delta, true, false)
	bodyMesh = bodyMesh.EliminateCoplanar(1e-5)

	bodyMesh.AddMesh(rollerMesh)
	bodyMesh.SaveGroupedSTL("bearing.stl")
}
