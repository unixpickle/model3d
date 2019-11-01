package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d"
)

const (
	BracketSize   = 40
	BracketHeight = 15
	BracketDepth  = 5

	HolePadding = 5
	HoleDepth   = 10
	HoleRadius  = 2
)

func main() {
	solid := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.RectSolid{
				MaxVal: model3d.Coord3D{X: BracketSize, Y: BracketHeight, Z: BracketDepth},
			},
			&model3d.RectSolid{
				MinVal: model3d.Coord3D{X: BracketSize - (HolePadding+HoleRadius)*2},
				MaxVal: model3d.Coord3D{
					X: BracketSize,
					Y: BracketHeight,
					Z: HoleDepth,
				},
			},
			&model3d.RectSolid{
				MaxVal: model3d.Coord3D{X: BracketDepth, Y: BracketHeight, Z: BracketSize},
			},
			&model3d.RectSolid{
				MinVal: model3d.Coord3D{Z: BracketSize - (HolePadding+HoleRadius)*2},
				MaxVal: model3d.Coord3D{
					Z: BracketSize,
					Y: BracketHeight,
					X: HoleDepth,
				},
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.CylinderSolid{
				P1: model3d.Coord3D{
					X: BracketSize - (HolePadding + HoleRadius),
					Y: BracketHeight / 2,
					Z: 0,
				},
				P2: model3d.Coord3D{
					X: BracketSize - (HolePadding + HoleRadius),
					Y: BracketHeight / 2,
					Z: HoleDepth,
				},
				Radius: HoleRadius,
			},
			&model3d.CylinderSolid{
				P1: model3d.Coord3D{
					X: 0,
					Y: BracketHeight / 2,
					Z: BracketSize - (HolePadding + HoleRadius),
				},
				P2: model3d.Coord3D{
					X: HoleDepth,
					Y: BracketHeight / 2,
					Z: BracketSize - (HolePadding + HoleRadius),
				},
				Radius: HoleRadius,
			},
		},
	}
	mesh := model3d.SolidToMesh(solid, 1, 1, 0.8, 5)
	ioutil.WriteFile("bracket.stl", mesh.EncodeSTL(), 0755)

	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}
