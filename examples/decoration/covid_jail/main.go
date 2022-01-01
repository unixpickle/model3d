package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	MoleculeRadius     = 1.0
	MoleculeRoughness  = 0.05
	MoleculeRoughBalls = 3000

	MarchingDelta = 0.02
)

func main() {
	solid := model3d.JoinedSolid{
		MoleculeBody(),
		Spikes(),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, MarchingDelta, 8)
	log.Println("Creating color function...")
	colorFunc := ColorFunc().Cached()
	log.Println("Rendering mesh...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
}

func MoleculeBody() model3d.Solid {
	body := &model3d.Sphere{Radius: MoleculeRadius}
	var randomPoints []model3d.Coord3D
	for i := 0; i < MoleculeRoughBalls; i++ {
		randomPoints = append(randomPoints, model3d.NewCoord3DRandUnit().Scale(MoleculeRadius))
	}
	surfaceBalls := model3d.NewCoordTree(randomPoints)
	return model3d.CheckedFuncSolid(
		body.Min().AddScalar(-MoleculeRoughness),
		body.Max().AddScalar(MoleculeRoughness),
		func(c model3d.Coord3D) bool {
			return body.Contains(c) || surfaceBalls.Dist(c) < MoleculeRoughness
		},
	)
}

func Spikes() model3d.Solid {
	curve := model2d.BezierCurve{
		model2d.XY(0.15, 0.0),
		model2d.XY(0.0, 0.1),
		model2d.XY(0.7, 0.3),
		model2d.XY(0.3, 0.5),
		model2d.XY(0.0, 0.5),
	}.Transpose()
	solids := model3d.JoinedSolid{}
	for _, v := range model3d.NewMeshIcosahedron().VertexSlice() {
		(func(v model3d.Coord3D) {
			p1 := v.Scale(MoleculeRadius)
			p2 := p1.Scale(MoleculeRadius + 0.5)
			cyl := &model3d.Cylinder{P1: p1, P2: p2, Radius: 0.4}
			solid := model3d.CheckedFuncSolid(
				cyl.Min(),
				cyl.Max(),
				func(c model3d.Coord3D) bool {
					t := v.Dot(c)
					if t < MoleculeRadius || t > MoleculeRadius+0.5 {
						return false
					}
					proj := v.Scale(t)
					dist := c.Dist(proj)
					r := curve.EvalX(t - MoleculeRadius)
					return dist < r
				},
			)
			solids = append(solids, solid)
		})(v.Normalize())
	}
	return solids
}

func ColorFunc() toolbox3d.CoordColorFunc {
	return toolbox3d.JoinedCoordColorFunc(
		model3d.MarchingCubesSearch(MoleculeBody(), MarchingDelta, 8),
		render3d.NewColor(0.7),
		model3d.MarchingCubesSearch(Spikes(), MarchingDelta, 8),
		render3d.NewColorRGB(0.7, 0.2, 0.1),
	)
}
