package main

import (
	"image"
	"image/png"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Radius        = 1.0
	LengthRadians = 1.0
	Thickness     = 0.1
	BaseThickness = 0.1
	BaseRadius    = 0.5
)

func main() {
	shape := NewFlowerShape()
	sphereMesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		if shape.Get(g.Lat, g.Lon) {
			return Radius
		} else {
			return 0
		}
	}, 500)
	sphereMesh.Iterate(func(t *model3d.Triangle) {
		if t[0].Norm() == 0 || t[1].Norm() == 0 || t[2].Norm() == 0 {
			sphereMesh.Remove(t)
		}
	})
	sphereMesh = sphereMesh.MapCoords(func(f model3d.Coord3D) model3d.Coord3D {
		f.Z *= -1
		f.Z += Radius
		return f
	})
	solid := model3d.JoinedSolid{
		model3d.NewColliderSolidHollow(model3d.MeshToCollider(sphereMesh), Thickness),
		BaseSolid{},
	}

	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8).SmoothAreas(0.1, 20)
	mesh.SaveGroupedSTL("flower.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type FlowerShape struct {
	Img image.Image
}

func NewFlowerShape() *FlowerShape {
	f, err := os.Open("shape.png")
	essentials.Must(err)
	defer f.Close()
	img, err := png.Decode(f)
	essentials.Must(err)
	return &FlowerShape{Img: img}
}

func (f *FlowerShape) Get(x, y float64) bool {
	intX := int(float64(f.Img.Bounds().Dx()) * (x + LengthRadians) / (LengthRadians * 2))
	intY := int(float64(f.Img.Bounds().Dy()) * (y + LengthRadians) / (LengthRadians * 2))
	_, _, _, a := f.Img.At(intX, intY).RGBA()
	return a > 0xffff/2
}

type BaseSolid struct{}

func (b BaseSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -Radius, Y: -Radius, Z: -BaseThickness}
}

func (b BaseSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Radius, Y: Radius, Z: Radius}
}

func (b BaseSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(b.Min()) != b.Min() || c.Max(b.Max()) != b.Max() {
		return false
	}
	cylinderDist := (model3d.Coord3D{X: c.X, Y: c.Y}).Norm()
	sphereDist := (model3d.Coord3D{X: c.X, Y: c.Y, Z: Radius - c.Z}).Norm()
	return cylinderDist < BaseRadius && sphereDist >= Radius
}
