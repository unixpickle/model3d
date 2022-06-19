package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const Production = false

func main() {
	log.Println("Creating lamp...")
	light := NewLampLight()

	log.Println("Creating scene...")
	solid, colorFunc := CreateScene()

	log.Println("Creating final solid...")
	hollowLight := &model3d.SubtractedSolid{
		Positive: light.Solid,
		Negative: model3d.JoinedSolid{
			model3d.NewColliderSolidInset(light.Object, 0.15),
			&model3d.Cylinder{P1: model3d.Z(1.0), P2: model3d.Z(2.21), Radius: 0.2},
		},
	}
	solid = model3d.JoinedSolid{
		hollowLight,
		solid,
	}

	log.Println("Creating final mesh...")
	delta := 0.02
	if Production {
		delta = 0.01
	}
	fullMesh := model3d.MarchingCubesSearch(solid, delta, 8)

	log.Println("Recoloring scene...")
	colorFunc = light.Recolor(fullMesh, colorFunc)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", fullMesh, 3, 3, 300, colorFunc.RenderColor)

	log.Println("Saving...")
	fullMesh.SaveQuantizedMaterialOBJ("lamp.zip", 32, colorFunc.Cached().TriangleColor)
}

func CreateScene() (model3d.Solid, toolbox3d.CoordColorFunc) {
	lampBase := model3d.JoinedSolid{
		model3d.NewRect(model3d.XYZ(-0.7, -0.2, 0.0), model3d.XYZ(0.7, 0.2, 1.7)),
		model3d.NewRect(model3d.XYZ(-0.7, -0.5, 0.0), model3d.XYZ(0.7, 0.5, 0.4)),
	}
	plant, plantColor := CreatePlant()
	base := model3d.NewRect(model3d.XYZ(-3.0, -1.5, -0.2), model3d.XYZ(1.0, 1.5, 0.001))
	return model3d.JoinedSolid{lampBase, plant, base}, toolbox3d.JoinedCoordColorFunc(
		model3d.MeshToSDF(model3d.MarchingCubesSearch(lampBase, 0.02, 8)),
		render3d.NewColor(1.0),
		model3d.MeshToSDF(model3d.MarchingCubesSearch(plant, 0.02, 8)),
		plantColor,
		model3d.MeshToSDF(model3d.NewMeshRect(base.MinVal, base.MaxVal)),
		render3d.NewColor(1.0),
	)
}
