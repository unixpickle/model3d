package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Scale     = 0.003
	Depth     = 0.06
	TextDepth = 0.1
)

func main() {
	body := Load2d("body.png")
	text := Load2d("text.png")
	solid := model3d.JoinedSolid{
		model3d.ProfileSolid(body, 0, Depth),
		model3d.ProfileSolid(text, 0, TextDepth),
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	log.Println("Smoothing...")
	mesh = mesh.SmoothAreas(0.05, 50)
	log.Println("Flattening...")
	mesh = mesh.FlattenBase(0)
	log.Println("Eliminating co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving...")
	mesh.SaveGroupedSTL("tag.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func Load2d(name string) model2d.Solid {
	collider := model2d.MeshToCollider(
		model2d.MustReadBitmap(name, nil).FlipY().Mesh().Blur(0.25).Blur(0.25),
	)
	return model2d.ScaleSolid(model2d.NewColliderSolid(collider), Scale)
}
