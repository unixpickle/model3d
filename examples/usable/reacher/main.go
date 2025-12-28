package main

import (
	"flag"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	// Control for rod
	Height      float64 `default:"240" help:"Height of each rod, in mm."`
	Radius      float64 `default:"20" help:"Radius of each rod, in mm."`
	Thickness   float64 `default:"4" help:"Thickness of rod sides, in mm."`
	GrooveSize  float64 `default:"4" help:"Size of screw groove, in mm."`
	GrooveCount float64 `default:"4" help:"Approx. number of grooves per screw."`
	ScrewSlack  float64 `default:"0.5" help:"Slack for space around screws."`
	Delta       float64 `default:"0.5" help:"Dual contouring delta."`

	// Control for paddle attachment
	PaddleWidth     float64 `default:"80"`
	PaddleHeight    float64 `default:"60"`
	PaddleThickness float64 `default:"5"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	screwPosHeight := args.GrooveCount * args.GrooveSize
	maxRad := args.Radius + args.Thickness
	rodBody := &model3d.ConeSlice{
		P1: model3d.Z(0),
		P2: model3d.Z(args.Height - screwPosHeight),
		R1: args.Radius + args.Thickness,
		R2: args.Radius,
	}

	posScrew := &toolbox3d.ScrewSolid{
		P1:         model3d.Z(args.Height - screwPosHeight - 1e-5),
		P2:         model3d.Z(args.Height),
		Radius:     args.Radius - args.ScrewSlack,
		GrooveSize: args.GrooveSize,
	}
	negScrew := &toolbox3d.ScrewSolid{
		P1:         model3d.Origin,
		P2:         model3d.Z(screwPosHeight + args.Radius),
		Radius:     args.Radius,
		GrooveSize: args.GrooveSize,
		Pointed:    true,
	}
	negHole := &model3d.Cylinder{
		P1:     model3d.Origin,
		P2:     model3d.Z(args.Height),
		Radius: args.Radius - (args.Thickness + args.GrooveSize),
	}

	fullSolid := model3d.Subtract(
		model3d.JoinedSolid{rodBody, posScrew},
		model3d.JoinedSolid{negScrew, negHole},
	)
	mesh := model3d.DualContour(fullSolid, args.Delta, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("rod.stl")

	paddleCylHeight := (args.GrooveCount + 1) * args.GrooveSize
	paddleCylOuter := &model3d.Cylinder{
		P1:     model3d.Origin,
		P2:     model3d.Z(paddleCylHeight),
		Radius: maxRad,
	}
	paddleCylNeg := &toolbox3d.ScrewSolid{
		P1:         model3d.Origin,
		P2:         model3d.Z(paddleCylHeight),
		Radius:     args.Radius,
		GrooveSize: args.GrooveSize,
	}
	panelRect := model3d.CheckedFuncSolid(
		model3d.XYZ(-args.PaddleWidth/2, -args.PaddleThickness/2, 0),
		model3d.XYZ(args.PaddleWidth/2, args.PaddleThickness/2, args.PaddleHeight),
		func(c model3d.Coord3D) bool {
			if c.X <= -maxRad || c.X >= maxRad {
				return true
			}
			return (c.Z - paddleCylHeight) > (maxRad - math.Abs(c.X))
		},
	)
	paddleSolid := model3d.JoinedSolid{
		model3d.Subtract(paddleCylOuter, paddleCylNeg),
		panelRect,
	}
	mesh = model3d.DualContour(paddleSolid, args.Delta, true, false)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("paddle.stl")
}
