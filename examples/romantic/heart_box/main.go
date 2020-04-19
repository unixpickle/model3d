package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	SideSize         = 7.0
	BottomThickness  = 0.2
	WallHeight       = 2.0
	WallThickness    = 0.3
	SectionHeight    = 1.7
	SectionThickness = 0.2

	LidHeight      = 1.5
	LidThickness   = 0.2
	LidHolderSize  = 0.2
	LidHolderInset = 0.06

	ImageSize = 1024.0
)

func main() {
	log.Println("Loading 2D shapes...")
	outline := model2d.MeshToCollider(
		model2d.MustReadBitmap("outline.png", nil).FlipY().Mesh().SmoothSq(200),
	)
	sections := model2d.MeshToCollider(
		model2d.MustReadBitmap("sections.png", nil).FlipY().Mesh().SmoothSq(100),
	)

	log.Println("Creating box...")
	mesh := model3d.MarchingCubesSearch(&BoxSolid{Outline: outline, Sections: sections}, 0.02, 8)
	log.Println(" - eliminating co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println(" - flattening base...")
	mesh = mesh.FlattenBase(0)
	log.Println(" - saving...")
	mesh.SaveGroupedSTL("box.stl")
	log.Println(" - rendering...")
	render3d.SaveRandomGrid("rendering_box.png", mesh, 3, 3, 300, nil)

	log.Println("Creating lid...")
	mesh = model3d.MarchingCubesSearch(&LidSolid{Outline: outline}, 0.02, 8)
	log.Println(" - eliminating co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println(" - flattening base...")
	mesh = mesh.FlattenBase(0)
	log.Println(" - saving...")
	mesh.SaveGroupedSTL("lid.stl")
	log.Println(" - rendering...")
	render3d.SaveRandomGrid("rendering_lid.png", mesh, 3, 3, 300, nil)
}

type BoxSolid struct {
	Outline  model2d.Collider
	Sections model2d.Collider
}

func (b *BoxSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (b *BoxSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: SideSize, Y: SideSize, Z: WallHeight}
}

func (b *BoxSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	scale := ImageSize / SideSize
	c2 := c.Coord2D().Scale(scale)
	if !model2d.ColliderContains(b.Outline, c2, 0) {
		return false
	}
	if c.Z < BottomThickness {
		return true
	}
	if b.Outline.CircleCollision(c2, WallThickness*scale) {
		return true
	}
	return c.Z < SectionHeight && b.Sections.CircleCollision(c2, 0.5*SectionThickness*scale)
}

type LidSolid struct {
	Outline model2d.Collider
}

func (l *LidSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (l *LidSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: SideSize, Y: SideSize, Z: LidHeight + LidHolderSize}
}

func (l *LidSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(l, c) {
		return false
	}
	scale := ImageSize / SideSize
	c2 := c.Coord2D().Scale(scale)
	if !model2d.ColliderContains(l.Outline, c2, 0) {
		return false
	}
	if c.Z < LidThickness {
		return true
	}
	if !model2d.ColliderContains(l.Outline, c2, scale*(WallThickness+LidHolderInset)) {
		return c.Z < LidHeight
	}
	return !model2d.ColliderContains(l.Outline, c2,
		scale*(WallThickness+LidHolderInset+LidThickness))
}
