package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	CubeSize = 0.5
	GapDepth = 0.02
	GapSize  = 0.03
)

func main() {
	rs := toolbox3d.NewRectSet()
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
				rs.Add(&model3d.Rect{
					MinVal: min,
					MaxVal: min.Add(model3d.XYZ(cubieSize, cubieSize, cubieSize)),
				})
			}
		}
	}
	// Join all the cubies together with a center cube
	// that takes up most of the volume.
	offset := (model3d.XYZ(1, 1, 1)).Scale(GapDepth)
	rs.Add(&model3d.Rect{
		MinVal: offset,
		MaxVal: rs.Max().Sub(offset),
	})
	mesh := rs.Mesh()
	ioutil.WriteFile("cube.stl", mesh.EncodeSTL(), 0755)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200, nil)
}
