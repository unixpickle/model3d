package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d"
)

const (
	PoleSpacing   = 0.75
	PoleRadius    = 0.15
	PoleLength    = 4.0
	BaseSize      = 1.0
	BaseThickness = 0.2
	NumPoles      = 6
)

func main() {
	length := NumPoles * PoleSpacing
	solid := model3d.JoinedSolid{
		&model3d.RectSolid{
			MaxVal: model3d.Coord3D{X: BaseSize, Y: length, Z: BaseThickness},
		},
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{X: BaseSize, Y: length/2 - PoleSpacing/2, Z: 0},
			MaxVal: model3d.Coord3D{X: BaseSize * 2, Y: length/2 + PoleSpacing/2, Z: BaseThickness},
		},
	}
	for i := 0; i < NumPoles; i++ {
		y := float64(i)*PoleSpacing + PoleSpacing/2
		x := BaseSize / 2
		solid = append(solid, &model3d.CylinderSolid{
			P1:     model3d.Coord3D{X: x, Y: y, Z: BaseThickness},
			P2:     model3d.Coord3D{X: x, Y: y, Z: BaseThickness + PoleLength},
			Radius: PoleRadius,
		})
	}
	mesh := model3d.SolidToMesh(solid, PoleRadius/3, 2, 0.8, 5)
	ioutil.WriteFile("shoe_rack.stl", mesh.EncodeSTL(), 0755)

	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}