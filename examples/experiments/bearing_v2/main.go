package main

import (
	"flag"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	Thickness        float64 `default:"10" help:"thickness of the entire bearing"`
	BallRadius       float64 `default:"3" help:"radius of balls"`
	OuterRadius      float64 `default:"30" help:"total radius of bearing"`
	InnerRadius      float64 `default:"15" help:"radius at which balls are placed"`
	InnerOuterGap    float64 `default:"4" help:"space between inner and outer parts"`
	BallSpacingFrac  float64 `default:"0.2"`
	BallGap          float64 `default:"0.1"`
	HoleCutoutRadius float64 `default:"4.0"`
	HoleCutoutSlack  float64 `default:"0.1"`
	Delta            float64 `default:"0.2" help:"Meshification delta"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	cutoutDisc := model3d.Subtract(
		&model3d.Cylinder{P2: model3d.Z(args.Thickness), Radius: args.InnerRadius + args.InnerOuterGap/2},
		&model3d.Cylinder{P2: model3d.Z(args.Thickness), Radius: args.InnerRadius - args.InnerOuterGap/2},
	)
	ballCutout := &model3d.Torus{
		Center:      model3d.Z(args.Thickness / 2),
		Axis:        model3d.Z(1),
		InnerRadius: args.BallRadius + args.BallGap,
		OuterRadius: args.InnerRadius,
	}

	outerBody := model3d.Subtract(
		&model3d.Cylinder{
			P2:     model3d.Z(args.Thickness),
			Radius: args.OuterRadius,
		},
		model3d.JoinedSolid{cutoutDisc, ballCutout},
	)

	bodyWithoutCutout := model3d.Subtract(
		outerBody,
		&model3d.Cylinder{
			P1:     model3d.XZ(args.InnerRadius, args.Thickness/2),
			P2:     model3d.XZ(args.OuterRadius, args.Thickness/2),
			Radius: args.HoleCutoutRadius + args.HoleCutoutSlack,
		},
	)
	cutoutPiece := model3d.IntersectedSolid{
		outerBody,
		&model3d.Cylinder{
			P1:     model3d.XZ(args.InnerRadius, args.Thickness/2),
			P2:     model3d.XZ(args.OuterRadius, args.Thickness/2),
			Radius: args.HoleCutoutRadius,
		},
	}

	rollerCount := int(2 * math.Pi * args.InnerRadius / (2 * args.BallRadius * (1 + args.BallSpacingFrac)))
	rollerCenter := func(i int) model2d.Coord {
		spacing := 2 * math.Pi / float64(rollerCount)
		x := math.Cos(spacing * float64(i))
		y := math.Sin(spacing * float64(i))
		return model2d.XY(x, y).Scale(args.InnerRadius)
	}

	rollers := model3d.NewMesh()
	for i := 0; i < rollerCount; i++ {
		center := rollerCenter(i)
		sphere := model3d.NewMeshIcosphere(
			model3d.XYZ(center.X, center.Y, args.Thickness/2),
			args.BallRadius,
			5,
		)
		rollers.AddMesh(sphere)
	}

	bodyMesh := model3d.DualContour(bodyWithoutCutout, args.Delta, true, false)
	bodyMesh = bodyMesh.EliminateCoplanar(1e-5)

	bodyMesh.AddMesh(rollers)
	bodyMesh.SaveGroupedSTL("bearing.stl")

	cutoutMesh := model3d.DualContour(cutoutPiece, args.Delta, true, false)
	cutoutMesh.SaveGroupedSTL("cutout_insert.stl")
}
