package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	KnobPoleRadius      = 0.3
	KnobPoleHoleRadius  = 0.25
	KnobPoleSlack       = 0.03
	KnobRadius          = 1.0
	KnobLength          = 1.0
	KnobThickness       = 0.2
	KnobNutRadius       = 0.5
	KnobNutThickness    = 0.3
	KnobScrewGrooveSize = 0.05
	KnobScrewSlack      = 0.02
)

func main() {
	var width float64
	var height float64
	var depth float64
	var thickness float64

	flag.Float64Var(&width, "width", 5.0, "width of the drawer in inches")
	flag.Float64Var(&height, "height", 5.0, "height of the drawer in inches")
	flag.Float64Var(&depth, "depth", 7.0, "depth of the drawer in inches")
	flag.Float64Var(&thickness, "thickness", 0.2, "thickness of the drawer sides")

	flag.Parse()

	binSolid := &model3d.SubtractedSolid{
		Positive: &model3d.Rect{MaxVal: model3d.XYZ(width, depth, height)},
		Negative: model3d.JoinedSolid{
			&model3d.Rect{
				MinVal: model3d.XYZ(thickness, thickness, thickness),
				MaxVal: model3d.XYZ(width-thickness, depth-thickness, height+1e-5),
			},
			toolbox3d.Teardrop3D(
				model3d.XYZ(width/2, -1e-5, height/2),
				model3d.XYZ(width/2, thickness+1e-5, height/2),
				KnobPoleHoleRadius+KnobPoleSlack,
			),
		},
	}

	log.Println("Creating bin mesh...")
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisY,
		Min:   thickness + 0.1,
		Max:   depth - (thickness + 0.1),
		Ratio: 0.025,
	}
	binMesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, binSolid), 0.015, 8)
	binMesh = binMesh.Transform(ax.Inverse())
	FinalizeMesh(binMesh, "bin")

	log.Println("Creating knob mesh...")
	knobSolid := model3d.StackSolids(
		&model3d.Cylinder{
			P2:     model3d.Z(KnobThickness),
			Radius: KnobRadius,
		},
		toolbox3d.Teardrop3D(
			model3d.Coord3D{},
			model3d.Z(KnobLength),
			KnobPoleRadius,
		),
		toolbox3d.Teardrop3D(
			model3d.Coord3D{},
			model3d.Z(thickness),
			KnobPoleHoleRadius,
		),
		&toolbox3d.ScrewSolid{
			P2:         model3d.Z(KnobNutThickness),
			Radius:     KnobPoleHoleRadius,
			GrooveSize: KnobScrewGrooveSize,
		},
	)
	squeeze := toolbox3d.NewSmartSqueeze(toolbox3d.AxisZ, 0, 0.02, 0)
	squeeze.AddUnsqueezable(KnobThickness+KnobLength-0.02, knobSolid.Max().Z)
	squeeze.AddPinch(0)
	squeeze.AddPinch(KnobThickness)
	knobMesh := squeeze.MarchingCubesSearch(knobSolid, 0.01, 8)
	FinalizeMesh(knobMesh, "knob")

	log.Println("Creating nut mesh...")
	nutSolid := &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P2:     model3d.Z(KnobNutThickness),
			Radius: KnobNutRadius,
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Z(-1e-5),
			P2:         model3d.Z(KnobNutThickness + 1e-5),
			Radius:     KnobPoleHoleRadius + KnobScrewSlack,
			GrooveSize: KnobScrewGrooveSize,
		},
	}
	nutMesh := model3d.MarchingCubesSearch(nutSolid, 0.01, 8)
	FinalizeMesh(nutMesh, "nut")
}

func FinalizeMesh(mesh *model3d.Mesh, name string) {
	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL(fmt.Sprintf("cubby_%s.stl", name))

	log.Println("Rendering...")
	render3d.SaveRandomGrid(fmt.Sprintf("rendering_%s.png", name), mesh, 3, 3, 300, nil)
}
