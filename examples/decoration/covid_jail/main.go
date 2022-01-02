package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	MoleculeRadius     = 1.0
	MoleculeRoughness  = 0.05
	MoleculeRoughBalls = 3000

	CageSize         = 3.7
	CageTop          = CageSize/2 - 0.2
	CageBottom       = -CageSize/2 + 0.6
	CageThickness    = 0.15
	CageBarCount     = 5
	CageBarRadius    = 0.15
	CageHolderRadius = 0.2

	MarchingDelta = 0.02
)

func main() {
	solid := model3d.JoinedSolid{
		MoleculeBody(),
		Spikes(),
		Cage(),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, MarchingDelta, 8)
	log.Println("Creating color function...")
	colorFunc := ColorFunc().Cached()
	log.Println("Saving mesh...")
	mesh.SaveMaterialOBJ("covid_jail.zip", colorFunc.TriangleColor)
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
	return model3d.IntersectedSolid{solids, internalCageBounds()}
}

func internalCageBounds() *model3d.Rect {
	return model3d.NewRect(
		model3d.XYZ(-CageSize/2, -CageSize/2, CageBottom-1e-5),
		model3d.XYZ(CageSize/2, CageSize/2, CageSize/2),
	)
}

func Cage() model3d.Solid {
	cage := model3d.JoinedSolid{
		// Bottom wall.
		model3d.NewRect(
			model3d.XYZ(-CageSize/2, -CageSize/2, CageBottom-CageThickness),
			model3d.XYZ(CageSize/2, CageSize/2, CageBottom),
		),
		// Holder attached to bottom of molecule.
		&model3d.Cylinder{
			P1:     model3d.Z(CageBottom - 1e-5),
			P2:     model3d.Z(0),
			Radius: CageHolderRadius,
		},
		// Top part for bars to intersect.
		&model3d.SubtractedSolid{
			Positive: model3d.NewRect(
				model3d.XYZ(-CageSize/2, -CageSize/2, CageTop),
				model3d.XYZ(CageSize/2, CageSize/2, CageTop+CageBarRadius*2),
			),
			Negative: model3d.NewRect(
				model3d.XYZ(-CageSize/2+CageBarRadius*2, -CageSize/2+CageBarRadius*2, math.Inf(-1)),
				model3d.XYZ(CageSize/2-CageBarRadius*2, CageSize/2-CageBarRadius*2, math.Inf(1)),
			),
		},
	}
	cagePos := func(i int) float64 {
		frac := float64(i) / (CageBarCount - 1)
		min := -CageSize/2 + CageBarRadius
		max := CageSize/2 - CageBarRadius
		return min + (max-min)*frac
	}
	addBar := func(i, j int) {
		p1 := model3d.XYZ(cagePos(i), cagePos(j), CageBottom-1e-5)
		p2 := model3d.XYZ(p1.X, p1.Y, CageTop+1e-5)
		cage = append(cage, &model3d.Cylinder{P1: p1, P2: p2, Radius: CageBarRadius})
	}
	addTopBar := func(i int) {
		p1 := model3d.XYZ(cagePos(i), -CageSize/2-1e-5, CageTop+CageBarRadius)
		p2 := model3d.XYZ(cagePos(i), CageSize/2+1e-5, CageTop+CageBarRadius)
		cage = append(cage, &model3d.Cylinder{P1: p1, P2: p2, Radius: CageBarRadius})
		p1.X, p1.Y = p1.Y, p1.X
		p2.X, p2.Y = p2.Y, p2.X
		cage = append(cage, &model3d.Cylinder{P1: p1, P2: p2, Radius: CageBarRadius})
	}
	for i := 0; i < CageBarCount; i++ {
		addBar(0, i)
		addBar(i, 0)
		addBar(CageBarCount-1, i)
		addBar(i, CageBarCount-1)
		if i > 0 && i < CageBarCount-1 {
			addTopBar(i)
		}
	}
	return cage
}

func ColorFunc() toolbox3d.CoordColorFunc {
	return toolbox3d.JoinedCoordColorFunc(
		model3d.MarchingCubesSearch(MoleculeBody(), MarchingDelta, 8),
		render3d.NewColor(0.7),
		model3d.MarchingCubesSearch(Spikes(), MarchingDelta, 8),
		render3d.NewColorRGB(0.7, 0.2, 0.1),
		model3d.MarchingCubesSearch(Cage(), MarchingDelta, 8),
		render3d.NewColor(0.1),
	)
}
