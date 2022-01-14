package main

import "github.com/unixpickle/model3d/model3d"

const Rounding = 0.05

func main() {
	cubes := []model3d.Solid{
		Cube(model3d.Z(1.0), 0),
		Cube(model3d.Z(3.0-1e-5), 0.3),
		Cube(model3d.Z(5.0-2e-5), 0.1),
	}
	solid := model3d.JoinedSolid(cubes)
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	mesh.SaveGroupedSTL("mesh.stl")
}

func Cube(center model3d.Coord3D, rotation float64) model3d.Solid {
	baseSolid := model3d.NewRect(model3d.Ones(-(1 - Rounding)), model3d.Ones(1-Rounding))
	rounded := model3d.NewColliderSolidInset(baseSolid, -Rounding)
	return model3d.TranslateSolid(
		model3d.RotateSolid(rounded, model3d.Z(1), rotation),
		center,
	)
}
