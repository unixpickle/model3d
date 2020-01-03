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
	inscription := NewInscription()
	inscriptionMesh := model3d.SolidToMesh(inscription, 0.01, 0, -1, 5)
	deepInscription := &DeepInscription{
		Collider: model3d.MeshToCollider(inscriptionMesh),
		Solid:    inscription,
	}

	log.Println("Creating main mesh...")
	solid := model3d.JoinedSolid{
		deepInscription,
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

type Inscription struct {
	Solid model2d.Solid
}

func NewInscription() *Inscription {
	bmp := model2d.MustReadBitmap("image.png", nil).FlipY()
	solid := model2d.BitmapToSolid(bmp)
	solid = model2d.ScaleSolid(solid, SideSize/float64(bmp.Width))
	return &Inscription{Solid: solid}
}

func (i *Inscription) Min() model3d.Coord3D {
	// Subtract a small epsilon to prevent the entire
	// solid from being removed by DeepInscription.
	return model3d.Coord3D{Z: Depth - 1e-3}
}

func (i *Inscription) Max() model3d.Coord3D {
	return model3d.Coord3D{X: SideSize, Y: SideSize, Z: Depth + InscriptionDepth*2}
}

func (i *Inscription) Contains(c model3d.Coord3D) bool {
	return model3d.InSolidBounds(i, c) && i.Solid.Contains(c.Coord2D())
}

type DeepInscription struct {
	Collider model3d.Collider
	Solid    *Inscription
}

func (d *DeepInscription) Min() model3d.Coord3D {
	return d.Solid.Min()
}

func (d *DeepInscription) Max() model3d.Coord3D {
	return d.Solid.Max()
}

func (d *DeepInscription) Contains(c model3d.Coord3D) bool {
	if !d.Solid.Contains(c) {
		return false
	}
	radius := c.Z - Depth
	return !d.Collider.SphereCollision(c, radius)
}
