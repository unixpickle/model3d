package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
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

const (
	ModelDir  = "models"
	RenderDir = "renderings"
)

func main() {
	if _, err := os.Stat(ModelDir); os.IsNotExist(err) {
		essentials.Must(os.Mkdir(ModelDir, 0755))
	}
	if _, err := os.Stat(RenderDir); os.IsNotExist(err) {
		essentials.Must(os.Mkdir(RenderDir, 0755))
	}

	CreateMesh(CreateDrawer(), "drawer", 0.015)
	CreateMesh(CreateFrame(), "frame", 0.02)
}

func CreateMesh(solid model3d.Solid, name string, resolution float64) {
	if _, err := os.Stat(filepath.Join(ModelDir, name+".stl")); err == nil {
		log.Printf("Skipping %s mesh", name)
		return
	}

	log.Printf("Creating %s mesh...", name)
	mesh := model3d.SolidToMesh(solid, resolution, 0, -1, 5)
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Printf("Saving %s mesh...", name)
	mesh.SaveGroupedSTL(filepath.Join(ModelDir, name+".stl"))
	log.Printf("Rendering %s mesh...", name)
	model3d.SaveRandomGrid(filepath.Join(RenderDir, name+".png"), model3d.MeshToCollider(mesh),
		3, 3, 300, 300)
}
