package main

import (
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model3d"
)

const (
	StarThickness   = 1.0
	StarPointRadius = 2.0

	NumPoints = 6
)

func main() {
	baseMesh := CreateStarMesh()
	baseMesh.SaveGroupedSTL("star.stl")
	render3d.SaveRandomGrid("rendering.png", baseMesh, 3, 3, 300, nil)
}

func CreateStarMesh() *model3d.Mesh {
	midPoint := model3d.Z(StarThickness / 2)

	mesh := model3d.NewMesh()
	for i := 0; i < NumPoints*2; i += 2 {
		theta0 := float64(i) / float64(NumPoints*2) * math.Pi * 2
		theta1 := float64(i+1) / float64(NumPoints*2) * math.Pi * 2
		theta2 := float64(i+2) / float64(NumPoints*2) * math.Pi * 2

		p1 := model3d.XY(math.Cos(theta0), math.Sin(theta0))
		p2 := model3d.XY(math.Cos(theta1), math.Sin(theta1)).Scale(StarPointRadius)
		p3 := model3d.XY(math.Cos(theta2), math.Sin(theta2))

		mesh.Add(&model3d.Triangle{p2, p1, midPoint})
		mesh.Add(&model3d.Triangle{p2, p3, midPoint})
	}
	mesh.AddMesh(mesh.MapCoords(model3d.XYZ(1, 1, -1).Mul))

	// We created the mesh in a lazy way, so we must
	// fix holes and normals.
	mesh = mesh.Repair(1e-5)
	mesh, _ = mesh.RepairNormals(1e-5)
	return mesh
}
