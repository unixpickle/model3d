package main

import (
	"image"
	"image/png"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	Radius        = 1.0
	LengthRadians = 1.0
	Thickness     = 0.1
	BaseRadius    = 0.2
)

func main() {
	shape := NewFlowerShape()
	sphereMesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		if shape.Get(g.Lat, g.Lon) {
			return Radius
		} else {
			return 0
		}
	}, 300)
	sphereMesh.Iterate(func(t *model3d.Triangle) {
		if t[0].Norm() == 0 || t[1].Norm() == 0 || t[2].Norm() == 0 {
			sphereMesh.Remove(t)
		}
	})
	solid := model3d.JoinedSolid{
		model3d.NewColliderSolidHollow(model3d.MeshToCollider(sphereMesh), Thickness),
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: Radius - Thickness},
			P2:     model3d.Coord3D{Z: Radius + Thickness},
			Radius: BaseRadius,
		},
	}

	mesh := model3d.SolidToMesh(solid, 0.01, 0, -1, 10)
	mesh.SaveGroupedSTL("flower.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
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
