package main

import (
	"log"
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
	BaseWidthFrac = 0.9
)

func main() {
	banana := NewBananaSolid()

	// Create a small square base at the lowest part of
	// the model to make it stickable onto a wall.
	log.Println("Creating base...")
	minPoint := SolidMinZ(banana)
	baseSize := BaseWidthFrac * (banana.Max().Y - banana.Min().Y)
	baseMin := model3d.XYZ(minPoint.X-baseSize/2, -baseSize/2, minPoint.Z-BaseThickness/2)
	base := &model3d.Rect{
		MinVal: baseMin,
		MaxVal: baseMin.Add(model3d.XYZ(baseSize, baseSize, BaseThickness)),
	}

	solid := model3d.JoinedSolid{
		base,
		banana,
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.015, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("banana.stl")

	log.Println("Rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(3.5, -4, 2), 500, 300, nil)
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

	x, axis1 := b.Curve.Project(c.XZ())
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
	ls := &toolbox3d.LineSearch{Stops: 100, Recursions: 4}
	_, y := ls.Maximize(0, 1, func(t float64) float64 {
		return radiusCurve.Eval(t).Y
	})
	return y
}

func SolidMinZ(s model3d.Solid) model3d.Coord3D {
	lowRes := model3d.MarchingCubesSearch(s, 0.05, 8)
	var lowest model3d.Coord3D
	for i, c := range lowRes.VertexSlice() {
		if i == 0 || c.Z < lowest.Z {
			lowest = c
		}
	}
	return lowest
}
