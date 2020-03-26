package main

import (
	"log"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/model2d"
)

const (
	Scale     = 0.003
	Width     = 682 * Scale
	Height    = 573 * Scale
	Depth     = 0.06
	TextDepth = 0.1
)

func main() {
	solid := &Solid{
		Body: model2d.MeshToCollider(
			model2d.MustReadBitmap("body.png", nil).FlipY().Mesh().Blur(0.25).Blur(0.25),
		),
		Text: model2d.MeshToCollider(
			model2d.MustReadBitmap("text.png", nil).FlipY().Mesh().Blur(0.25).Blur(0.25),
		),
	}
	log.Println("Creating mesh...")
	mesh := model3d.SolidToMesh(solid, 0.005, 0, 0, 0)
	log.Println("Smoothing...")
	mesh = mesh.SmoothAreas(0.05, 50)
	log.Println("Flattening...")
	mesh = mesh.FlattenBase(0)
	log.Println("Eliminating co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving...")
	mesh.SaveGroupedSTL("tag.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

type Solid struct {
	Body model2d.Collider
	Text model2d.Collider
}

func (s *Solid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (s *Solid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Width, Y: Height, Z: TextDepth}
}

func (s *Solid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}
	c2d := c.Coord2D().Scale(1 / Scale)
	if c.Z < Depth && model2d.ColliderContains(s.Body, c2d, 0) {
		return true
	}
	return c.Z < TextDepth && model2d.ColliderContains(s.Text, c2d, 0)
}
