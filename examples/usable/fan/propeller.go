package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func PropellerMesh() *model3d.Mesh {
	solid := PropellerSolid()
	mesh := model3d.MarchingCubesSearch(solid, 0.015, 8)
	return mesh
}

func PropellerSolid() model3d.Solid {
	positive := model3d.JoinedSolid{
		&model3d.Cylinder{
			P2:     model3d.Coord3D{Z: BladeDepth},
			Radius: PropellerHubRadius,
		},
	}
	for i := 0; i < BladeCount; i++ {
		positive = append(positive, BladeSolid{
			Theta: float64(i) * math.Pi * 2 / BladeCount,
		})
	}
	return &model3d.SubtractedSolid{
		Positive: positive,
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: -1e-3},
			P2:         model3d.Coord3D{Z: BladeDepth + 1e-3},
			Radius:     ScrewRadius,
			GrooveSize: ScrewGrooveSize,
		},
	}
}

type BladeSolid struct {
	Theta float64
}

func (b BladeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -BladeRadius, Y: -BladeRadius, Z: 0}
}

func (b BladeSolid) Max() model3d.Coord3D {
	return b.Min().Scale(-1).Add(model3d.Coord3D{Z: BladeDepth})
}

func (b BladeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	c2 := c.Coord2D()
	if c2.Norm() < PropellerHubRadius || c2.Norm() > BladeRadius {
		return false
	}

	vec := model3d.Coord2D{X: math.Cos(b.Theta), Y: math.Sin(b.Theta)}
	if vec.Dot(c2) < 0 {
		return false
	}
	normal := model3d.Coord3D{X: -vec.Y, Y: vec.X, Z: 1}.Normalize()

	dist := math.Abs(c.Sub(model3d.Coord3D{Z: BladeDepth / 2}).Dot(normal))
	return dist <= BladeThickness/2
}
