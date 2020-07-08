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

	BananaLength  = 4
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
	Curve        model2d.BezierCurve
}

func NewBananaSolid() *BananaSolid {
	return &BananaSolid{
		CrossSection: CreateSquircle(),
		Radius: model2d.BezierCurve{
			model2d.XY(0, 0.2),
			model2d.XY(0.1, 1.0),
			model2d.XY(0.2, 1.0),
			model2d.XY(0.5, 1.4),
			model2d.XY(0.6, 1.2),
			model2d.XY(0.7, 1.1),
			model2d.XY(0.9, 0.2),
			model2d.XY(0.95, 0.2),
			model2d.XY(1.0, 0.2),
		},
		Curve: model2d.BezierCurve{
			model2d.XY(0, 1.5),
			model2d.XY(0.5, 0),
			model2d.XY(0.9, 1.9),
			model2d.XY(1, 2),
		},
	}
}

func (b *BananaSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: 0, Y: -CrossSectionRadius, Z: 0}
}

func (b *BananaSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BananaLength, Y: CrossSectionRadius, Z: 3}
}

func (b *BananaSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}

	x := c.X / BananaLength

	radius := b.Radius.EvalX(x)
	if radius <= 0 {
		return false
	}

	curveZ := b.Curve.EvalX(x)

	c2d := c.YZ()
	c2d.Y -= curveZ
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
