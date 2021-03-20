package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model3d"
)

type Args struct {
	BaseRadius float64
	BaseLength float64

	PoleLength float64
	PoleMin    float64
	PoleMax    float64
	NumSizes   int

	TickThickness float64
	TickLength    float64
}

func main() {
	var a Args
	flag.Float64Var(&a.BaseRadius, "base-radius", 0.3, "radius of cylinder base")
	flag.Float64Var(&a.BaseLength, "base-length", 0.5, "length of cylinder base")
	flag.Float64Var(&a.PoleLength, "pole-length", 0.3, "length of measuring pole")
	flag.Float64Var(&a.PoleMin, "pole-min", 0.04, "minimum radius of measuring pole")
	flag.Float64Var(&a.PoleMax, "pole-max", 0.2, "maximum radius of measuring pole")
	flag.IntVar(&a.NumSizes, "num-sizes", 9, "number of sizes to try")
	flag.Float64Var(&a.TickThickness, "tick-thickness", 0.02, "thickness of counter ticks")
	flag.Float64Var(&a.TickLength, "tick-length", 0.06, "length of counter ticks")
	flag.Parse()

	if a.BaseRadius < a.PoleMax {
		panic("base must be larger than poles")
	}

	sizeIncrement := (a.PoleMax - a.PoleMin) / float64(a.NumSizes-1)
	var lastMesh *model3d.Mesh
	for i := 0; i < a.NumSizes; i++ {
		log.Printf("Creating size %d/%d", (i + 1), a.NumSizes)
		holeSize := a.PoleMin + sizeIncrement*float64(i)
		solid := model3d.JoinedSolid{
			&model3d.SubtractedSolid{
				Positive: &model3d.Cylinder{
					P2:     model3d.Z(a.BaseLength),
					Radius: a.BaseRadius,
				},
				Negative: SizeTicks(&a, i+1),
			},
			&model3d.Cylinder{
				P2:     model3d.Z(a.PoleLength + a.BaseLength),
				Radius: holeSize / 2,
			},
		}
		sq := toolbox3d.NewSmartSqueeze(toolbox3d.AxisZ, 0.05, 0.04, 0.1)
		sq.AddUnsqueezable(a.BaseLength-a.TickThickness-0.02, a.BaseLength+0.02)
		sq.AddPinch(0)
		sq.AddPinch(a.BaseLength + a.PoleLength)
		mesh := sq.MarchingCubesSearch(solid, 0.0025, 8)
		mesh = mesh.EliminateCoplanar(1e-5)
		mesh.SaveGroupedSTL(fmt.Sprintf("peg_%0.2f.stl", holeSize))
		lastMesh = mesh
	}

	log.Println("Rendering last mesh...")
	render3d.SaveRandomGrid("rendering.png", lastMesh, 3, 3, 300, nil)
}

func SizeTicks(a *Args, numTicks int) model3d.Solid {
	tickCylinder := &model3d.Cylinder{
		P1:     model3d.Z(a.BaseLength - a.TickThickness),
		P2:     model3d.Z(a.BaseLength),
		Radius: a.BaseRadius,
	}
	center := model3d.Z(a.BaseLength)

	thetaIncrement := math.Pi * 2 / float64(a.NumSizes)
	ticks := model3d.JoinedSolid{}
	for i := 0; i < numTicks; i++ {
		theta := float64(i) * thetaIncrement
		direction := model3d.XY(math.Cos(theta), math.Sin(theta))
		tick := &model3d.Cylinder{
			P1:     center.Add(direction.Scale(a.BaseRadius - a.TickLength)),
			P2:     center.Add(direction.Scale(a.BaseRadius)),
			Radius: a.TickThickness,
		}
		ticks = append(ticks, tick)
	}
	return model3d.IntersectedSolid{
		tickCylinder,
		ticks,
	}
}
