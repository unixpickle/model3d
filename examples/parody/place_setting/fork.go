package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	ForkLength      = 5.0
	ForkDip         = 0.5
	ForkMaxWidth    = 1.0
	ForkTipWidth    = 0.2
	ForkThickness   = 0.1
	ForkProngLength = 1.5 * ForkLength / 5
	ForkNumProngs   = 4
)

type ForkSolid struct {
	HeightCurve model2d.BezierCurve
	WidthCurve  model2d.BezierCurve
	MinVal      model3d.Coord3D
	MaxVal      model3d.Coord3D
}

func NewForkSolid() *ForkSolid {
	c := model2d.BezierCurve{
		model2d.Coord{X: 0, Y: 0},
		model2d.Coord{X: ForkProngLength, Y: -ForkDip * 2},
		model2d.Coord{X: 3 * ForkLength / 5, Y: 0},
		model2d.Coord{X: 3 * ForkLength / 5, Y: 0},
		model2d.Coord{X: ForkLength, Y: 0},
	}
	var minZ float64
	for t := 0.0; t <= 1.0; t += 1.0 / 1000 {
		minZ = math.Min(minZ, c.Eval(t).Y)
	}

	wc := model2d.BezierCurve{
		model2d.Coord{X: 0, Y: ForkMaxWidth / 2},
		model2d.Coord{X: ForkProngLength, Y: ForkMaxWidth / 2},
		model2d.Coord{X: ForkProngLength, Y: ForkMaxWidth / 2},
		model2d.Coord{X: 1.1 * ForkProngLength, Y: ForkMaxWidth / 2},
		model2d.Coord{X: 1.5 * ForkProngLength, Y: ForkTipWidth / 2},
		model2d.Coord{X: 1.5 * ForkProngLength, Y: ForkTipWidth / 2},
		model2d.Coord{X: ForkLength, Y: ForkTipWidth / 2},
	}
	return &ForkSolid{
		HeightCurve: c,
		WidthCurve:  wc,
		MinVal:      model3d.XYZ(0, -ForkMaxWidth/2, minZ),
		MaxVal:      model3d.XYZ(ForkLength, ForkMaxWidth/2, ForkThickness),
	}
}

func (f *ForkSolid) Min() model3d.Coord3D {
	return f.MinVal
}

func (f *ForkSolid) Max() model3d.Coord3D {
	return f.MaxVal
}

func (f *ForkSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(f, c) {
		return false
	}
	zMin := f.HeightCurve.EvalX(c.X)
	if c.Z < zMin || c.Z > zMin+ForkThickness {
		return false
	}
	width := f.WidthCurve.EvalX(c.X)
	if math.Abs(c.Y) > width {
		return false
	}
	if c.X < ForkProngLength {
		prongWidth := width / 8 * math.Sqrt(1-(ForkProngLength-c.X)/ForkProngLength)

		for i := 0.0; i < 4; i++ {
			minY := -(width - width/8)
			maxY := -minY
			prongY := minY + i*(maxY-minY)/3
			if math.Abs(c.Y-prongY) < prongWidth {
				return true
			}
		}
		return false
	}
	return true
}
