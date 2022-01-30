package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const StemThickness = 0.2

type Stem struct {
	Solid model3d.Solid
	Tip   model3d.Coord3D
}

func NewStem(baseCenter model3d.Coord3D, height, tilt float64) *Stem {
	arcRadius := 0.5
	offset := arcRadius * math.Sin(tilt)
	segs := []model3d.Segment{}
	segs = append(segs, model3d.NewSegment(model3d.Z(-(StemThickness+0.1)), model3d.Z(height-offset)))

	pTheta := func(theta float64) model3d.Coord3D {
		x := 1 + math.Cos(math.Pi-theta)
		z := math.Sin(theta)
		return model3d.XZ(x*arcRadius, z*arcRadius+height-offset)
	}
	eps := 0.01
	for theta := 0.0; theta+eps < tilt; theta += eps {
		segs = append(segs, model3d.NewSegment(pTheta(theta), pTheta(theta+eps)))
	}
	solid := toolbox3d.LineJoin(StemThickness, segs...)

	// The tip of the stem is rounded and therefore goes
	// further than the last segment.
	finalPoint := pTheta(tilt)
	finalDir := finalPoint.Sub(pTheta(tilt - eps)).Normalize()
	tip := finalPoint.Add(finalDir.Scale(StemThickness))

	xform := model3d.JoinedTransform{
		model3d.Rotation(model3d.Z(1), math.Atan2(baseCenter.Y, baseCenter.X)),
		&model3d.Translate{Offset: baseCenter},
	}
	return &Stem{
		Solid: model3d.TransformSolid(xform, solid),
		Tip:   xform.Apply(tip),
	}
}
