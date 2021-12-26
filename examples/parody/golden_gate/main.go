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
	// Resolutions for rendering
	MarchingDelta      = 0.02
	ColorMarchingDelta = 0.015

	// Configuration of road size
	RoadWidth         = 1.0
	RoadThickness     = 0.2
	RoadSideThickness = 0.3
	RoadSmoothRadius  = 0.03
	RoadMinY          = 0.0
	RoadMaxY          = 7.0
	RoadLineWidth     = 0.05

	// Configuration of pillar size
	PillarDepth        = 0.3
	PillarWidth        = 0.2
	PillarHeight       = 3.0
	PillarBelowRoad    = 1.0
	PillarSmoothRadius = 0.03
	Pillar1Y           = 1.5
	Pillar2Y           = 5.0

	CableRadius = PillarWidth / 2
)

func main() {
	joined := model3d.JoinedSolid{
		CreatePillars(),
		CreateRoad(),
		CreateCables(),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(joined, MarchingDelta, 8)
	log.Println("Creating color func...")
	colorFn := ColorFunc()
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFn.RenderColor)
}

func CreatePillars() model3d.Solid {
	return model3d.JoinedSolid{
		createPillarsAtY(Pillar1Y),
		createPillarsAtY(Pillar2Y),
	}
}

func createPillarsAtY(y float64) model3d.Solid {
	rs := toolbox3d.NewRectSet()

	// Left pillar
	rs.Add(model3d.NewRect(
		model3d.XYZ(-(RoadWidth/2+PillarWidth), y-PillarDepth/2, -PillarBelowRoad),
		model3d.XYZ(-RoadWidth/2, y+PillarDepth/2, PillarHeight),
	))
	// Right pillar
	rs.Add(model3d.NewRect(
		model3d.XYZ(RoadWidth/2, y-PillarDepth/2, -PillarBelowRoad),
		model3d.XYZ(RoadWidth/2+PillarWidth, y+PillarDepth/2, PillarHeight),
	))

	// Connectors
	for i := 0; i < 4; i++ {
		maxZ := PillarHeight / 4 * float64(i+1)
		minZ := maxZ - PillarDepth
		rs.Add(model3d.NewRect(
			model3d.XYZ(-RoadWidth/2, y-PillarWidth/2, minZ),
			model3d.XYZ(RoadWidth/2, y+PillarWidth/2, maxZ),
		))
	}

	mesh := rs.Mesh()
	return model3d.NewColliderSolidInset(model3d.MeshToCollider(mesh), -PillarSmoothRadius)
}

func CreateRoad() model3d.Solid {
	rs := toolbox3d.NewRectSet()

	// Road itself.
	rs.Add(model3d.NewRect(
		model3d.XYZ(-RoadWidth/2, RoadMinY, -RoadThickness/2),
		model3d.XYZ(RoadWidth/2, RoadMaxY, RoadThickness/2),
	))

	// Sides of road.
	for _, x := range []float64{-(RoadWidth/2 + PillarWidth), RoadWidth / 2} {
		rs.Add(model3d.NewRect(
			model3d.XYZ(x, RoadMinY, -RoadSideThickness/2),
			model3d.XYZ(x+PillarWidth, RoadMaxY, RoadSideThickness/2),
		))
	}

	mesh := rs.Mesh()
	return model3d.NewColliderSolidInset(model3d.MeshToCollider(mesh), -RoadSmoothRadius)
}

func CreateCables() model3d.Solid {
	var solids model3d.JoinedSolid
	for _, x := range cableXs() {
		solids = append(solids, createCablesAtX(x))
	}
	return solids
}

func cableXs() [2]float64 {
	return [2]float64{-(RoadWidth/2 + PillarWidth/2), RoadWidth/2 + PillarWidth/2}
}

func createCablesAtX(x float64) model3d.Solid {
	var solids model3d.JoinedSolid

	// First half cable.
	solids = append(solids, cableForCurve(x, model2d.BezierCurve{
		model2d.XY(RoadMinY, RoadThickness/2),
		model2d.XY(RoadMinY*0.7+Pillar1Y*0.3, 0),
		model2d.XY(Pillar1Y, PillarHeight*0.8),
		model2d.XY(Pillar1Y-PillarDepth/2+CableRadius, PillarHeight),
	}))

	// Middle dipping cable.
	solids = append(solids, cableForCurve(x, model2d.BezierCurve{
		model2d.XY(Pillar1Y, PillarHeight),
		model2d.XY(Pillar1Y*0.7+Pillar2Y*0.3, PillarHeight*0.4),
		model2d.XY(Pillar1Y*0.3+Pillar2Y*0.7, PillarHeight*0.4),
		model2d.XY(Pillar2Y, PillarHeight),
	}))

	// Second half cable.
	solids = append(solids, cableForCurve(x, model2d.BezierCurve{
		model2d.XY(RoadMaxY, RoadThickness/2),
		model2d.XY(RoadMaxY*0.7+Pillar2Y*0.3, 0),
		model2d.XY(Pillar2Y, PillarHeight*0.8),
		model2d.XY(Pillar2Y+PillarDepth/2-CableRadius, PillarHeight),
	}))

	b1 := model3d.BoundsRect(CreateRoad())
	b2 := model3d.BoundsRect(CreatePillars())
	return model3d.IntersectedSolid{
		model3d.NewRect(b1.Min().Min(b2.Min()), b1.Max().Max(b2.Max())),
		solids,
	}
}

func cableForCurve(x float64, c model2d.Curve) model3d.Solid {
	var segs []model3d.Segment
	model2d.CurveMesh(c, 100).Iterate(func(s *model2d.Segment) {
		segs = append(segs, model3d.NewSegment(
			model3d.XYZ(x, s[0].X, s[0].Y),
			model3d.XYZ(x, s[1].X, s[1].Y),
		))
	})
	return toolbox3d.LineJoin(CableRadius, segs...)
}

func ColorFunc() toolbox3d.CoordColorFunc {
	pillars1 := createPillarsAtY(Pillar1Y)
	pillars2 := createPillarsAtY(Pillar2Y)
	pillarColor := render3d.NewColorRGB(1.0, 0.62, 0)

	road := CreateRoad()
	roadColor := func(c model3d.Coord3D) render3d.Color {
		cx := math.Abs(c.X)
		if cx >= RoadWidth/2-RoadSmoothRadius {
			return pillarColor
		} else if cx < RoadLineWidth {
			return render3d.NewColor(1.0)
		} else {
			return render3d.NewColor(0.2)
		}
	}

	cables1 := createCablesAtX(cableXs()[0])
	cables2 := createCablesAtX(cableXs()[1])
	cableColor := pillarColor

	return toolbox3d.JoinedCoordColorFunc(
		model3d.MarchingCubesSearch(pillars1, ColorMarchingDelta, 8),
		pillarColor,
		model3d.MarchingCubesSearch(pillars2, ColorMarchingDelta, 8),
		pillarColor,
		model3d.MarchingCubesSearch(road, ColorMarchingDelta, 8),
		roadColor,
		model3d.MarchingCubesSearch(cables1, ColorMarchingDelta, 8),
		cableColor,
		model3d.MarchingCubesSearch(cables2, ColorMarchingDelta, 8),
		cableColor,
	)
}
