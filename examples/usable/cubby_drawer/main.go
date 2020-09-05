package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
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
			&PoleCutout{
				X:      width / 2,
				Z:      height / 2,
				Y1:     -1e-5,
				Y2:     thickness + 1e-5,
				Radius: KnobPoleHoleRadius + KnobPoleSlack,
			},
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
	binMesh = binMesh.MapCoords(ax.Inverse().Apply)
	FinalizeMesh(binMesh, "bin")

	log.Println("Creating knob mesh...")
	knobSolid := model3d.StackSolids(
		&model3d.Cylinder{
			P2:     model3d.Z(KnobThickness),
			Radius: KnobRadius,
		},
		&model3d.Cylinder{
			P2:     model3d.Z(KnobLength),
			Radius: KnobPoleRadius,
		},
		&model3d.Cylinder{
			P2:     model3d.Z(thickness),
			Radius: KnobPoleHoleRadius,
		},
		&toolbox3d.ScrewSolid{
			P2:         model3d.Z(KnobNutThickness),
			Radius:     KnobPoleHoleRadius,
			GrooveSize: KnobScrewGrooveSize,
		},
	)
	knobMesh := model3d.MarchingCubesSearch(knobSolid, 0.01, 8)
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
			Radius:     KnobPoleHoleRadius + KnobScrewGrooveSize,
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

type PoleCutout struct {
	X      float64
	Z      float64
	Y1     float64
	Y2     float64
	Radius float64
}

func (p *PoleCutout) Min() model3d.Coord3D {
	return model3d.XYZ(p.X-p.Radius, p.Y1, p.Z-p.Radius)
}

func (p *PoleCutout) Max() model3d.Coord3D {
	return model3d.XYZ(p.X+p.Radius, p.Y2, p.Z+p.Radius*2)
}

func (p *PoleCutout) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	c2 := c.XZ().Sub(model2d.XY(p.X, p.Z))
	if c2.Norm() < p.Radius {
		return true
	}

	// A triangular top of the hole for printing without
	// support structures.
	if c2.Y < p.Radius/math.Sqrt2 {
		return false
	}
	triangleVec := model2d.XY(math.Sqrt2, math.Sqrt2).Scale(0.5)
	if c2.Dot(triangleVec) > p.Radius {
		return false
	}
	triangleVec.X *= -1
	if c2.Dot(triangleVec) > p.Radius {
		return false
	}
	return true
}
