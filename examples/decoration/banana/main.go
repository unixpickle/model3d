package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model2d"
)

const (
	CrossSectionRadius = 0.5
	CrossSectionJitter = 0.2

	BananaLength     = 3.0
	BananaStemLength = 0.3
	BananaStemFrac   = BananaStemLength / BananaLength
	BananaStemRadius = 0.3

	BaseThickness = 0.15
)

func main() {
	solid := NewBananaSolid()

	// Create a base at the bottom of the solid.
	lowRes := model3d.MarchingCubesSearch(solid, 0.1, 8)
	min, max := lowRes.Min(), lowRes.Max()
	withBase := model3d.JoinedSolid{
		&model3d.Rect{
			MinVal: min.Sub(model3d.Z(BaseThickness / 2)),
			MaxVal: model3d.XYZ(max.X, max.Y, min.Z+BaseThickness/2),
		},
		solid,
	}

	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisX,
		Min:   0.1,
		Max:   BananaLength - 0.1,
		Ratio: 0.3,
	}
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, withBase), 0.015, 8)
	mesh = mesh.MapCoords(ax.Inverse().Apply)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
	mesh.SaveGroupedSTL("banana.stl")
}

type BananaSolid struct {
	CrossSection model2d.Solid
	Radius       model2d.BezierCurve
	Curve        *Curve

	MinVal model3d.Coord3D
	MaxVal model3d.Coord3D
}

func NewBananaSolid() *BananaSolid {
	radiusCurve := model2d.BezierCurve{
		model2d.XY(0, 0),
		model2d.XY(0, 1.5),
		model2d.XY(1-BananaStemFrac, 1.5),
		model2d.XY(1-BananaStemFrac, 0),
	}
	squircle := CreateSquircle()
	curve := NewCurve()

	maxRadius := CurveMaxY(radiusCurve) * squircle.Max().Sub(squircle.Min()).X / 2
	curveMin := curve.Min()
	curveMax := curve.Max()

	return &BananaSolid{
		CrossSection: squircle,
		Radius:       radiusCurve,
		Curve:        curve,
		MinVal:       model3d.XYZ(curveMin.X-maxRadius, -maxRadius, curveMin.Y-maxRadius),
		MaxVal:       model3d.XYZ(curveMax.X+maxRadius, maxRadius, curveMax.Y+maxRadius),
	}
}

func (b *BananaSolid) Min() model3d.Coord3D {
	return b.MinVal
}

func (b *BananaSolid) Max() model3d.Coord3D {
	return b.MaxVal
}

func (b *BananaSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}

	x, axis1, collides := b.Curve.Project(c.XZ())
	if !collides {
		return false
	}
	c2d := model2d.XY(axis1, c.Y)

	var radius float64
	if x < 1-BananaStemFrac {
		radius = b.Radius.EvalX(x)
	}
	if x > 0.5 {
		// Add in a stem.
		radius = math.Max(radius, BananaStemRadius)
	}

	if radius <= 0 {
		return false
	}

	c2d = c2d.Scale(1 / radius)
	return b.CrossSection.Contains(c2d)
}

func CreateSquircle() model2d.Solid {
	var res model2d.IntersectedSolid
	for i := 0; i < 4; i++ {
		theta := float64(i) * math.Pi / 2
		center := model2d.Coord{X: math.Cos(theta), Y: math.Sin(theta)}
		res = append(res, &model2d.Circle{
			Radius: CrossSectionRadius,
			Center: center.Scale(CrossSectionJitter),
		})
	}
	return res
}

func CurveMaxY(radiusCurve model2d.Curve) float64 {
	var max float64
	for i := 0; i <= 10000; i++ {
		c := radiusCurve.Eval(float64(i) / 1000)
		max = math.Max(max, c.Y)
	}
	// Make up for approximate results
	return max + 0.01
}
