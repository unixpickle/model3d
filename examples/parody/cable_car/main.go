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
	FastMode             = false
	AxleRadius           = 0.15
	SideOverhang         = 0.1
	TopBottomThickness   = 0.15
	WheelThickness       = 0.15
	WheelRadius          = 0.35
	WheelSlip            = 0.05
	ArchInset            = 0.18
	ArchTopInset         = 0.2
	ArchPanelWidth       = 1.0
	ArchPanelHeight      = 2.6
	ArchPanelThickness   = 0.15
	ArchOutlineThickness = 0.025
	TopLipSpace          = 0.02
	ArchedTopHeight      = 0.5
)

var (
	BottomColor      = render3d.NewColor(0.1)
	CarColor         = render3d.NewColorRGB(1.0, 0.3, 0.3)
	ArchOutlineColor = render3d.NewColorRGB(1.0, 0.5, 0.0)
	SideTextColor    = render3d.NewColorRGB(1.0, 0.5, 0.0)
)

type ModelFn func() (model3d.Solid, toolbox3d.CoordColorFunc)

func main() {
	log.Println("Building base...")
	base, baseColor := BuildMesh(BaseWheels, BottomPanel, ArchSides)

	log.Println("Building top...")
	top, topColor := BuildMesh(TopPanel, ArchedTop)

	log.Println("Building full color function...")
	baseCoordTree := model3d.NewCoordTree(base.VertexSlice())
	topCoordTree := model3d.NewCoordTree(top.VertexSlice())
	fullColor := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if baseCoordTree.Dist(c) < topCoordTree.Dist(c) {
			return baseColor(c)
		} else {
			return topColor(c)
		}
	})

	log.Println("Decimating...")
	base = Decimate(base, baseColor)
	top = Decimate(top, topColor)

	log.Println("Saving...")
	base.SaveMaterialOBJ("base.zip", baseColor.TriangleColor)
	top.SaveMaterialOBJ("top.zip", topColor.TriangleColor)

	log.Println("Rendering...")
	fullMesh := base.Copy()
	fullMesh.AddMesh(top)
	render3d.SaveRandomGrid("rendering.png", fullMesh, 3, 3, 300, fullColor.RenderColor)
}

func BuildMesh(fns ...ModelFn) (*model3d.Mesh, toolbox3d.CoordColorFunc) {
	var solids model3d.JoinedSolid
	solidsAndColors := []any{}
	for _, partFn := range fns {
		solid, colorFn := partFn()
		solids = append(solids, solid)
		solidsAndColors = append(solidsAndColors, solid, colorFn)
	}
	delta := 0.015
	if FastMode {
		delta = 0.04
	}
	mesh, interior := model3d.DualContourInterior(solids.Optimize(), delta, true, false)
	colorFn := toolbox3d.JoinedSolidCoordColorFunc(interior, solidsAndColors...)
	return mesh, colorFn
}

func Decimate(m *model3d.Mesh, cf toolbox3d.CoordColorFunc) *model3d.Mesh {
	d := &model3d.Decimator{
		FeatureAngle:     0.03,
		BoundaryDistance: 1e-5,
		PlaneDistance:    1e-4,
		FilterFunc:       cf.ChangeFilterFunc(m, 0.02),
	}
	newMesh := d.Decimate(m)
	oldCount := len(m.TriangleSlice())
	newCount := len(newMesh.TriangleSlice())
	log.Printf("Went from %d triangles to %d (%f%% reduction)", oldCount, newCount,
		100*(1.0-float64(newCount)/float64(oldCount)))
	return newMesh
}

func BaseWheels() (model3d.Solid, toolbox3d.CoordColorFunc) {
	var wheels model3d.JoinedSolid
	firstX := 3.5 - WheelRadius*2
	for _, x := range []float64{-firstX, -firstX + WheelRadius*2.5, firstX - WheelRadius*2.5, firstX} {
		wheelZ := -(TopBottomThickness + WheelRadius - WheelSlip)
		for _, y := range []float64{-1.3, 1.3} {
			wheel := &model3d.Cylinder{
				P1:     model3d.XYZ(x, y-WheelThickness/2, wheelZ),
				P2:     model3d.XYZ(x, y+WheelThickness/2, wheelZ),
				Radius: WheelRadius,
			}
			wheels = append(wheels, wheel)
		}
		axle := model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.XYZ(x, -1.3, wheelZ),
				P2:     model3d.XYZ(x, 1.3, wheelZ),
				Radius: AxleRadius,
			},
			model3d.NewRect(
				model3d.XYZ(x-AxleRadius, -1.3, wheelZ),
				model3d.XYZ(x+AxleRadius, 1.3, -TopBottomThickness+1e-5),
			),
		}
		wheels = append(wheels, axle...)
	}
	return wheels.Optimize(), toolbox3d.ConstantCoordColorFunc(BottomColor)
}

