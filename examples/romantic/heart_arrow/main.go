package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	DeltaT      = 0.001
	Slope       = 0.015
	Radius      = 0.2
	ArrowLength = 0.5
)

func main() {
	curve := HeartCurve()
	centerMesh := model3d.NewMesh()
	curZ := 0.0
	var lastSeg [2]model3d.Coord3D
	for t := 0.0; t+DeltaT < 1.0; t += DeltaT {
		p1 := curve.Eval(t)
		p2 := curve.Eval(t + DeltaT)
		meshP1 := model3d.XYZ(p1.X, p1.Y, curZ)
		curZ += p2.Dist(p1) * Slope
		meshP2 := model3d.XYZ(p2.X, p2.Y, curZ)
		lastSeg = [2]model3d.Coord3D{meshP1, meshP2}
		centerMesh.Add(&model3d.Triangle{meshP1, meshP2, meshP1})
	}

	// Create tip.
	orthog := lastSeg[1].Sub(lastSeg[0])
	direction := lastSeg[0].Sub(lastSeg[1])
	direction.Z = 0
	orthog.X, orthog.Y, orthog.Z = orthog.Y, -orthog.X, 0
	direction1 := direction.Add(orthog.Scale(0.5)).Normalize().Scale(ArrowLength)
	direction2 := direction.Reflect(direction1)
	direction1.Z = -Slope * direction1.Dot(direction.Normalize())
	direction2.Z = direction1.Z
	centerMesh.Add(&model3d.Triangle{
		lastSeg[1],
		lastSeg[1].Add(direction1),
		lastSeg[1].Add(direction2),
	})

	heartSolid := model3d.NewColliderSolidHollow(model3d.MeshToCollider(centerMesh), Radius)
	heartMesh := model3d.MarchingCubesSearch(heartSolid, 0.02, 8)
	heartMesh.SaveGroupedSTL("heart_arrow.stl")
	render3d.SaveRandomGrid("rendering.png", heartMesh, 3, 3, 300, nil)
}

func HeartCurve() model2d.JoinedCurve {
	return model2d.JoinedCurve{
		model2d.BezierCurve{
			model2d.XY(1.5, -1.5),
			model2d.XY(-4, 0),
			model2d.XY(-1, 4),
			model2d.XY(0, 1.5),
		},
		model2d.BezierCurve{
			model2d.XY(0, 1.5),
			model2d.XY(1, 4),
			model2d.XY(2.5, 1.5),
			model2d.XY(1.5, 0),
		},
		model2d.BezierCurve{
			model2d.XY(1.5, 0),
			model2d.XY(0.5, -1.5),
			model2d.XY(0.5, -1.5),
			model2d.XY(-0.5, -3.0),
		},
	}
}
