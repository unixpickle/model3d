package main

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func Holeify(obj model3d.Solid, bounds model3d.Bounder, axis model3d.Coord3D) model3d.Solid {
	var holes []*Hole
SampleLoop:
	for i := 0; i < 20; i++ {
		radius := 0.1 + rand.Float64()*0.3
		margin := model3d.XZ(radius+0.1, radius+0.1)
		pos := model3d.NewCoord3DRandBounds(
			bounds.Min().Add(margin),
			bounds.Max().Sub(margin),
		)
		pos = pos.ProjectOut(axis)
		for _, h := range holes {
			if h.center.Dist(pos) < 0.1+h.radius+radius {
				i--
				continue SampleLoop
			}
		}
		holes = append(holes, NewHoleRandomized(pos, axis, radius, Thickness))
	}

	var negative model3d.JoinedSolid
	var positive model3d.JoinedSolid
	for _, h := range holes {
		negative = append(negative, h.Subtracted())
		positive = append(positive, h.Added())
	}

	return model3d.JoinedSolid{
		&model3d.SubtractedSolid{
			Positive: obj,
			Negative: negative.Optimize(),
		},
		positive.Optimize(),
	}
}

type Hole struct {
	center    model3d.Coord3D
	radius    float64
	thickness float64

	// Local coordinate system of the hole.
	axis   model3d.Coord3D
	basis1 model3d.Coord3D
	basis2 model3d.Coord3D

	// Configuration of the lining of the hole.
	liningCurve       model2d.BezierCurve
	liningInnerRadius float64
	liningMaxY        float64
}

func NewHoleRandomized(center, axis model3d.Coord3D, radius, thickness float64) *Hole {
	h := &Hole{
		center:    center,
		radius:    radius,
		thickness: thickness,
		axis:      axis.Normalize(),
	}
	h.basis1, h.basis2 = h.axis.OrthoBasis()

	h.liningInnerRadius = math.Min(radius-0.03, radius*(0.7+rand.Float64()*0.2))
	h.liningCurve = model2d.BezierCurve{
		model2d.XY(0, 0),
		model2d.XY(0, thickness/2*(1+rand.Float64())),
		model2d.XY(0.3+rand.Float64()*0.6, thickness/2),
		model2d.XY(1, thickness/2),
	}

	h.liningMaxY = thickness

	// Use calculus to find maximum of curve (yay!).
	yAsFuncOfT := h.liningCurve.Polynomials()[1]
	yAsFuncOfT.Derivative().IterRealRoots(func(t float64) bool {
		t = math.Max(0.0, math.Min(1.0, t))
		y := yAsFuncOfT.Eval(t)
		h.liningMaxY = math.Max(h.liningMaxY, y)
		return true
	})

	return h
}

func (h *Hole) Subtracted() *model3d.Cylinder {
	return &model3d.Cylinder{
		P1:     h.center.Sub(h.axis.Scale(h.thickness/2 + 1e-5)),
		P2:     h.center.Add(h.axis.Scale(h.thickness/2 + 1e-5)),
		Radius: h.radius,
	}
}

func (h *Hole) Added() model3d.Solid {
	bounder := &model3d.Cylinder{
		P1:     h.center.Sub(h.axis.Scale(h.liningMaxY + 1e-5)),
		P2:     h.center.Add(h.axis.Scale(h.liningMaxY + 1e-5)),
		Radius: h.radius,
	}
	return model3d.CheckedFuncSolid(
		bounder.Min(),
		bounder.Max(),
		func(c model3d.Coord3D) bool {
			c = c.Sub(h.center)
			y := c.Dot(h.axis)
			d1 := c.Dot(h.basis1)
			d2 := c.Dot(h.basis2)
			r := math.Sqrt(d1*d1 + d2*d2)
			if r < h.liningInnerRadius || r > h.radius+1e-5 {
				return false
			}
			x := math.Max(0.0, math.Min(1.0, (r-h.liningInnerRadius)/(h.radius-h.liningInnerRadius)))
			return math.Abs(y) < h.liningCurve.EvalX(x)
		},
	)
}
