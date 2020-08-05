package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	OtterWidth    = 3.0
	OtterHeight   = 0.15
	TextHeight    = 0.1
	TextSpace     = 0.2
	Bessel        = 0.3
	BaseThickness = 0.1

	SmoothingOutset = 0.02
)

func main() {
	otter := NewOtterSolid()
	headlines := ReadHeadlines(otter.Max().Y)
	solid := model3d.JoinedSolid{
		headlines[0],
		headlines[1],
		otter,
	}

	// Add base.
	min, max := solid.Min(), solid.Max()
	solid = append(solid, &model3d.Rect{
		MinVal: model3d.XYZ(min.X-Bessel, min.Y-Bessel, min.Z-BaseThickness),
		MaxVal: model3d.XYZ(max.X+Bessel, max.Y+Bessel, min.Z+1e-3),
	})

	log.Println("Creating rough mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	log.Println("Smoothing mesh...")
	mesh = mesh.FlipDelaunay()
	mesh = mesh.SmoothAreas(0.05, 10)
	log.Println("Simplifying mesh...")
	mesh = mesh.FlattenBase(0)
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("otter_love.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type OtterSolid struct {
	Shape model2d.SDF
}

func NewOtterSolid() *OtterSolid {
	img := model2d.MustReadBitmap("otter.png", nil).FlipY()
	scale := OtterWidth / float64(img.Width)
	mesh := img.Mesh().SmoothSq(20).Scale(scale)
	mesh = mesh.MapCoords(mesh.Min().Scale(-1).Add)
	return &OtterSolid{
		Shape: model2d.MeshToSDF(mesh),
	}
}

func (o *OtterSolid) Min() model3d.Coord3D {
	m := o.Shape.Min()
	return model3d.XY(m.X, m.Y)
}

func (o *OtterSolid) Max() model3d.Coord3D {
	max := o.Shape.Max()
	return model3d.XYZ(max.X, max.Y, OtterHeight)
}

func (o *OtterSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(o, c) {
		return false
	}
	sdf := o.Shape.SDF(c.XY())
	return sdf > c.Z
}

func ReadHeadlines(otterY float64) [2]model3d.Solid {
	m1 := ReadHeadline("line_1.png")
	m2 := ReadHeadline("line_2.png")
	m1 = m1.MapCoords(m1.Min().Scale(-1).Add)
	m2 = m2.MapCoords(m2.Min().Scale(-1).Add)
	scale := OtterWidth / math.Max(m1.Max().X, m2.Max().X)
	result := [2]model3d.Solid{}
	for i, m := range []*model2d.Mesh{m1, m2} {
		m = m.Scale(scale)
		m = m.MapCoords(model2d.X((OtterWidth - m.Max().X) / 2).Add)
		if i == 0 {
			m = m.MapCoords(model2d.Y(otterY + TextSpace).Add)
		} else {
			m = m.MapCoords(model2d.Y(-m.Max().Y - TextSpace).Add)
		}
		solid := model2d.NewColliderSolid(model2d.MeshToCollider(m))
		result[i] = model3d.ProfileSolid(solid, 0, TextHeight)
	}
	return result
}

func ReadHeadline(path string) *model2d.Mesh {
	bmp := model2d.MustReadBitmap(path, nil).FlipY()
	return bmp.Mesh().SmoothSq(50)
}
