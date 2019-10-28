package main

import "github.com/unixpickle/model3d"

func JewelryPiece(idx int, p1, p2 model3d.Coord3D) model3d.Solid {
	return &model3d.SphereSolid{
		Center: p1.Mid(p2),
		Radius: p1.Dist(p2) / 2,
	}
}
