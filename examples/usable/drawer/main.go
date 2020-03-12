package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	DrawerWidth      = 6.0
	DrawerHeight     = 2.0
	DrawerDepth      = 6.0
	DrawerSlack      = 0.05
	DrawerThickness  = 0.4
	DrawerBottom     = 0.2
	DrawerHoleRadius = 0.1

	DrawerCount = 3

	FrameThickness      = 0.2
	FrameFootWidth      = 0.6
	FrameFootHeight     = 0.2
	FrameFootRampHeight = FrameFootWidth / 2
	FrameHoleWidth      = 1.0
	FrameHoleMargin     = 0.4

	BottomHoleRadius = 5.0

	RidgeDepth = 0.2

	KnobBaseRadius = 0.4
	KnobBaseLength = 0.2
	KnobPoleRadius = 0.2
	KnobPoleLength = 0.6

	KnobScrewRadius  = 0.08
	KnobScrewGroove  = 0.03
	KnobScrewSlack   = 0.02
	KnobNutRadius    = 0.2
	KnobNutThickness = 0.2
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

	// Don't render the middle part in as high
	// resolution, since it's uniform along the
	// y axis.
	squeeze := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisY,
		Min:   1.0,
		Max:   DrawerDepth - 1.0,
		Ratio: 0.1,
	}

	knobSqueeze := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   KnobBaseLength + KnobPoleLength*0.1,
		Max:   KnobBaseLength + KnobPoleLength*0.9,
		Ratio: 0.1,
	}

	CreateMesh(CreateDrawer(), "drawer", 0.015, squeeze)
	CreateMesh(CreateFrame(), "frame", 0.02, nil)
	CreateMesh(CreateKnob(), "knob", 0.0025, knobSqueeze)
	CreateMesh(CreateKnobNut(), "nut", 0.0025, nil)
}

func CreateMesh(solid model3d.Solid, name string, resolution float64, ax *toolbox3d.AxisSqueeze) {
	if _, err := os.Stat(filepath.Join(ModelDir, name+".stl")); err == nil {
		log.Printf("Skipping %s mesh", name)
		return
	}

	log.Printf("Creating %s mesh...", name)
	var mesh *model3d.Mesh
	if ax != nil {
		mesh = ax.SolidToMesh(solid, resolution, 0, -1, 5)
	} else {
		mesh = model3d.SolidToMesh(solid, resolution, 0, -1, 5)
	}
	log.Println("Eliminating co-planar polygons...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Printf("Saving %s mesh...", name)
	mesh.SaveGroupedSTL(filepath.Join(ModelDir, name+".stl"))
	log.Printf("Rendering %s mesh...", name)
	model3d.SaveRandomGrid(filepath.Join(RenderDir, name+".png"), model3d.MeshToCollider(mesh),
		3, 3, 300, 300)
}
