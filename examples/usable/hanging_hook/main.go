package main

import (
	"flag"
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

	flag.Float64Var(&holeRadius, "hole-radius", 0.5, "radius of ring for hanging this hook")
	flag.Float64Var(&holeThickness, "hole-thickness", 0.15, "thickness of ring for hanging this hook")
	flag.Float64Var(&hangWidth, "hang-width", 0.3, "width of the long, dangling part")
	flag.Float64Var(&hangLength, "hang-length", 4.0, "length of the long, dangling part")
	flag.Float64Var(&hangThickness, "hang-thickness", 0.15, "thickness of the long, dangling part")
	flag.Float64Var(&hookLength, "hook-length", 1.5, "length of the hook itself")

	solid := model3d.JoinedSolid{
		model3d.ProfileSolid(
			&model2d.SubtractedSolid{
				Positive: model2d.JoinedSolid{
					model2d.NewRect(
						model2d.XY(-hangWidth/2, holeRadius),
						model2d.XY(hangWidth/2, hangLength),
					),
					&model2d.Circle{Center: model2d.Y(holeRadius), Radius: holeRadius + holeThickness},
				},
				Negative: &model2d.Circle{
					Center: model2d.Y(holeRadius),
					Radius: holeRadius,
				},
			},
			0,
			hangThickness,
		),
		HookSolid(hangLength, hookLength, hangWidth, hangThickness),
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	mesh.SaveGroupedSTL("hanging_hook.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func HookSolid(minY, maxZ, width, thickness float64) model3d.Solid {
	eval := func(t float64) model2d.Coord {
		return model2d.XY(minY-thickness*0.5+(0.25-math.Pow(t-0.5, 2))*maxZ, t*maxZ)
	}
	lines := model2d.NewMesh()
	eps := 0.01
	for t := 0.0; t+eps < 1.0; t += eps {
		lines.Add(&model2d.Segment{eval(t), eval(t + eps)})
	}
	collider := model2d.MeshToCollider(lines)
	return model3d.CheckedFuncSolid(
		model3d.XY(-thickness, minY-thickness),
		model3d.XYZ(thickness, minY+maxZ*0.25, maxZ),
		func(c model3d.Coord3D) bool {
			if math.Abs(c.X) > width/2 {
				return false
			}
			return collider.CircleCollision(c.YZ(), thickness/2)
		},
	)
}
