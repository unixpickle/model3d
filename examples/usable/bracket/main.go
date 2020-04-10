package main

import (
	"io/ioutil"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
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
			&model3d.Rect{
				MaxVal: model3d.Coord3D{X: BracketSize, Z: BracketHeight, Y: BracketDepth},
			},
			&model3d.Rect{
				MinVal: model3d.Coord3D{X: BracketSize - (HolePadding+HoleRadius)*2},
				MaxVal: model3d.Coord3D{
					X: BracketSize,
					Z: BracketHeight,
					Y: HoleDepth,
				},
			},
			&model3d.Rect{
				MaxVal: model3d.Coord3D{X: BracketDepth, Z: BracketHeight, Y: BracketSize},
			},
			&model3d.Rect{
				MinVal: model3d.Coord3D{Y: BracketSize - (HolePadding+HoleRadius)*2},
				MaxVal: model3d.Coord3D{
					Y: BracketSize,
					Z: BracketHeight,
					X: HoleDepth,
				},
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1: model3d.Coord3D{
					X: BracketSize - (HolePadding + HoleRadius),
					Z: BracketHeight / 2,
					Y: 0,
				},
				P2: model3d.Coord3D{
					X: BracketSize - (HolePadding + HoleRadius),
					Z: BracketHeight / 2,
					Y: HoleDepth,
				},
				Radius: HoleRadius,
			},
			&model3d.Cylinder{
				P1: model3d.Coord3D{
					X: 0,
					Z: BracketHeight / 2,
					Y: BracketSize - (HolePadding + HoleRadius),
				},
				P2: model3d.Coord3D{
					X: HoleDepth,
					Z: BracketHeight / 2,
					Y: BracketSize - (HolePadding + HoleRadius),
				},
				Radius: HoleRadius,
			},
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.5, 8)
	ioutil.WriteFile("bracket.stl", mesh.EncodeSTL(), 0755)

	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200, nil)
}
