package main

import (
	"log"
	"math"
	"os"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	. "github.com/unixpickle/model3d/shorthand"
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
		mesh := model3d.MarchingCubesConj(LegSolid(), 0.01, 8, ax)
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
		mesh := model3d.NewMeshRect(Origin3, XYZ(
			FootRadius*2,
			FootRadius*2,
			ScrewLength+ScrewRadius,
		))
		mesh.SaveGroupedSTL("infill_cube.stl")
	}
}

func StandSolid() Solid3 {
	topCenter := Z(StandRadius)
	var corners [3]C3
	for i := range corners {
		corners[i] = XYZ(
			StandRadius*math.Cos(float64(i)*math.Pi*2/3),
			StandRadius*math.Sin(float64(i)*math.Pi*2/3),
			0,
		)
	}
	return Join3(
		Sub3(
			Join3(
				Sphere(
					topCenter,
					FootRadius,
				),
				Cylinder(
					topCenter,
					corners[0],
					PoleThickness,
				),
				Cylinder(
					topCenter,
					corners[1],
					PoleThickness,
				),
				Cylinder(
					topCenter,
					corners[2],
					PoleThickness,
				),
				Sphere(
					corners[0],
					FootRadius,
				),
				Sphere(
					corners[1],
					FootRadius,
				),
				Sphere(
					corners[2],
					FootRadius,
				),
			),
			Join3(
				Rect3(
					XYZ(-StandRadius*2, -StandRadius*2, StandRadius),
					XYZ(StandRadius*2, StandRadius*2, StandRadius*2),
				),
				Rect3(
					XYZ(-StandRadius*2, -StandRadius*2, -StandRadius),
					XYZ(StandRadius*2, StandRadius*2, 0),
				),
			),
		),
		&toolbox3d.ScrewSolid{
			P1:         topCenter,
			P2:         topCenter.Add(Z(ScrewLength)),
			GrooveSize: ScrewGrooves,
			Radius:     ScrewRadius - ScrewSlack,
		},
	)
}

func ConeStandSolid() Solid3 {
	return Join3(
		ConeSolid(),
		&toolbox3d.ScrewSolid{
			P1:         Z(StandRadius),
			P2:         Z(StandRadius + ScrewLength),
			GrooveSize: ScrewGrooves,
			Radius:     ScrewRadius - ScrewSlack,
		},
	)
}

func ConeSolid() Solid3 {
	return MakeSolid3(
		XYZ(-StandRadius, -StandRadius, 0),
		XYZ(StandRadius, StandRadius, StandRadius),
		func(c C3) bool {
			radiusAtZ := StandRadius - c.Z
			radiusAtZInner := radiusAtZ - ConeThickness
			rad := c.XY().Norm()
			if rad <= radiusAtZ && rad >= radiusAtZInner {
				return true
			}
			// Top cylinder for mounting the screw.
			return rad < FootRadius && rad > radiusAtZ
		},
	)
}

func LegSolid() Solid3 {
	return Sub3(
		Join3(
			Cylinder(Origin3, Z(LegLength), PoleThickness),
			&toolbox3d.Ramp{
				Solid: &model3d.Cylinder{
					P2:     Z(FootRadius),
					Radius: FootRadius,
				},
				P1: Z(FootRadius),
			},
			&toolbox3d.Ramp{
				Solid: &model3d.Cylinder{
					P1:     Z(LegLength - FootRadius),
					P2:     Z(LegLength),
					Radius: FootRadius,
				},
				P1: Z(LegLength - FootRadius),
				P2: Z(LegLength),
			},
			&toolbox3d.ScrewSolid{
				P1:         Z(LegLength),
				P2:         Z(LegLength + ScrewLength),
				Radius:     ScrewRadius - ScrewSlack,
				GrooveSize: ScrewGrooves,
			},
		),
		&toolbox3d.ScrewSolid{
			P2:         Z(ScrewLength + ScrewRadius + ScrewSlack),
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooves,
			Pointed:    true,
		},
	)
}

func TopSolid() Solid3 {
	return Sub3(
		Cylinder(
			Origin3,
			Z(ScrewLength),
			TopRadius,
		),
		&toolbox3d.ScrewSolid{
			P2:         Z(ScrewLength),
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooves,
		},
	)
}
