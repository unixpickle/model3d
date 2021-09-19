package main

import (
	"flag"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var holeRadius float64
	var holeThickness float64
	var hangWidth float64
	var hangLength float64
	var hangThickness float64
	var hookLength float64
	var noDip bool

	flag.Float64Var(&holeRadius, "hole-radius", 0.5, "radius of ring for hanging this hook")
	flag.Float64Var(&holeThickness, "hole-thickness", 0.15, "thickness of ring for hanging this hook")
	flag.Float64Var(&hangWidth, "hang-width", 0.3, "width of the long, dangling part")
	flag.Float64Var(&hangLength, "hang-length", 4.0, "length of the long, dangling part")
	flag.Float64Var(&hangThickness, "hang-thickness", 0.15, "thickness of the long, dangling part")
	flag.Float64Var(&hookLength, "hook-length", 1.5, "length of the hook itself")
	flag.BoolVar(&noDip, "no-dip", false, "make low point of the hook flush with dangling part")
	flag.Parse()

	solid := model3d.JoinedSolid{
		model3d.ProfileSolid(
			&model2d.SubtractedSolid{
				Positive: model2d.JoinedSolid{
					// Length of hanging part
					model2d.NewRect(
						model2d.XY(-hangWidth/2, holeRadius),
						model2d.XY(hangWidth/2, hangLength),
					),
					// Ring out of which we cut the hole
					&model2d.Circle{Center: model2d.Y(holeRadius), Radius: holeRadius + holeThickness},
				},
				// The hole for hanging this hook
				Negative: &model2d.Circle{
					Center: model2d.Y(holeRadius),
					Radius: holeRadius,
				},
			},
			0,
			hangThickness,
		),
		HookSolid(noDip, hangLength, hookLength, hangWidth, hangThickness),
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("hanging_hook.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func HookSolid(noDip bool, minY, maxZ, width, thickness float64) model3d.Solid {
	eval := func(t float64) model2d.Coord {
		// Use part of a parabola that never has an overhang
		// of more than 45 degrees.
		if noDip {
			return model2d.XY(minY-maxZ/2+(0.25-math.Pow(t/2, 2))*maxZ*2, t*maxZ)
		} else {
			return model2d.XY(minY-thickness*0.5+(0.25-math.Pow(t-0.5, 2))*maxZ, t*maxZ)
		}
	}
	lines := model2d.NewMesh()
	eps := 0.01
	for t := 0.0; t+eps < 1.1; t += eps {
		lines.Add(&model2d.Segment{eval(t), eval(t + eps)})
	}
	collider := model2d.MeshToCollider(lines)
	return model3d.CheckedFuncSolid(
		model3d.XY(-width/2, collider.Min().X-thickness),
		model3d.XYZ(width/2, collider.Max().X+thickness, maxZ),
		func(c model3d.Coord3D) bool {
			return collider.CircleCollision(c.YZ(), thickness/2)
		},
	)
}
