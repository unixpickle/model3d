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
	ScrewRadius  = 0.25
	ScrewGrooves = 0.05
	ScrewSlack   = 0.03
	ScrewLength  = 0.5

	StandRadius = 3.5
	TopRadius   = 3.5

	PoleThickness = 0.4
	FootRadius    = 0.8
	LegLength     = 6.0
	ConeThickness = 0.2
)

func main() {
	if _, err := os.Stat("stand.stl"); os.IsNotExist(err) {
		log.Println("Creating stand...")
		mesh := model3d.MarchingCubesSearch(StandSolid(), 0.01, 8)
		mesh.SaveGroupedSTL("stand.stl")
		render3d.SaveRandomGrid("stand.png", mesh, 3, 3, 300, nil)
	}

	if _, err := os.Stat("cone_stand.stl"); os.IsNotExist(err) {
		log.Println("Creating stand (cone)...")
		mesh := model3d.MarchingCubesSearch(ConeStandSolid(), 0.01, 8)
		mesh.SaveGroupedSTL("cone_stand.stl")
		render3d.SaveRandomGrid("cone_stand.png", mesh, 3, 3, 300, nil)
	}

	if _, err := os.Stat("leg.stl"); os.IsNotExist(err) {
		log.Println("Creating leg...")
		ax := &toolbox3d.AxisSqueeze{
			Axis:  toolbox3d.AxisZ,
			Min:   1,
			Max:   LegLength - 1,
			Ratio: 0.1,
		}
		mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, LegSolid()), 0.01, 8)
		mesh = mesh.MapCoords(ax.Inverse().Apply)
		mesh.SaveGroupedSTL("leg.stl")
		render3d.SaveRandomGrid("leg.png", mesh, 3, 3, 300, nil)
	}

	if _, err := os.Stat("top.stl"); os.IsNotExist(err) {
		log.Println("Creating top...")
		mesh := model3d.MarchingCubesSearch(TopSolid(), 0.01, 8)
		log.Println("Eliminating co-planar...")
		mesh = mesh.EliminateCoplanar(1e-8)
		mesh.SaveGroupedSTL("top.stl")
		render3d.SaveRandomGrid("top.png", mesh, 3, 3, 300, nil)
	}

	if _, err := os.Stat("infill_cube.stl"); os.IsNotExist(err) {
		log.Println("Creating infill cube for filling in top screws...")
		mesh := model3d.NewMeshRect(model3d.Coord3D{}, model3d.Coord3D{
			X: FootRadius * 2,
			Y: FootRadius * 2,
			Z: ScrewLength + ScrewRadius,
		})
		mesh.SaveGroupedSTL("infill_cube.stl")
	}
}

func StandSolid() model3d.Solid {
	topCenter := model3d.Z(StandRadius)
	var corners [3]model3d.Coord3D
	for i := range corners {
		corners[i] = model3d.Coord3D{
			X: StandRadius * math.Cos(float64(i)*math.Pi*2/3),
			Y: StandRadius * math.Sin(float64(i)*math.Pi*2/3),
		}
	}
	return model3d.JoinedSolid{
		&model3d.SubtractedSolid{
			Positive: model3d.JoinedSolid{
				&model3d.Sphere{
					Center: topCenter,
					Radius: FootRadius,
				},
				&model3d.Cylinder{
					P1:     topCenter,
					P2:     corners[0],
					Radius: PoleThickness,
				},
				&model3d.Cylinder{
					P1:     topCenter,
					P2:     corners[1],
					Radius: PoleThickness,
				},
				&model3d.Cylinder{
					P1:     topCenter,
					P2:     corners[2],
					Radius: PoleThickness,
				},
				&model3d.Sphere{
					Center: corners[0],
					Radius: FootRadius,
				},
				&model3d.Sphere{
					Center: corners[1],
					Radius: FootRadius,
				},
				&model3d.Sphere{
					Center: corners[2],
					Radius: FootRadius,
				},
			},
			Negative: model3d.JoinedSolid{
				&model3d.Rect{
					MinVal: model3d.Coord3D{X: -StandRadius * 2, Y: -StandRadius * 2, Z: StandRadius},
					MaxVal: model3d.Coord3D{X: StandRadius * 2, Y: StandRadius * 2, Z: StandRadius * 2},
				},
				&model3d.Rect{
					MinVal: model3d.Coord3D{X: -StandRadius * 2, Y: -StandRadius * 2, Z: -StandRadius},
					MaxVal: model3d.Coord3D{X: StandRadius * 2, Y: StandRadius * 2, Z: 0},
				},
			},
		},
		&toolbox3d.ScrewSolid{
			P1:         topCenter,
			P2:         topCenter.Add(model3d.Z(ScrewLength)),
			GrooveSize: ScrewGrooves,
			Radius:     ScrewRadius - ScrewSlack,
		},
	}
}

func ConeStandSolid() model3d.Solid {
	return model3d.JoinedSolid{
		ConeSolid{},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Z(StandRadius),
			P2:         model3d.Coord3D{Z: StandRadius + ScrewLength},
			GrooveSize: ScrewGrooves,
			Radius:     ScrewRadius - ScrewSlack,
		},
	}
}

type ConeSolid struct{}

func (c ConeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -StandRadius, Y: -StandRadius}
}

func (c ConeSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: StandRadius, Y: StandRadius, Z: StandRadius}
}

func (c ConeSolid) Contains(coord model3d.Coord3D) bool {
	if !model3d.InBounds(c, coord) {
		return false
	}
	radiusAtZ := StandRadius - coord.Z
	radiusAtZInner := radiusAtZ - ConeThickness
	rad := coord.Coord2D().Norm()
	if rad <= radiusAtZ && rad >= radiusAtZInner {
		return true
	}
	// Top cylinder for mounting the screw.
	return rad < FootRadius && rad > radiusAtZ
}

func LegSolid() model3d.Solid {
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.Cylinder{
				P2:     model3d.Z(LegLength),
				Radius: PoleThickness,
			},
			&toolbox3d.Ramp{
				Solid: &model3d.Cylinder{
					P2:     model3d.Z(FootRadius),
					Radius: FootRadius,
				},
				P1: model3d.Z(FootRadius),
			},
			&toolbox3d.Ramp{
				Solid: &model3d.Cylinder{
					P1:     model3d.Coord3D{Z: LegLength - FootRadius},
					P2:     model3d.Z(LegLength),
					Radius: FootRadius,
				},
				P1: model3d.Coord3D{Z: LegLength - FootRadius},
				P2: model3d.Z(LegLength),
			},
			&toolbox3d.ScrewSolid{
				P1:         model3d.Z(LegLength),
				P2:         model3d.Coord3D{Z: LegLength + ScrewLength},
				Radius:     ScrewRadius - ScrewSlack,
				GrooveSize: ScrewGrooves,
			},
		},
		Negative: &toolbox3d.ScrewSolid{
			P2:         model3d.Coord3D{Z: ScrewLength + ScrewRadius + ScrewSlack},
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooves,
			Pointed:    true,
		},
	}
}

func TopSolid() model3d.Solid {
	return &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P2:     model3d.Z(ScrewLength),
			Radius: TopRadius,
		},
		Negative: &toolbox3d.ScrewSolid{
			P2:         model3d.Z(ScrewLength),
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooves,
		},
	}
}
