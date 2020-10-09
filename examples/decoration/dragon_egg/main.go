package main

import (
	"log"
	"math"
	"sort"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	solid := model3d.JoinedSolid{
		NewEggSolid(),
		// Stand
		&model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Z(0.3),
			Radius: 0.2,
		},
	}
	log.Println("Creating mesh...")

	// Fix artifacts on the base
	squeeze := &toolbox3d.AxisPinch{
		Axis:  toolbox3d.AxisZ,
		Min:   -0.01,
		Max:   0.01,
		Power: 0.25,
	}

	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(squeeze, solid), 0.005, 8)
	mesh = mesh.Transform(squeeze.Inverse())

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("egg.stl")
	log.Println("Rendering mesh...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type EggSolid struct {
	RadiusFunc     model2d.BezierCurve
	RadiusIntegral *RadiusIntegral
	MaxRadius      float64
	Height         float64

	RidgeInset float64
	RidgeFreq  float64
	RidgeFrac  model2d.BezierCurve
}

func NewEggSolid() *EggSolid {
	radius := model2d.BezierCurve{
		model2d.Coord{},
		model2d.Y(0.6),
		model2d.XY(1.0, 0.3),
		model2d.X(1.0),
	}
	maxRadius := 0.0
	for t := 0.0; t < 1.0; t += 0.001 {
		maxRadius = math.Max(maxRadius, radius.EvalX(t))
	}
	maxRadius += 0.02
	return &EggSolid{
		RadiusFunc:     radius,
		RadiusIntegral: NewRadiusIntegral(radius),
		MaxRadius:      maxRadius,
		Height:         1.0,

		RidgeInset: 0.025,
		RidgeFreq:  18.0,
		RidgeFrac:  radius,
	}
}

func (e *EggSolid) Min() model3d.Coord3D {
	return model3d.XY(-e.MaxRadius, -e.MaxRadius)
}

func (e *EggSolid) Max() model3d.Coord3D {
	return model3d.XYZ(e.MaxRadius, e.MaxRadius, e.Height)
}

func (e *EggSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(e, c) {
		return false
	}
	c2d := c.XY()
	maxRadius := e.RadiusFunc.EvalX(c.Z)
	curRadius := c2d.Norm()

	if curRadius < maxRadius-e.RidgeInset {
		return true
	} else if curRadius > maxRadius {
		return false
	}

	xTheta := math.Atan2(c2d.X, c2d.Y)
	yTheta := e.RadiusIntegral.EvalX(c.Z)
	yOffset := math.Sin(e.RidgeFreq * yTheta)
	xOffset := math.Sin(e.RidgeFreq * xTheta)
	inset := e.RidgeFrac.EvalX(c.Z) * 0.5 * e.RidgeInset * (2 - math.Abs(xOffset+yOffset))

	return curRadius < maxRadius-inset
}

type RadiusIntegral struct {
	X []float64
	Y []float64
}

func NewRadiusIntegral(radius model2d.BezierCurve) *RadiusIntegral {
	res := &RadiusIntegral{}
	dx := 0.0001
	var sum float64
	for x := 0.0; x < 1.0; x += dx {
		res.X = append(res.X, x)
		res.Y = append(res.Y, sum)
		r := radius.EvalX(x)
		if r < 1e-5 {
			r = 1e-5
		}
		sum += math.Sqrt(math.Pow(r*dx, 2)+dx*dx) / r
	}
	return res
}

func (r *RadiusIntegral) EvalX(x float64) float64 {
	min := sort.SearchFloat64s(r.X, x)
	if min == 0 {
		return r.Y[0]
	} else if min == len(r.Y) {
		return r.Y[len(r.Y)-1]
	}
	min -= 1
	prevX := r.X[min]
	nextX := r.X[min+1]
	prevFrac := (x - prevX) / (nextX - prevX)
	nextFrac := 1 - prevFrac
	return prevFrac*r.Y[min] + nextFrac*r.Y[min+1]
}
