package main

import (
	"log"
	"math"
	"os"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	HolderSize           = 0.5
	TrackSize            = HolderSize / math.Sqrt2
	PieceSize            = HolderSize + TrackSize - 0.04
	PieceThickness       = 0.2
	FullPieceBottomSize  = TrackSize + 0.2
	SmallPieceBottomSize = TrackSize + 0.2
	PoleRadius           = 0.14

	BottomThickness = 0.2
	TotalThickness  = 0.6
	WallThickness   = 0.25

	SupportSlope = 1.3
)

func main() {
	if _, err := os.Stat("board.stl"); os.IsNotExist(err) {
		log.Println("Creating board...")
		mesh := model3d.MarchingCubesSearch(BoardSolid(), 0.01, 8)
		log.Println("Eliminating co-planar polygons...")
		mesh = mesh.EliminateCoplanar(1e-8)
		log.Println("Saving board...")
		mesh.SaveGroupedSTL("board.stl")
		log.Println("Rendering board...")
		render3d.SaveRandomGrid("rendering_board.png", mesh, 3, 3, 500, nil)
	}

	if _, err := os.Stat("piece.stl"); os.IsNotExist(err) {
		log.Println("Creating piece...")
		mesh := model3d.MarchingCubesSearch(PieceSolid(FullPieceBottomSize), 0.005, 8)
		log.Println("Eliminating co-planar polygons...")
		mesh = mesh.EliminateCoplanar(1e-8)
		log.Println("Saving piece...")
		mesh.SaveGroupedSTL("piece.stl")
		log.Println("Rendering piece...")
		render3d.SaveRandomGrid("rendering_piece.png", mesh, 3, 3, 300, nil)
	}

	if _, err := os.Stat("small_piece.stl"); os.IsNotExist(err) {
		log.Println("Creating small piece...")
		mesh := model3d.MarchingCubesSearch(PieceSolid(SmallPieceBottomSize), 0.005, 8)
		log.Println("Eliminating co-planar polygons...")
		mesh = mesh.EliminateCoplanar(1e-8)
		log.Println("Saving small piece...")
		mesh.SaveGroupedSTL("small_piece.stl")
	}
}

func BoardSolid() model3d.Solid {
	bottomSize := WallThickness*2 + HolderSize*4 + TrackSize*4
	solid := model3d.JoinedSolid{
		&model3d.Rect{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.XYZ(bottomSize, bottomSize, BottomThickness),
		},

		// Front and back sides
		&model3d.Rect{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.XYZ(bottomSize, WallThickness, TotalThickness),
		},
		&model3d.Rect{
			MinVal: model3d.Coord3D{Y: bottomSize - WallThickness},
			MaxVal: model3d.XYZ(bottomSize, bottomSize, TotalThickness),
		},

		// Left and right sides
		&model3d.Rect{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{Y: bottomSize, X: WallThickness, Z: TotalThickness},
		},
		&model3d.Rect{
			MinVal: model3d.Coord3D{X: bottomSize - WallThickness},
			MaxVal: model3d.Coord3D{Y: bottomSize, X: bottomSize, Z: TotalThickness},
		},
	}

	// Create all edge holders.
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			rect := &model3d.Rect{
				MinVal: model3d.Coord3D{
					X: WallThickness - HolderSize/2 + float64(x)*(HolderSize+TrackSize),
					Y: WallThickness - HolderSize/2 + float64(y)*(HolderSize+TrackSize),
					Z: TotalThickness - SupportSlope*HolderSize/2 - PieceThickness,
				},
				MaxVal: model3d.Coord3D{
					X: WallThickness + HolderSize/2 + float64(x)*(HolderSize+TrackSize),
					Y: WallThickness + HolderSize/2 + float64(y)*(HolderSize+TrackSize),
					Z: TotalThickness - PieceThickness,
				},
			}
			mid := rect.MinVal.Mid(rect.MaxVal)
			p1 := mid
			p1.Z = rect.MinVal.Z
			p2 := mid
			p2.Z = rect.MaxVal.Z
			solid = append(solid, &toolbox3d.Ramp{
				Solid: rect,
				P1:    p1,
				P2:    p2,
			})

			if !(x == 0 || x == 4 || y == 0 || y == 4) {
				// Non-side edge holders need something to
				// hold them up.
				solid = append(solid, &model3d.Cylinder{
					P1:     model3d.Coord3D{X: p2.X, Y: p2.Y},
					P2:     p2,
					Radius: PoleRadius,
				})
			}
		}
	}

	return solid
}

func PieceSolid(bottomSize float64) model3d.Solid {
	center := PieceSize / 2
	return model3d.JoinedSolid{
		&model3d.Rect{
			MinVal: model3d.Z(BottomThickness),
			MaxVal: model3d.Coord3D{
				X: PieceSize,
				Y: PieceSize,
				Z: BottomThickness + PieceThickness,
			},
		},
		&model3d.Cylinder{
			P1: model3d.XYZ(center, center, BottomThickness),
			P2: model3d.Coord3D{
				X: center,
				Y: center,
				Z: TotalThickness,
			},
			Radius: PoleRadius,
		},
		&toolbox3d.Ramp{
			Solid: &model3d.Rect{
				MinVal: model3d.Coord3D{
					X: center - bottomSize/2,
					Y: center - bottomSize/2,
					Z: TotalThickness - SupportSlope*bottomSize/2,
				},
				MaxVal: model3d.Coord3D{
					X: center + bottomSize/2,
					Y: center + bottomSize/2,
					Z: TotalThickness,
				},
			},
			P1: model3d.Coord3D{
				X: center,
				Y: center,
				Z: TotalThickness - SupportSlope*bottomSize/2,
			},
			P2: model3d.Coord3D{
				X: center,
				Y: center,
				Z: TotalThickness,
			},
		},
	}
}
