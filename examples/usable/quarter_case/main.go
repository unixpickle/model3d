package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Height        = 3.0
	InnerDiameter = 1.0
	OuterDiameter = 1.2

	LidHeight        = 0.4
	LidScrewHeight   = 0.2
	LidScrewDiameter = 1.1
	LidScrewSlack    = 0.03
	LidScrewGroove   = 0.05
	LidSides         = 6

	PreviewZ1        = 0.4
	PreviewZ2        = 2.5
	PreviewWidth     = 0.1
	PreviewBarSpace  = 0.4
	PreviewBarHeight = 0.1
)

func main() {
	lidBase := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Coord3D{Z: -1},
			Max:    0,
		},
		&model3d.LinearConstraint{
			Normal: model3d.Z(1),
			Max:    LidHeight,
		},
	}
	for i := 0; i < LidSides; i++ {
		theta := math.Pi * 2 * float64(i) / LidSides
		lidBase = append(lidBase, &model3d.LinearConstraint{
			Normal: model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta)},
			Max:    OuterDiameter / 2,
		})
	}
	lidSolid := &model3d.StackedSolid{
		lidBase.Solid(),
		&toolbox3d.ScrewSolid{
			P2:         model3d.Z(LidScrewHeight),
			Radius:     LidScrewDiameter/2 - LidScrewSlack,
			GrooveSize: LidScrewGroove,
		},
	}
	bodySolid := &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P2:     model3d.Z(Height),
			Radius: OuterDiameter / 2,
		},
		Negative: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Coord3D{Z: OuterDiameter - InnerDiameter},
				P2:     model3d.Z(Height),
				Radius: InnerDiameter / 2,
			},
			&toolbox3d.ScrewSolid{
				P1:         model3d.Coord3D{Z: Height - (LidScrewHeight + LidScrewSlack)},
				P2:         model3d.Z(Height),
				Radius:     LidScrewDiameter / 2,
				GrooveSize: LidScrewGroove,
			},
			PreviewCutout{},
		},
	}

	log.Println("Creating lid mesh...")
	lidMesh := model3d.MarchingCubesSearch(lidSolid, 0.005, 8)
	log.Println("Saving lid mesh...")
	lidMesh.SaveGroupedSTL("qh_lid.stl")

	log.Println("Creating body mesh...")
	bodyMesh := model3d.MarchingCubesSearch(bodySolid, 0.005, 8)
	log.Println("Saving body mesh...")
	bodyMesh.SaveGroupedSTL("qh_body.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_lid.png", lidMesh, 3, 3, 300, nil)
	render3d.SaveRandomGrid("rendering_body.png", bodyMesh, 3, 3, 300, nil)
}

type PreviewCutout struct{}

func (p PreviewCutout) Min() model3d.Coord3D {
	return model3d.XYZ(-OuterDiameter/2, -OuterDiameter/2, PreviewZ1)
}

func (p PreviewCutout) Max() model3d.Coord3D {
	return model3d.XYZ(OuterDiameter/2, OuterDiameter/2, PreviewZ2)
}

func (p PreviewCutout) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	if math.Mod(c.Z+c.X, PreviewBarSpace) < PreviewBarHeight {
		return false
	}
	width := math.Min(PreviewWidth, math.Abs(c.Z-p.Max().Z))
	return math.Abs(c.X) < width
}
