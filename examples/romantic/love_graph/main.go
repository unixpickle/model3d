package main

import (
	"fmt"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
)

const (
	TextHeight  = 1.0
	TextPadding = 0.5

	HeartWidth   = 1.0
	HeartSpacing = 0.5

	BaseThickness = 0.2
	TextThickness = 0.1
)

func main() {
	labels, centers := LoadLabels()
	heart := LoadHeart()

	fullSolid := model3d.JoinedSolid{
		&model3d.Rect{
			MinVal: model3d.Y(-TextPadding),
			MaxVal: model3d.XYZ(
				labels.Max().X,
				labels.Max().Y+heart.Max().Y+HeartSpacing*2,
				BaseThickness,
			),
		},
		model3d.ProfileSolid(labels, BaseThickness-1e-5, BaseThickness+TextThickness),
	}

	for i, center := range centers {
		heartSolid := model3d.TransformSolid(
			&model3d.Translate{
				Offset: model3d.XY(
					center-HeartWidth,
					labels.Max().Y+HeartSpacing,
				),
			},
			model3d.ProfileSolid(heart, BaseThickness-1e-5, BaseThickness+float64(i+1)),
		)
		fullSolid = append(fullSolid, heartSolid)
	}

	mesh := model3d.MarchingCubesSearch(fullSolid, 0.02, 8)
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

	return &fixedBounds{
		Solid:  model2d.NewColliderSolid(model2d.MeshToCollider(fullMesh)),
		MaxVal: model2d.XY(curX, maxHeight),
	}, centers
}

func LoadHeart() model2d.Solid {
	bmp := model2d.MustReadBitmap("heart.png", nil)
	mesh := bmp.FlipY().Mesh().SmoothSq(20)
	min, max := mesh.Min(), mesh.Max()
	mesh = mesh.Scale(HeartWidth / (max.X - min.X))
	mesh = mesh.MapCoords(mesh.Min().Scale(-1).Add)
	return model2d.NewColliderSolid(model2d.MeshToCollider(mesh))
}

type fixedBounds struct {
	model2d.Solid
	MinVal model2d.Coord
	MaxVal model2d.Coord
}

func (f *fixedBounds) Min() model2d.Coord {
	return f.MinVal
}

func (f *fixedBounds) Max() model2d.Coord {
	return f.MaxVal
}
