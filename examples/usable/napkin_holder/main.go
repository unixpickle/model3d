package main

import (
	"log"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/model2d"
)

const (
	SideSize         = 5
	Depth            = 2
	Thickness        = 0.2
	InscriptionDepth = 0.1
)

func main() {
	log.Println("Creating inscription...")
	inscription := NewDeepInscription()
	log.Println("Creating main mesh...")
	solid := model3d.JoinedSolid{
		inscription,
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: SideSize, Y: SideSize, Z: Thickness},
		},
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{Z: Depth - Thickness},
			MaxVal: model3d.Coord3D{X: SideSize, Y: SideSize, Z: Depth},
		},
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: SideSize, Y: Thickness, Z: Depth},
		},
	}
	mesh := model3d.SolidToMesh(solid, 0.01, 0, -1, 5)
	log.Println("Eliminating co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		return model3d.Coord3D{X: -c.X, Y: c.Z, Z: c.Y}
	})
	mesh.SaveGroupedSTL("napkin_holder.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

type DeepInscription struct {
	Collider model2d.Collider
	Solid    model2d.Solid
	Scale    float64
}

func NewDeepInscription() *DeepInscription {
	bmp := model2d.MustReadBitmap("image.png", nil).FlipY()
	solid := model2d.BitmapToSolid(bmp)
	collider := model2d.MeshToCollider(bmp.Mesh())
	return &DeepInscription{
		Collider: collider,
		Solid:    solid,
		Scale:    float64(bmp.Width) / SideSize,
	}
}

func (d *DeepInscription) Min() model3d.Coord3D {
	return model3d.Coord3D{Z: Depth - 1e-3}
}

func (d *DeepInscription) Max() model3d.Coord3D {
	return model3d.Coord3D{X: SideSize, Y: SideSize, Z: Depth + InscriptionDepth}
}

func (d *DeepInscription) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(d, c) {
		return false
	}
	c2d := c.Coord2D().Scale(d.Scale)
	if !d.Solid.Contains(c2d) {
		return false
	}
	margin := d.Scale * (c.Z - Depth)
	if margin < 0 {
		return false
	}
	return model2d.ColliderContains(d.Collider, c2d, margin)
}