func TopPanel() (model3d.Solid, toolbox3d.CoordColorFunc) {
	// Create a lip so that this lid slides into the base
	// of the structure but doesn't move around too much.
	sliceZ := ArchPanelHeight - 1e-5
	solid3d, _ := ArchSides()
	solid2d := toolbox3d.SliceSolid(solid3d, toolbox3d.AxisZ, sliceZ)
	outline2d := model2d.MarchingSquaresSearch(solid2d, 0.005, 8)
	mesh := model2d.MeshToHierarchy(outline2d)[0].Children[0].Mesh
	collider2d := model2d.MeshToCollider(mesh)
	lip2d := &model2d.SubtractedSolid{
		Positive: model2d.NewColliderSolidInset(collider2d, TopLipSpace),
		Negative: model2d.NewColliderSolidInset(collider2d, TopLipSpace+0.1),
	}
	lip3d := model3d.ProfileSolid(lip2d, sliceZ-0.1, ArchPanelHeight+1e-5)
	topSolid, colorFn := topBottomPanel(true)
	return model3d.JoinedSolid{topSolid, lip3d}, func(c model3d.Coord3D) render3d.Color {
		if lip3d.Contains(c) {
			return CarColor
		}
		return colorFn(c)
	}
}

func BottomPanel() (model3d.Solid, toolbox3d.CoordColorFunc) {
	return topBottomPanel(false)
}

func topBottomPanel(top bool) (model3d.Solid, toolbox3d.CoordColorFunc) {
	// Ends are rounded according to a virtual circle
	// with the given radius. Higher radii mean less
	// curvature.
	virtualRadius := 5.0
	maxDim := 4.0
	base := model3d.CheckedFuncSolid(
		model3d.XYZ(-maxDim, -1.5-SideOverhang, -TopBottomThickness-1e-5),
		model3d.XYZ(maxDim, 1.5+SideOverhang, 1e-5),
		func(c model3d.Coord3D) bool {
			if c.X < 0 {
				return c.XY().Dist(model2d.X(virtualRadius-maxDim)) < virtualRadius
			} else {
				return c.XY().Dist(model2d.X(-virtualRadius+maxDim)) < virtualRadius
			}
		},
	)
	if top {
		return model3d.TranslateSolid(base, model3d.Z(ArchPanelHeight+TopBottomThickness)),
			toolbox3d.ConstantCoordColorFunc(CarColor)
	} else {
		return base, toolbox3d.ConstantCoordColorFunc(BottomColor)
	}
}

func ArchSides() (model3d.Solid, toolbox3d.CoordColorFunc) {
	baseSolid := outsetArchSides(0)
	midOutset := outsetArchSides(ArchOutlineThickness)
	fullOutset := outsetArchSides(ArchOutlineThickness * 2)

	textSolid := ArchSidesText()

	return baseSolid, func(c model3d.Coord3D) render3d.Color {
		if !fullOutset.Contains(c) && midOutset.Contains(c) {
			return ArchOutlineColor
		} else if textSolid.Contains(c) {
			return SideTextColor
		} else if math.Abs(c.Y) < 1.5-ArchPanelThickness {
			return CarColor
		} else {
			return BottomColor
		}
	}
}

// outsetArchSides creates the sides of the car with arches
// cut out. The outset controls how much extra space is cut
// out along the rims of each window.
func outsetArchSides(outset float64) model3d.Solid {
	var result model3d.JoinedSolid
	for i := 0; i < 7; i++ {
		x := -3.5 + ArchPanelWidth/2 + float64(i)*ArchPanelWidth
		fullHeight := i/3%2 == 0
		for _, y := range []float64{-1.5 + ArchPanelThickness/2, 1.5 - ArchPanelThickness/2} {
			result = append(
				result,
				ArchPanel(
					model3d.XYZ(x, y, ArchPanelHeight/2),
					fullHeight,
					outset,
				),
			)
		}
	}

	var uncurvedPanel model3d.JoinedSolid
	for i := 0; i < 3; i++ {
		// Shrink the two angled arches to prevent them
		// from coming close to the sides of the car.
		extraOutset := -0.03
		if i == 1 {
			extraOutset = 0
		}
		uncurvedPanel = append(
			uncurvedPanel,
			ArchPanel(
				model3d.XYZ(ArchPanelWidth*(0.5+float64(i)), 3.5, ArchPanelHeight/2),
				false,
				outset+extraOutset,
			),
		)
	}
	curveOffset := 0.2
	curvedPanel := model3d.CheckedFuncSolid(
		uncurvedPanel.Min(),
		uncurvedPanel.Max().Add(model3d.Y(curveOffset)),
		func(c model3d.Coord3D) bool {
			if c.X < ArchPanelWidth {
				c.Y -= curveOffset * c.X / ArchPanelWidth
			} else if c.X < ArchPanelWidth*2 {
				c.Y -= curveOffset
			} else {
				c.Y -= curveOffset * (ArchPanelWidth*3 - c.X) / ArchPanelWidth
			}
			return uncurvedPanel.Contains(c)
		},
	)

	sidePanel := model3d.RotateSolid(
		model3d.TranslateSolid(curvedPanel, model3d.X(-ArchPanelWidth*1.5)),
		model3d.Z(1),
		math.Pi/2,
	)
	otherSide := model3d.RotateSolid(sidePanel, model3d.Z(1), math.Pi)
	result = append(result, sidePanel, otherSide)
	return result
}

