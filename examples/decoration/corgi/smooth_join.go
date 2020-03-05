package main

import (
	"math"
	"sort"

	"github.com/unixpickle/model3d"
)

const (
	smoothJoinSDFDelta   = 0.04
	smoothJoinDelta      = 0.02
	smoothJoinConstraint = 0.5
)

// SmoothJoin joins solids and smooths them out so that
// they look cleanly joined together.
//
// The smoothDist argument controls how close to a joint a
// point should be to be considered for smoothing.
func SmoothJoin(constraint float64, solids ...model3d.Solid) model3d.Solid {
	sdfs := make([]model3d.SDF, len(solids))
	for i, solid := range solids {
		subMesh := model3d.SolidToMesh(solid, smoothJoinSDFDelta, 0, -1, 5)
		sdfs[i] = model3d.MeshToSDF(subMesh)
	}
	mesh := model3d.SolidToMesh(model3d.JoinedSolid(solids), smoothJoinDelta, 0, 0, 0)
	sdfCache := map[model3d.Coord3D]float64{}
	smoother := &model3d.MeshSmoother{
		StepSize:   0.1,
		Iterations: 100,
		ConstraintFunc: func(origin, newCoord model3d.Coord3D) model3d.Coord3D {
			var diff float64
			if d, ok := sdfCache[origin]; !ok {
				var distances []float64
				for _, s := range sdfs {
					distances = append(distances, math.Abs(s.SDF(origin)))
				}
				sort.Float64s(distances)
				diff = distances[len(distances)-1] - distances[len(distances)-2]
				sdfCache[origin] = diff
			} else {
				diff = d
			}
			dist := origin.Sub(newCoord).Norm()
			maxDist := smoothJoinDelta + math.Max(0, constraint-diff)
			if maxDist >= dist {
				return model3d.Coord3D{}
			}
			return origin.Sub(newCoord).Scale(smoothJoinConstraint * (dist - maxDist) / dist)
		},
	}
	mesh = smoother.Smooth(mesh)
	return model3d.NewColliderSolid(model3d.MeshToCollider(mesh))
}
