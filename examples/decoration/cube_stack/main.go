package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Rounding         = 0.1
	HollowThickness  = 0.2
	HollowHoleRadius = 0.2
)

func main() {
	cubes := [3]model3d.Solid{
		Cube(model3d.XZ(-0.1, 1.0), 1.0, 0, false),
		Cube(model3d.XZ(0.0, 2.8-1e-5), 0.8, 0.3, true),
		Cube(model3d.XZ(-0.05, 4.3-2e-5), 0.7, -0.2, true),
	}
	solid := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid(cubes[:]),
		Negative: &model3d.Cylinder{
			P1:     model3d.XZ(-0.05, -1e-5),
			P2:     model3d.XZ(-0.05, cubes[2].Max().Z-HollowThickness),
			Radius: HollowHoleRadius,
		},
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	log.Println("Simplifying mesh...")
	preCount := len(mesh.TriangleSlice())
	mesh = mesh.EliminateCoplanar(1e-5)
	postCount := len(mesh.TriangleSlice())
	log.Printf("Reduced triangle count from %d to %d", preCount, postCount)

	log.Println("Rendering...")
	colorFn := CubeColorFunc(cubes)
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(2.0, -10.0, 7.0), 1024, 1024,
		colorFn.RenderColor)
	log.Println("Saving...")
	mesh.SaveMaterialOBJ("cubes.zip", colorFn.TriangleColor)
}

func Cube(center model3d.Coord3D, size, rotation float64, hollow bool) model3d.Solid {
	baseSolid := model3d.NewRect(model3d.Ones(-(size - Rounding)), model3d.Ones(size-Rounding))
	rounded := model3d.NewColliderSolidInset(baseSolid, -Rounding)
	var solid model3d.Solid = rounded
	if hollow {
		solid = &model3d.SubtractedSolid{
			Positive: rounded,
			Negative: model3d.NewColliderSolidInset(baseSolid, HollowThickness-Rounding),
		}
	}
	return model3d.TranslateSolid(
		model3d.RotateSolid(solid, model3d.Z(1), rotation),
		center,
	)
}

func CubeColorFunc(cubes [3]model3d.Solid) toolbox3d.CoordColorFunc {
	return toolbox3d.JoinedCoordColorFunc(
		model3d.MarchingCubesSearch(cubes[0], 0.02, 16),
		render3d.NewColorRGB(0.2, 0.3, 1),
		model3d.MarchingCubesSearch(cubes[1], 0.02, 16),
		render3d.NewColorRGB(0, 1, 0),
		model3d.MarchingCubesSearch(cubes[2], 0.02, 16),
		render3d.NewColorRGB(1, 0, 0),
	).Cached()
}
