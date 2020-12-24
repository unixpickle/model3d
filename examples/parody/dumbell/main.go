package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	SideHeight    = 2.0
	SideMinRadius = 1.5
	SideEdgeCut   = 0.3

	BarRadius      = 0.4
	BarSlack       = 0.02
	BarLength      = 5.0
	BarTotalLength = BarLength + SideHeight

	// Cause a slope of more than 45 degrees to prevent
	// support structures.
	ZWithExtraSlope = 0.999
)

func main() {
	sideWithHole := &model3d.SubtractedSolid{
		Positive: CreateSide().Solid(),
		Negative: &model3d.Cylinder{
			P1:     model3d.Z(SideHeight / 2),
			P2:     model3d.Z(SideHeight + 1e-5),
			Radius: BarRadius + BarSlack,
		},
	}
	log.Println("Creating meshes...")
	squeeze := &toolbox3d.SmartSqueeze{
		Axis:         toolbox3d.AxisZ,
		SqueezeRatio: 0.1,
		PinchRange:   0.03,
		PinchPower:   0.25,
	}
	squeeze.AddUnsqueezable(0, SideEdgeCut+0.02)
	squeeze.AddUnsqueezable(SideHeight-SideEdgeCut-0.02, SideHeight)
	squeeze.AddPinch(SideHeight / 2)
	mesh := squeeze.MarchingCubesSearch(sideWithHole, 0.01, 16)

	squeeze.Pinches = nil
	squeeze.Unsqueezable = nil
	squeeze.AddPinch(0)
	squeeze.AddPinch(BarTotalLength)
	bar := &model3d.Cylinder{
		P2:     model3d.Z(BarTotalLength),
		Radius: BarRadius,
	}
	barMesh := squeeze.MarchingCubesSearch(bar, 0.01, 16)

	log.Println("Simplifying meshes...")
	mesh = mesh.EliminateCoplanar(1e-5)
	barMesh = barMesh.EliminateCoplanar(1e-5)

	log.Println("Saving meshes...")
	mesh.SaveGroupedSTL("side.stl")
	barMesh.SaveGroupedSTL("bar.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_side.png", mesh, 3, 3, 300, nil)
	render3d.SaveRandomGrid("rendering_bar.png", barMesh, 3, 3, 300, nil)
}

func CreateSide() model3d.ConvexPolytope {
	rawShape := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Z(1),
			Max:    SideHeight,
		},
		&model3d.LinearConstraint{
			Normal: model3d.Z(-1),
			Max:    0,
		},
	}
	for i := 0; i < 6; i++ {
		theta := float64(i) / 6 * math.Pi * 2
		normal := model3d.XY(math.Cos(theta), math.Sin(theta))
		rawShape = append(rawShape, &model3d.LinearConstraint{
			Normal: normal,
			Max:    SideMinRadius,
		})

		// Do the top edge cut.
		angledUp := normal.Add(model3d.Z(ZWithExtraSlope)).Normalize()
		topPoint := normal.Scale(SideMinRadius - SideEdgeCut).Add(model3d.Z(SideHeight))
		rawShape = append(rawShape, &model3d.LinearConstraint{
			Normal: angledUp,
			Max:    angledUp.Dot(topPoint),
		})

		// Do the bottom edge cut.
		angledDown := normal.Add(model3d.Z(-ZWithExtraSlope)).Normalize()
		bottomPoint := normal.Scale(SideMinRadius - SideEdgeCut)
		rawShape = append(rawShape, &model3d.LinearConstraint{
			Normal: angledDown,
			Max:    angledDown.Dot(bottomPoint),
		})
	}
	return rawShape
}
