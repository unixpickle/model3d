package main

import (
	"log"

	"github.com/unixpickle/model3d"
)

const (
	ShelfWidth      = 6.0
	ShelfHeight     = 2.0
	ShelfDepth      = 6.0
	ShelfSlack      = 0.05
	ShelfThickness  = 0.4
	ShelfHoleRadius = 0.1

	ShelfCount = 3

	ContainerThickness      = 0.2
	ContainerFootWidth      = 0.6
	ContainerFootHeight     = 0.2
	ContainerFootRampHeight = ContainerFootWidth / 2

	RidgeDepth = 0.2
)

func main() {
	container := CreateContainer()
	shelf := CreateShelf()

	log.Println("Creating container mesh...")
	mesh := model3d.SolidToMesh(container, 0.02, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving container mesh...")
	mesh.SaveGroupedSTL("container.stl")
	log.Println("Rendering container mesh...")
	model3d.SaveRandomGrid("container.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	log.Println("Creating shelf mesh...")
	mesh = model3d.SolidToMesh(shelf, 0.02, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving shelf mesh...")
	mesh.SaveGroupedSTL("shelf.stl")
	log.Println("Rendering shelf mesh...")
	model3d.SaveRandomGrid("shelf.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}
