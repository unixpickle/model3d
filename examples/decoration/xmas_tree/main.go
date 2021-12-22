package main

import (
	"fmt"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	GrooveRatio    = 0.15
	MinZ           = -8
	MaxTwist       = 2.0
	RidgeThickness = 0.1
	FinalScale     = 0.5
)

func main() {
	outerRad := OuterRadiusFunc()
	twist := TwistFunc()
	crossSection := CrossSectionFunc()

	solid := model3d.CheckedFuncSolid(
		model3d.XYZ(-5.0, -5.0, MinZ),
		model3d.XYZ(5.0, 5.0, 0),
		func(c model3d.Coord3D) bool {
			r := c.XY().Norm() / (outerRad(c.Z) + 1e-8)
			theta := twist(c.Z) + math.Atan2(c.Y, c.X)
			return crossSection(r, theta)
		},
	)
	log.Println("Creating main mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.03, 8)
	log.Println("Saving main mesh...")
	mesh.Scale(FinalScale).SaveGroupedSTL("tree.stl")

	log.Println("Creating separate ridges...")
	ridgeMesh := model3d.NewMesh()
	for i, z := range []float64{-2.0, -3.0, -4.0, -5.0} {
		solid2d := RidgeSolid(z)
		solid3d := model3d.ProfileSolid(solid2d, 0, RidgeThickness)
		solid3d = model3d.TranslateSolid(solid3d, model3d.Z(z))
		mesh := model3d.MarchingCubesSearch(solid3d, 0.02, 8)
		mesh.Scale(FinalScale).SaveGroupedSTL(fmt.Sprintf("ridge_%d.stl", i))
		ridgeMesh.AddMesh(mesh)
	}

	log.Println("Rendering full model...")
	colorFunc := toolbox3d.JoinedCoordColorFunc(
		ridgeMesh, render3d.NewColor(0.8),
		mesh, render3d.NewColorRGB(0, 0.8, 0),
	)
	mesh.AddMesh(ridgeMesh)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
}

func OuterRadiusFunc() func(float64) float64 {
	curve := model2d.JoinedCurve{
		model2d.BezierCurve{
			model2d.XY(0, 0),
			model2d.XY(-3, 1.5),
		},
		model2d.BezierCurve{
			model2d.XY(-3, 1.5),
			model2d.XY(-6, 3),
			model2d.XY(-7, 4),
			model2d.XY(-10, 0),
		},
	}
	return func(z float64) float64 {
		return model2d.CurveEvalX(curve, z)
	}
}

func TwistFunc() func(float64) float64 {
	curve := model2d.BezierCurve{
		model2d.XY(0, 0),
		model2d.XY(0.3, 0.05),
		model2d.XY(0.8, 1),
		model2d.XY(1, 1),
	}
	return func(z float64) float64 {
		x := math.Min(1, math.Max(0, (z-MinZ)/(-MinZ)))
		return -curve.EvalX(x) * MaxTwist
	}

}

func CrossSectionFunc() func(r, twist float64) bool {
	return func(r, theta float64) bool {
		offset := math.Abs(math.Sin(theta * 10))
		rad := GrooveRatio*offset + (1 - GrooveRatio)
		return r < rad
	}
}

func RidgeSolid(z float64) model2d.Solid {
	outerRad := OuterRadiusFunc()(z)
	twist := TwistFunc()(z)

	pointFn := func(theta float64) model2d.Coord {
		p := model2d.XY(math.Cos(theta), math.Sin(theta))
		offset := math.Abs(math.Sin((theta - twist) * 10))
		rad := GrooveRatio*offset + (1 - GrooveRatio)
		return p.Scale(rad * outerRad)
	}
	mesh := model2d.NewMesh()
	for i := 0; i < 1000; i++ {
		t1 := math.Pi * 2 * float64(i) / 1000
		t2 := math.Pi * 2 * float64(i+1) / 1000
		mesh.Add(&model2d.Segment{pointFn(t1), pointFn(t2)})
	}
	sdf := model2d.MeshToSDF(mesh)

	return model2d.CheckedFuncSolid(
		model2d.XY(-5.0, -5.0),
		model2d.XY(5.0, 5.0),
		func(c model2d.Coord) bool {
			s := sdf.SDF(c)
			return s < 0 && s > -RidgeThickness
		},
	)
}
