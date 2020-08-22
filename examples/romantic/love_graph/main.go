package main

import (
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
)

const (
	TextHeight  = 0.5
	TextPadding = 0.25

	HeartWidth   = 0.5
	HeartSpacing = 0.25

	BaseThickness = 0.2
	TextThickness = 0.1
)

func main() {
	labels, centers := LoadLabels()
	heart := LoadHeart()

	fullSolid := model3d.JoinedSolid{
		// Base
		&model3d.Rect{
			MinVal: model3d.Y(-TextPadding),
			MaxVal: model3d.XYZ(
				labels.Max().X,
				labels.Max().Y+heart.Max().Y+HeartSpacing*2,
				BaseThickness,
			),
		},
		// Text profile
		model3d.ProfileSolid(labels, BaseThickness-1e-5, BaseThickness+TextThickness),
	}

	// Bar graph bars (in shapes of hearts)
	heights := []float64{0.5, 0.4, 1.5}
	for i, center := range centers {
		height := heights[i]
		heartSolid := model3d.TransformSolid(
			&model3d.Translate{
				Offset: model3d.XY(
					center-HeartWidth/2,
					labels.Max().Y+HeartSpacing,
				),
			},
			model3d.ProfileSolid(heart, BaseThickness-1e-5, BaseThickness+height),
		)
		fullSolid = append(fullSolid, heartSolid)
	}

	log.Println("Creating mesh...")
	xform := GraphAxisSqueeze(heights)
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(xform, fullSolid), 0.01, 16)
	mesh = mesh.MapCoords(xform.Inverse().Apply)

	log.Println("Flattening Z surfaces...")
	// Without this, the top of the text and the hearts may be jagged.
	pinchZ := append([]float64{TextThickness}, heights...)
	for _, z := range pinchZ {
		mesh = mesh.MapCoords((&toolbox3d.AxisPinch{
			Axis:  toolbox3d.AxisZ,
			Min:   BaseThickness + z - 0.02,
			Max:   BaseThickness + z + 0.02,
			Power: 4,
		}).Apply)
	}

	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("love_graph.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func LoadLabels() (solidOut model2d.Solid, centers []float64) {
	var meshes []*model2d.Mesh
	var maxHeight float64
	for i := 1; i <= 3; i++ {
		bmp := model2d.MustReadBitmap(fmt.Sprintf("label%d.png", i), nil)
		mesh := bmp.FlipY().Mesh().SmoothSq(20)
		min, max := mesh.Min(), mesh.Max()
		maxHeight = math.Max(maxHeight, max.Y-min.Y)
		meshes = append(meshes, mesh)
	}

	scale := TextHeight / maxHeight
	maxHeight *= scale

	fullMesh := model2d.NewMesh()
	curX := 0.0
	for _, m := range meshes {
		m = m.Scale(scale)
		min, max := m.Min(), m.Max()
		width := max.X - min.X
		if width < HeartWidth {
			width = HeartWidth
		}
		m = m.MapCoords(min.Scale(-1).Add)
		m = m.MapCoords(model2d.XY(
			curX+TextPadding+(width-max.X+min.X)/2,
			(maxHeight-max.Y+min.Y)/2,
		).Add)
		centers = append(centers, curX+width/2+TextPadding)
		curX += width + TextPadding*2
		fullMesh.AddMesh(m)
	}

	return model2d.ForceSolidBounds(
		model2d.NewColliderSolid(model2d.MeshToCollider(fullMesh)),
		model2d.Coord{},
		model2d.XY(curX, maxHeight),
	), centers
}

func LoadHeart() model2d.Solid {
	bmp := model2d.MustReadBitmap("heart.png", nil)
	mesh := bmp.FlipY().Mesh().SmoothSq(20)
	min, max := mesh.Min(), mesh.Max()
	mesh = mesh.Scale(HeartWidth / (max.X - min.X))
	mesh = mesh.MapCoords(mesh.Min().Scale(-1).Add)
	return model2d.NewColliderSolid(model2d.MeshToCollider(mesh))
}

func GraphAxisSqueeze(heights []float64) *toolbox3d.AxisSqueeze {
	sorted := append([]float64{}, heights...)
	sort.Float64s(sorted)
	minZ := sorted[len(sorted)-2] + BaseThickness + 0.02
	maxZ := sorted[len(sorted)-1] - 0.02
	return &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   minZ,
		Max:   maxZ,
		Ratio: 0.1,
	}
}
