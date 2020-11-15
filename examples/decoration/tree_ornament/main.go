package main

import (
	"log"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
)

func main() {
	log.Println("Creating star mesh...")
	solid := CreateStarSolid()
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	log.Println("Simplifying star mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving star mesh...")
	mesh.SaveGroupedSTL("star.stl")
	log.Println("Rendering star mesh...")
	render3d.SaveRandomGrid("rendering_star.png", mesh, 3, 3, 300, nil)

	log.Println("Creating hanger mesh...")
	solid = CreateHangerSolid()
	solid1 := CreateHangerSpacerSolid()
	mesh = model3d.MarchingCubesSearch(solid, 0.02, 8)
	mesh1 := model3d.MarchingCubesSearch(solid1, 0.02, 8)
	log.Println("Simplifying hanger mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh1 = mesh1.EliminateCoplanar(1e-5)
	log.Println("Saving hanger mesh...")
	mesh.SaveGroupedSTL("hanger.stl")
	mesh1.SaveGroupedSTL("hanger_spacer.stl")
	log.Println("Rendering hanger mesh...")
	render3d.SaveRandomGrid("rendering_hanger.png", mesh, 3, 3, 300, nil)
}
