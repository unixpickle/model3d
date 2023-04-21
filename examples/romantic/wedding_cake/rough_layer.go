package main

import (
	"math/rand"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	RoughRoundLayerThickness = 0.8
	RoughRoundLayerRadius    = 0.6
	RoughRoundLayerNoise     = 0.01
	RoundRoughLayerGridSize  = 0.03
	HexRoughLayerThickness   = 0.5
	HexRoughLayerRadius      = 2.0
)

func RoughRoundLayer() (model3d.Solid, toolbox3d.CoordColorFunc) {
	cyl := &model3d.Cylinder{
		P1:     model3d.Origin,
		P2:     model3d.Z(RoughRoundLayerThickness),
		Radius: RoughRoundLayerRadius,
	}
	solid := MakeRoughShape(cyl)
	return solid, toolbox3d.ConstantCoordColorFunc(GoldDripColor)
}

func HexRoughLayer() (model3d.Solid, toolbox3d.CoordColorFunc) {
	hexMesh := model2d.NewMeshPolar(func(theta float64) float64 {
		return HexRoughLayerRadius
	}, 6)
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(hexMesh))
	solid3d := model3d.ProfileSolid(solid2d, 0.0, HexThickness)
	solid := MakeRoughShape(solid3d)
	return solid, toolbox3d.ConstantCoordColorFunc(GoldDripColor)
}

func MakeRoughShape(shape model3d.Solid) model3d.Solid {
	mesh := model3d.MarchingCubesSearch(shape, RoundRoughLayerGridSize, 8)
	normals := mesh.VertexNormals()
	mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		return c.Add(normals.Value(c).Scale(rand.NormFloat64() * RoughRoundLayerNoise))
	})
	mesh = model3d.LoopSubdivision(mesh, 1)
	solid := model3d.NewColliderSolid(model3d.MeshToCollider(mesh))
	min := solid.Min()
	min.Z = shape.Min().Z
	return model3d.ForceSolidBounds(solid, min, solid.Max())
}