func ArchPanel(center model3d.Coord3D, fullHeight bool, outset float64) model3d.Solid {
	outer := model3d.NewRect(
		model3d.XYZ(-ArchPanelWidth/2, -ArchPanelThickness/2, -ArchPanelHeight/2),
		model3d.XYZ(ArchPanelWidth/2, ArchPanelThickness/2, ArchPanelHeight/2),
	)
	topInset := ArchTopInset - outset
	inset := ArchInset - outset
	archRadius := ArchPanelWidth/2 - inset
	arch := &model3d.Cylinder{
		P1:     model3d.XYZ(0, -ArchPanelThickness-1e-5, ArchPanelHeight/2-(topInset+archRadius)),
		P2:     model3d.XYZ(0, ArchPanelThickness+1e-5, ArchPanelHeight/2-(topInset+archRadius)),
		Radius: archRadius,
	}
	archBottom := model3d.NewRect(
		model3d.XYZ(-archRadius, -ArchPanelThickness-1e-5, -ArchPanelHeight/2+inset),
		model3d.XYZ(archRadius, ArchPanelThickness+1e-5, ArchPanelHeight/2-(topInset+archRadius)),
	)
	if !fullHeight {
		archBottom.MinVal.Z = 0
	}
	archBottom.MinVal.Z -= outset
	untranslated := &model3d.SubtractedSolid{
		Positive: outer,
		Negative: model3d.JoinedSolid{arch, archBottom},
	}
	return model3d.TranslateSolid(untranslated, center)
}

func ArchSidesText() model3d.Solid {
	minX := -3.5 + 3*ArchPanelWidth
	maxX := -3.5 + 6*ArchPanelWidth
	minZ := -ArchPanelHeight / 2
	maxZ := 0.0

	textMesh := model2d.MustReadBitmap("text.png", nil).Mesh().SmoothSq(20)
	size := textMesh.Max().Sub(textMesh.Min())
	textMesh = textMesh.Scale(math.Min((maxX-minX)/size.X, (maxZ-minZ)/size.Y))
	center2d := model2d.XY((minX+maxX)/2, (minZ+maxZ)/2)
	textMesh = textMesh.Center().Translate(center2d)
	var solid2d model2d.Solid = model2d.NewColliderSolid(model2d.MeshToCollider(textMesh))
	var res model3d.JoinedSolid
	for _, y := range []float64{-1.5 + ArchPanelThickness/2, 1.5 - ArchPanelThickness/2} {
		profile := model3d.ProfileSolid(solid2d, y-ArchPanelThickness/2, y+ArchPanelThickness/2)
		profile = model3d.RotateSolid(profile, model3d.X(1), -math.Pi/2)
		res = append(res, profile)
		solid2d = model2d.TranslateSolid(
			model2d.VecScaleSolid(
				model2d.TranslateSolid(solid2d, center2d.Scale(-1)),
				model2d.XY(-1, 1),
			),
			center2d,
		)
	}
	return res
}

func ArchedTop() (model3d.Solid, toolbox3d.CoordColorFunc) {
	virtualRadius := 8.0
	minZ := ArchPanelHeight + TopBottomThickness - 1e-5
	maxZ := minZ + ArchedTopHeight

	outerCurve := model3d.Cylinder{
		P1:     model3d.XYZ(-2.5, 0, maxZ-virtualRadius),
		P2:     model3d.XYZ(2.5, 0, maxZ-virtualRadius),
		Radius: virtualRadius,
	}
	innerCurve := outerCurve
	innerCurve.Radius -= TopBottomThickness

	bounds := model3d.NewRect(
		model3d.XYZ(-2.5, -1.0, minZ),
		model3d.XYZ(2.5, 1.0, maxZ),
	)
	x1 := 0.6
	x2 := 0.6 + TopBottomThickness
	return model3d.IntersectedSolid{
		model3d.JoinedSolid{
			model3d.IntersectedSolid{
				model3d.JoinedSolid{
					model3d.NewRect(model3d.XYZ(-2.5, -x2, minZ), model3d.XYZ(2.5, -x1, maxZ)),
					model3d.NewRect(model3d.XYZ(-2.5, x1, minZ), model3d.XYZ(2.5, x2, maxZ)),
				},
				&outerCurve,
			},
			&model3d.SubtractedSolid{Positive: &outerCurve, Negative: &innerCurve},
		},
		bounds,
	}, toolbox3d.ConstantCoordColorFunc(CarColor)
}
