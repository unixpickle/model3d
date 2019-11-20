package main

import (
	_ "image/gif"
	"log"
	"math"
	"os"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d"
)

const (
	HolderSize      = 0.5
	HolderThickness = 0.2
	TrackSize       = HolderSize / math.Sqrt2
	PieceSize       = HolderSize + TrackSize - 0.05
	PieceBottomSize = TrackSize + 0.1
	PieceThickness  = 0.2

	ScrewRadius     = 0.1
	ScrewSlack      = 0.015
	ScrewGrooveSize = 0.04

	BottomThickness = 0.3
	TotalThickness  = 1.0
	WallThickness   = 0.3
)

func main() {
	if _, err := os.Stat("board.stl"); os.IsNotExist(err) {
		log.Println("Creating board...")
		mesh := model3d.SolidToMesh(BoardSolid(), 0.01, 0, -1, 5)
		log.Println("Eliminating co-planar polygons...")
		mesh = mesh.EliminateCoplanar(1e-8)
		mesh.SaveGroupedSTL("board.stl")
	}

	if _, err := os.Stat("holder.stl"); os.IsNotExist(err) {
		log.Println("Creating holder...")
		mesh := model3d.SolidToMesh(HolderSolid(), 0.01, 1, -1, 5)
		mesh.SaveGroupedSTL("holder.stl")
	}

	if _, err := os.Stat("piece_top.stl"); os.IsNotExist(err) {
		log.Println("Creating piece top...")
		mesh := model3d.SolidToMesh(PieceTopSolid(), 0.01, 1, -1, 5)
		mesh.SaveGroupedSTL("piece_top.stl")
	}

	if _, err := os.Stat("piece_bottom.stl"); os.IsNotExist(err) {
		log.Println("Creating piece bottom...")
		mesh := model3d.SolidToMesh(PieceBottomSolid(), 0.01, 1, -1, 5)
		mesh.SaveGroupedSTL("piece_bottom.stl")
	}
}

func BoardSolid() model3d.Solid {
	bottomSize := WallThickness*2 + HolderSize*4 + TrackSize*4
	solid := model3d.JoinedSolid{
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: bottomSize, Y: bottomSize, Z: BottomThickness},
		},

		// Front and back sides
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: bottomSize, Y: WallThickness, Z: TotalThickness},
		},
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{Y: bottomSize - WallThickness},
			MaxVal: model3d.Coord3D{X: bottomSize, Y: bottomSize, Z: TotalThickness},
		},

		// Left and right sides
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{Y: bottomSize, X: WallThickness, Z: TotalThickness},
		},
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{X: bottomSize - WallThickness},
			MaxVal: model3d.Coord3D{Y: bottomSize, X: bottomSize, Z: TotalThickness},
		},
	}

	// Edge holders which can be built in to the board.
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			if x == 0 || x == 4 || y == 0 || y == 4 {
				solid = append(solid, &model3d.RectSolid{
					MinVal: model3d.Coord3D{
						X: WallThickness - HolderSize/2 + float64(x)*(HolderSize+TrackSize),
						Y: WallThickness - HolderSize/2 + float64(y)*(HolderSize+TrackSize),
						Z: TotalThickness - HolderThickness - PieceThickness,
					},
					MaxVal: model3d.Coord3D{
						X: WallThickness + HolderSize/2 + float64(x)*(HolderSize+TrackSize),
						Y: WallThickness + HolderSize/2 + float64(y)*(HolderSize+TrackSize),
						Z: TotalThickness - PieceThickness,
					},
				})
			}
		}
	}

	// Screw holes for holders.
	screws := model3d.JoinedSolid{}
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			cx := WallThickness + float64(x+1)*(HolderSize+TrackSize)
			cy := WallThickness + float64(y+1)*(HolderSize+TrackSize)
			screws = append(screws, &toolbox3d.ScrewSolid{
				P1:         model3d.Coord3D{X: cx, Y: cy, Z: TotalThickness - PieceThickness},
				P2:         model3d.Coord3D{X: cx, Y: cy, Z: 0},
				Radius:     ScrewRadius,
				GrooveSize: ScrewGrooveSize,
			})
		}
	}

	return &model3d.SubtractedSolid{
		Positive: solid,
		Negative: screws,
	}
}

func HolderSolid() model3d.Solid {
	center := HolderSize / 2
	return model3d.JoinedSolid{
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: HolderSize, Y: HolderSize, Z: HolderThickness},
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{X: center, Y: center, Z: 0},
			P2:         model3d.Coord3D{X: center, Y: center, Z: TotalThickness - PieceThickness},
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGrooveSize,
		},
	}
}

func PieceTopSolid() model3d.Solid {
	center := PieceSize / 2
	return model3d.JoinedSolid{
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: PieceSize, Y: PieceSize, Z: PieceThickness},
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{X: center, Y: center, Z: 0},
			P2:         model3d.Coord3D{X: center, Y: center, Z: TotalThickness - BottomThickness},
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGrooveSize,
		},
	}
}

func PieceBottomSolid() model3d.Solid {
	center := PieceBottomSize / 2
	return &model3d.SubtractedSolid{
		Positive: &model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: PieceBottomSize, Y: PieceBottomSize, Z: PieceThickness},
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{X: center, Y: center, Z: TotalThickness - BottomThickness},
			P2:         model3d.Coord3D{X: center, Y: center, Z: 0},
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooveSize,
		},
	}
}
