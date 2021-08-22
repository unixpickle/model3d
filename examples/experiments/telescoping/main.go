package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model3d"
)

const (
	NotchSize     = 0.05
	TubeThickness = 0.1
	BottomNotchZ  = 0.5

	OuterRadius = 0.7
	OuterHeight = 3.0

	GapRadius = TubeThickness + NotchSize*1.5
	GapHeight = NotchSize * 2

	NumTubes = 3
)

func main() {
	radius := OuterRadius
	height := OuterHeight
	tubes := model3d.JoinedSolid{
		// Outer tube.
		&Tube{
			Radius:       radius,
			Height:       height,
			InnerNotches: []float64{height - NotchSize},
		},
	}

	// Create tubes going inward.
	for i := 0; i < NumTubes-1; i++ {
		radius -= GapRadius
		height += GapHeight
		tubes = append(tubes, &Tube{
			Radius: radius,
			Height: height,
			// The notch at 0.0 makes things jiggle less, and
			// makes the print have more surface area to stick
			// to the build plate.
			OuterNotches: []float64{0.0, BottomNotchZ, height - NotchSize},
			InnerNotches: []float64{height - NotchSize},
		})
	}

	log.Println("Creating mesh...")
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   BottomNotchZ + NotchSize*2 + 0.1,
		Max:   OuterHeight - NotchSize*2 - 0.1,
		Ratio: 0.1,
	}
	mesh := model3d.MarchingCubesConj(tubes, 0.01, 16, ax)
	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("telescoping.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type Tube struct {
	Radius       float64
	Height       float64
	InnerNotches []float64
	OuterNotches []float64
}

func (t *Tube) Min() model3d.Coord3D {
	mx := -(t.Radius + TubeThickness/2 + NotchSize)
	return model3d.XY(mx, mx)
}

func (t *Tube) Max() model3d.Coord3D {
	mx := t.Radius + TubeThickness/2 + NotchSize
	return model3d.XYZ(mx, mx, t.Height)
}

func (t *Tube) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(t, c) {
		return false
	}
	cRadius := c.XY().Norm()

	// Fast path for most of the volume.
	if cRadius < (t.Radius - TubeThickness/2 - NotchSize) {
		return false
	} else if cRadius > (t.Radius + TubeThickness/2 + NotchSize) {
		return false
	}

	minRadius := t.Radius - TubeThickness/2
	maxRadius := t.Radius + TubeThickness/2

	for _, notchZ := range t.InnerNotches {
		extraRadius := NotchSize - math.Abs(c.Z-notchZ)
		if extraRadius > 0 {
			minRadius = math.Min(minRadius, t.Radius-TubeThickness/2-extraRadius)
		}
	}

	for _, notchZ := range t.OuterNotches {
		extraRadius := NotchSize - math.Abs(c.Z-notchZ)
		if extraRadius > 0 {
			maxRadius = math.Max(maxRadius, t.Radius+TubeThickness/2+extraRadius)
		}
	}

	return cRadius >= minRadius && cRadius <= maxRadius
}
