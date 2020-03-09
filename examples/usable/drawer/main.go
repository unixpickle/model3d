package main

import (
	"log"

	"github.com/unixpickle/model3d"
)

const (
	DrawerWidth      = 6.0
	DrawerHeight     = 2.0
	DrawerDepth      = 6.0
	DrawerSlack      = 0.05
	DrawerThickness  = 0.4
	DrawerHoleRadius = 0.1

	DrawerCount = 3

	FrameThickness      = 0.2
	FrameFootWidth      = 0.6
	FrameFootHeight     = 0.2
	FrameFootRampHeight = FrameFootWidth / 2

	RidgeDepth = 0.2
)

func main() {
	frame := CreateFrame()
	drawer := CreateDrawer()

	log.Println("Creating frame mesh...")
	mesh := model3d.SolidToMesh(frame, 0.02, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving frame mesh...")
	mesh.SaveGroupedSTL("frame.stl")
	log.Println("Rendering frame mesh...")
	model3d.SaveRandomGrid("frame.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	log.Println("Creating drawer mesh...")
	mesh = model3d.SolidToMesh(drawer, 0.02, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving drawer mesh...")
	mesh.SaveGroupedSTL("drawer.stl")
	log.Println("Rendering drawer mesh...")
	model3d.SaveRandomGrid("drawer.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}
