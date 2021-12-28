package main

import (
	"log"
	"math"

	"github.com/unixpickle/essentials"
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

	// Configuration of base size.
	BaseEdgeRadius = 0.25
	BaseHeight     = 0.3

	// Configuration of car size.
	CarWidth      = 0.2
	CarLength     = 0.4
	CarHeight     = 0.2
	CarTireRadius = 0.07

	CableRadius = PillarWidth / 2
)

func main() {
	joined := model3d.JoinedSolid{
		CreatePillars(),
		CreateRoad(),
		CreateCables(),
		CreateBase(),
		CreateCars(),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(joined, MarchingDelta, 8)
	log.Println("Creating color func...")
	colorFn := ColorFunc()
	log.Println("Eliminating mesh...")
	pre := len(mesh.TriangleSlice())
	mesh = mesh.EliminateCoplanarFiltered(1e-8, VertexFilter(mesh, colorFn))
	post := len(mesh.TriangleSlice())
	log.Printf("Went from %d triangles to %d", pre, post)
	log.Println("Saving mesh...")
	mesh.SaveMaterialOBJ("bridge.zip", colorFn.TriangleColor)
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

func CreateBase() model3d.Solid {
	baseRect := model3d.BoundsRect(CreateRoad())
	baseRect.MinVal.Z = -(PillarBelowRoad + BaseHeight)
	baseRect.MaxVal.Z = -(PillarBelowRoad + BaseEdgeRadius) + 1e-5
	return &model3d.IntersectedSolid{
		model3d.NewColliderSolidInset(baseRect, -BaseEdgeRadius),
		model3d.NewRect(
			model3d.XYZ(math.Inf(-1), math.Inf(-1), baseRect.Min().Z),
			model3d.Ones(math.Inf(1)),
		),
	}
}

func CreateCars() model3d.Solid {
	var cars model3d.JoinedSolid
	for _, pos := range carPositions() {
		cars = append(cars, createCar(pos))
	}
	return cars
}

func carPositions() []model2d.Coord {
	return []model2d.Coord{
		model2d.XY(0.25, 0.5),
		model2d.XY(-0.25, 1.0),
		model2d.XY(0.25, 2.0),
		model2d.XY(-0.25, 3.0),
		model2d.XY(0.25, 3.5),
		model2d.XY(-0.25, 4.8),
		model2d.XY(-0.25, 5.5),
		model2d.XY(0.25, 6.0),
	}
}

func carColors() []render3d.Color {
	return []render3d.Color{
		render3d.NewColorRGB(0.2, 0.1, 0.8),
		render3d.NewColorRGB(0.23, 0.62, 0.46),
		render3d.NewColorRGB(0, 0.8, 0.2),
		render3d.NewColorRGB(0.8, 0.1, 0.2),
		render3d.NewColorRGB(0.7, 0.7, 0.1),
		render3d.NewColorRGB(0.2, 0.1, 0.8),
		render3d.NewColorRGB(0.5, 0.2, 0.2),
		render3d.NewColorRGB(0.86, 0.42, 0.11),
	}
}

func createCar(pos model2d.Coord) model3d.Solid {
	baseSolid := model3d.JoinedSolid{
		model3d.NewRect(
			model3d.XYZ(-CarWidth/2, -CarLength/2, 0),
			model3d.XYZ(CarWidth/2, CarLength/2, CarHeight/2),
		),
		model3d.NewRect(
			model3d.XYZ(-CarWidth/2, -0.8*CarLength/2, CarHeight/2),
			model3d.XYZ(CarWidth/2, 0.8*CarLength/2, CarHeight),
		),
		&model3d.Cylinder{
			P1:     model3d.XYZ(-CarWidth/2, -(CarLength/2 - CarTireRadius), 0),
			P2:     model3d.XYZ(CarWidth/2, -(CarLength/2 - CarTireRadius), 0),
			Radius: CarTireRadius,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(-CarWidth/2, CarLength/2-CarTireRadius, 0),
			P2:     model3d.XYZ(CarWidth/2, CarLength/2-CarTireRadius, 0),
			Radius: CarTireRadius,
		},
	}
	return model3d.TranslateSolid(
		baseSolid,
		model3d.XYZ(pos.X, pos.Y, RoadThickness/2+CarTireRadius/2+RoadSmoothRadius),
	)
}

func ColorFunc() toolbox3d.CoordColorFunc {
	pillars1 := createPillarsAtY(Pillar1Y)
	pillars2 := createPillarsAtY(Pillar2Y)
	pillarColor := render3d.NewColorRGB(1.0, 0.28*1.2, 0.1*1.2)

	road := CreateRoad()
	roadColor := func(c model3d.Coord3D) render3d.Color {
		cx := math.Abs(c.X)
		if cx >= RoadWidth/2-RoadSmoothRadius {
			return pillarColor
		} else if cx < RoadLineWidth && math.Mod(c.Y, 0.5) < 0.35 {
			return render3d.NewColor(1.0)
		} else {
			return render3d.NewColor(0.2)
		}
	}

	base := CreateBase()
	baseColor := render3d.NewColorRGB(0, 0.5, 0.8)

	cables1 := createCablesAtX(cableXs()[0])
	cables2 := createCablesAtX(cableXs()[1])
	cableColor := pillarColor

	coloredObjs := []interface{}{
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
		model3d.MarchingCubesSearch(base, ColorMarchingDelta, 8),
		baseColor,
	}

	for i, pos := range carPositions() {
		coloredObjs = append(
			coloredObjs,
			model3d.MarchingCubesSearch(createCar(pos), ColorMarchingDelta, 8),
			carColors()[i],
		)
	}

	return toolbox3d.JoinedCoordColorFunc(coloredObjs...).Cached()
}

func VertexFilter(m *model3d.Mesh, f toolbox3d.CoordColorFunc) func(model3d.Coord3D) bool {
	vertices := m.VertexSlice()
	safeDelete := make([]bool, len(vertices))
	essentials.ConcurrentMap(0, len(vertices), func(i int) {
		safeDelete[i] = consistentColor(m, vertices[i], f)
	})
	deleteMap := map[model3d.Coord3D]bool{}
	for i, flag := range safeDelete {
		if flag {
			deleteMap[vertices[i]] = true
		}
	}
	return func(c model3d.Coord3D) bool {
		return deleteMap[c]
	}
}

func consistentColor(m *model3d.Mesh, c model3d.Coord3D, f toolbox3d.CoordColorFunc) bool {
	centerColor := f(c)
	for _, t := range m.Find(c) {
		for _, c1 := range t {
			if c1 != c {
				otherColor := f(c1)
				if otherColor != centerColor {
					return false
				}
			}
		}
	}
	return true
}
