package main

import (
	"math/rand"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	. "github.com/unixpickle/model3d/shorthand"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	RoughRoundLayerThickness = 1.0
	RoughRoundLayerRadius    = 0.6
	RoughRoundLayerNoise     = 0.01
	RoundRoughLayerGridSize  = 0.03
	HexRoughLayerThickness   = 1.0
	HexRoughLayerRadius      = 2.0
)

func RoughRoundLayer() (Solid3, toolbox3d.CoordColorFunc) {
	cyl := Cylinder(
		Origin3,
		Z(RoughRoundLayerThickness),
		RoughRoundLayerRadius,
	)
	solid := MakeRoughShape(cyl)
	return solid, toolbox3d.ConstantCoordColorFunc(GoldDripColor)
}

func HexRoughLayer() (Solid3, toolbox3d.CoordColorFunc) {
	hexMesh := model2d.NewMeshPolar(func(theta float64) float64 {
		return HexRoughLayerRadius
	}, 6)
	solid2d := hexMesh.Solid()
	solid3d := model3d.ProfileSolid(solid2d, 0.0, HexThickness)
	solid := MakeRoughShape(solid3d)
	return solid, toolbox3d.ConstantCoordColorFunc(GoldDripColor)
}

func MakeRoughShape(shape Solid3) Solid3 {
	mesh := model3d.MarchingCubesSearch(shape, RoundRoughLayerGridSize, 8)
	normals := mesh.VertexNormals()
	mesh = mesh.MapCoords(func(c C3) C3 {
		return c.Add(normals.Value(c).Scale(rand.NormFloat64() * RoughRoundLayerNoise))
	})
	mesh = model3d.LoopSubdivision(mesh, 1)
	solid := mesh.Solid()
	min := solid.Min()
	min.Z = shape.Min().Z
	return model3d.ForceSolidBounds(solid, min, solid.Max())
}
