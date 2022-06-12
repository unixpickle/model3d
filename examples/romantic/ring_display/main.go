package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const Resolution = 0.01

func main() {
	log.Println("Creating box...")
	box, boxColor := RingBox()

	log.Println("Creating metal ring...")
	metalRing, metalRingColor := MetalRing()
	metalRingXf := &model3d.Translate{Offset: model3d.X(0.33)}
	metalRing = model3d.TransformSolid(metalRingXf, metalRing)
	metalRingColor = metalRingColor.Transform(metalRingXf)

	log.Println("Creating diamond ring...")
	diamondRing, diamondRingColor := DiamondRing()
	diamondRingXf := model3d.JoinedTransform{
		model3d.Rotation(model3d.Y(1), -0.35),
		&model3d.Translate{Offset: model3d.XZ(-0.33, 0.1)},
	}
	diamondRing = model3d.TransformSolid(diamondRingXf, diamondRing)
	diamondRingColor = diamondRingColor.Transform(diamondRingXf)

	log.Println("Creating joint mesh...")
	joinedSolid := model3d.JoinedSolid{box, metalRing, diamondRing}
	mesh := model3d.MarchingCubesSearch(joinedSolid, Resolution, 8)

	log.Println("Creating joint color func...")
	colorFn := toolbox3d.JoinedCoordColorFunc(
		model3d.MarchingCubesSearch(box, Resolution, 8),
		boxColor,
		model3d.MarchingCubesSearch(metalRing, Resolution, 8),
		metalRingColor,
		model3d.MarchingCubesSearch(diamondRing, Resolution, 8),
		diamondRingColor,
	)

	log.Println("Saving...")
	mesh.SaveMaterialOBJ("ring_display.zip", colorFn.TriangleColor)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFn.RenderColor)
}
