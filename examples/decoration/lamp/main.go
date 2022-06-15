package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	light := NewLampLight()
	holder := model3d.JoinedSolid{
		model3d.NewRect(model3d.XYZ(-0.7, -0.2, -2.0), model3d.XYZ(0.7, 0.2, -0.1)),
		model3d.NewRect(model3d.XYZ(-0.7, -0.5, -2.4), model3d.XYZ(0.7, 0.5, -1.99)),
	}
	holderMesh := model3d.MarchingCubesSearch(holder, 0.01, 8)
	holderMeshColor := light.Cast(holderMesh)
	tree := model3d.NewCoordTree(holderMesh.VertexSlice())

	lightSDF := model3d.MeshToSDF(light.Mesh)
	colorFn := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		if lightSDF.SDF(c) > -0.05 {
			return light.Color
		} else {
			neighbor := tree.NearestNeighbor(c)
			return holderMeshColor[neighbor].Scale(0.7).Add(render3d.NewColor(0.3))
		}
	})

	joinedMesh := model3d.NewMesh()
	joinedMesh.AddMesh(holderMesh)
	joinedMesh.AddMesh(light.Mesh)
	render3d.SaveRandomGrid("rendering.png", joinedMesh, 3, 3, 300, colorFn.RenderColor)
}
