package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d"
)

const (
	CubeSize = 0.5
	GapDepth = 0.02
	GapSize  = 0.03
)

func main() {
	solid := model3d.JoinedSolid{}
	cubieSize := (CubeSize - GapSize*2) / 3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				if i == 1 && j == 1 && k == 1 {
					// Center cubie is not visible.
					continue
				}
				min := model3d.Coord3D{
					X: float64(i) * (cubieSize + GapSize),
					Y: float64(j) * (cubieSize + GapSize),
					Z: float64(k) * (cubieSize + GapSize),
				}
				solid = append(solid, &model3d.RectSolid{
					MinVal: min,
					MaxVal: min.Add(model3d.Coord3D{X: cubieSize, Y: cubieSize, Z: cubieSize}),
				})
			}
		}
	}
	// Join all the cubies together with a center cube
	// that takes up most of the volume.
	offset := (model3d.Coord3D{X: 1, Y: 1, Z: 1}).Scale(GapDepth)
	solid = append(solid, &model3d.RectSolid{
		MinVal: offset,
		MaxVal: solid.Max().Sub(offset),
	})
	mesh := model3d.SolidToMesh(solid, 0.01, 2, -1, 5)
	ioutil.WriteFile("cube.stl", mesh.EncodeSTL(), 0755)
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}