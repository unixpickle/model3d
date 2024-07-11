package main

import (
	"flag"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	SideLength     float64 `default:"1.6"`
	Height         float64 `default:"4.0"`
	TopThickness   float64 `default:"0.1"`
	BottomHeight   float64 `default:"0.5"`
	TopCoverHeight float64 `default:"0.2"`

	BaseSmoothing float64 `default:"0.025"`
	BaseStairSize float64 `default:"0.1"`

	TextInset float64 `default:"0.025"`

	DividerWidth     float64 `default:"0.1"`
	DividerThickness float64 `default:"0.1"`
	SlotThickness    float64 `default:"0.1"`
	SlotExtraEdge    float64 `default:"0.05"`
	CornerMargin     float64 `default:"0.3"`
	VerticalWindows  int     `default:"4"`

	CutoutThickness float64 `default:"0.1"`
	CutoutPlay      float64 `default:"0.01"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	mainSolid, headingHeight := BodySolid(&args)
	_ = headingHeight

	textCutout := TelephoneTextSolid(&args, headingHeight)
	solid := &model3d.SubtractedSolid{
		Positive: mainSolid,
		Negative: textCutout,
	}
	outsetCutout := model3d.NewColliderSolidInset(
		model3d.MeshToCollider(model3d.DualContour(textCutout, 0.005, true, false)),
		-0.001,
	)
	colorFunc := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if outsetCutout.Contains(c) {
			return render3d.NewColor(1)
		} else {
			return render3d.NewColorRGB(1, 0, 0)
		}
	})

	topRect := model3d.NewRect(
		model3d.XYZ(-10000, -10000, args.Height-headingHeight),
		model3d.XYZ(10000, 10000, args.Height),
	)

	createMesh := func(top bool) *model3d.Mesh {
		var cutSolid model3d.Solid
		if top {
			cutSolid = model3d.IntersectedSolid{topRect, solid}
		} else {
			cutSolid = &model3d.SubtractedSolid{Positive: solid, Negative: topRect}
		}
		if top {
			// Inside part to prevent the top from sliding
			inset := args.SlotThickness + args.SlotThickness + args.CutoutThickness + args.CutoutPlay
			side := args.SideLength/2 - inset
			holder := model3d.NewRect(
				model3d.XYZ(-side, -side, topRect.Min().Z-args.CutoutThickness),
				model3d.XYZ(side, side, topRect.Min().Z+1e-5),
			)
			cutSolid = model3d.JoinedSolid{cutSolid, holder}
		} else {
			inset := args.SlotThickness + args.SlotThickness + args.CutoutThickness
			side := args.SideLength/2 - inset
			cutout := model3d.NewRect(
				model3d.XYZ(-side, -side, 0),
				model3d.XYZ(side, side, args.Height),
			)
			cutSolid = &model3d.SubtractedSolid{Positive: cutSolid, Negative: cutout}
		}
		mesh := model3d.DualContour(cutSolid, 0.01, true, false)
		oldCount := mesh.NumTriangles()
		mesh = mesh.EliminateCoplanarFiltered(1e-5, colorFunc.ChangeFilterFunc(mesh, 0.1))
		newCount := mesh.NumTriangles()
		log.Printf(" - went from %d to %d triangles", oldCount, newCount)
		return mesh
	}

	log.Println("Creating top...")
	topMesh := createMesh(true)
	log.Println("Creating bottom...")
	bottomMesh := createMesh(false)

	log.Println("Saving...")
	topMesh.SaveMaterialOBJ("phone_booth_top.zip", colorFunc.TriangleColor)
	bottomMesh.SaveMaterialOBJ("phone_booth_bottom.zip", colorFunc.TriangleColor)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_top.png", topMesh, 3, 3, 300, colorFunc.RenderColor)
	render3d.SaveRandomGrid("rendering_bottom.png", bottomMesh, 3, 3, 300, colorFunc.RenderColor)
}

func BodySolid(a *Args) (model3d.Solid, float64) {
	intersectTopSphere := &model3d.Sphere{
		Center: model3d.Z(a.Height * 0.2),
		Radius: a.Height * 0.8,
	}

	bodyRect := model3d.NewRect(
		model3d.XYZ(-a.SideLength/2, -a.SideLength/2, 0),
		model3d.XYZ(a.SideLength/2, a.SideLength/2, a.Height),
	)

	lowerTopSphere := *intersectTopSphere
	lowerTopSphere.Center.Z -= a.TopThickness
	top := &model3d.SubtractedSolid{
		Positive: model3d.IntersectedSolid{
			bodyRect.Expand(a.TopThickness),
			intersectTopSphere,
		},
		Negative: model3d.TranslateSolid(intersectTopSphere, model3d.Z(-a.TopThickness)),
	}

	headingHeight := a.Height * (0.7 / 4.0)

	var cuts model3d.JoinedSolid

	cutAreaHeight := a.Height - headingHeight - a.TopCoverHeight - a.BottomHeight
	cutHeight := cutAreaHeight / float64(a.VerticalWindows)
	for i := 0; i < a.VerticalWindows; i++ {
		cutWidth := (a.SideLength - a.CornerMargin*2 - a.DividerWidth) / 2
		cutYMin := a.SideLength/2 - a.DividerThickness - a.SlotThickness
		cutZMax := a.Height - headingHeight - a.TopCoverHeight - cutHeight*float64(i)
		cutZMin := a.Height - headingHeight - a.TopCoverHeight - cutHeight*float64(i+1) + a.DividerWidth
		cuts = append(cuts, model3d.NewRect(
			model3d.XYZ(-a.SideLength/2+a.CornerMargin, cutYMin, cutZMin),
			model3d.XYZ(-a.SideLength/2+a.CornerMargin+cutWidth, a.SideLength/2+1e-5, cutZMax),
		))
		cuts = append(cuts, model3d.NewRect(
			model3d.XYZ(a.SideLength/2-a.CornerMargin-cutWidth, cutYMin, cutZMin),
			model3d.XYZ(a.SideLength/2-a.CornerMargin, a.SideLength/2+1e-5, cutZMax),
		))
	}

	// Cut out for slot
	slot := model3d.NewRect(
		model3d.XYZ(
			-a.SideLength/2+a.CornerMargin-a.SlotExtraEdge,
			a.SideLength/2-a.DividerThickness-a.SlotThickness,
			a.BottomHeight-a.SlotExtraEdge,
		),
		model3d.XYZ(
			a.SideLength/2-a.CornerMargin+a.SlotExtraEdge,
			a.SideLength/2-a.DividerThickness,
			a.Height-headingHeight+a.SlotExtraEdge,
		),
	)
	cuts = append(cuts, slot)
	log.Printf("Slot size: %f x %f", slot.Max().X-slot.Min().X, slot.Max().Z-slot.Min().Z)

	// Cut out all four sides, with radial symmetry.
	allCuts := model3d.JoinedSolid{}
	for i := 0; i < 4; i++ {
		theta := math.Pi / 2 * float64(i)
		allCuts = append(allCuts, model3d.RotateSolid(cuts, model3d.Z(1), theta))
	}

	fullSolid := model3d.JoinedSolid{
		model3d.IntersectedSolid{
			&model3d.SubtractedSolid{Positive: bodyRect, Negative: allCuts},
			intersectTopSphere,
		},
		top,
		BottomSolid(a),
	}

	return fullSolid, headingHeight
}

func BottomSolid(a *Args) model3d.Solid {
	bottomThickness := a.BaseStairSize
	var colliders []model3d.Collider
	for i := 0; i < 3; i++ {
		outset := float64(i+1)*bottomThickness - a.BaseSmoothing
		top := -float64(i) * bottomThickness
		bottom := -float64(i+1) * bottomThickness
		colliders = append(colliders, model3d.NewRect(
			model3d.XYZ(-a.SideLength/2-outset, -a.SideLength/2-outset, bottom),
			model3d.XYZ(a.SideLength/2+outset, a.SideLength/2+outset, top),
		))
	}
	return model3d.NewColliderSolidInset(model3d.NewJoinedCollider(colliders), -a.BaseSmoothing)
}

func TelephoneTextSolid(a *Args, headingHeight float64) model3d.Solid {
	img := model2d.MustReadBitmap("images/telephone.png", nil).FlipY().Mesh().SmoothSq(20)
	scale := 0.7 * a.SideLength / (img.Max().X - img.Min().X)
	img = img.Scale(scale)
	img = img.Translate(img.Min().Mid(img.Max()).Scale(-1))
	img = img.Translate(model2d.Y(a.Height - headingHeight*0.7))
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(img))
	solid := model3d.RotateSolid(
		model3d.ProfileSolid(solid2d, a.SideLength/2-a.TextInset, a.SideLength/2+1e-5),
		model3d.X(1),
		math.Pi/2,
	)
	var allSides model3d.JoinedSolid
	for i := 0; i < 4; i++ {
		allSides = append(allSides, model3d.RotateSolid(
			solid,
			model3d.Z(1),
			math.Pi/2*float64(i),
		))
	}
	return allSides
}
