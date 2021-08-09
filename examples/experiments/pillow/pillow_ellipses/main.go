package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	shape := model2d.BezierCurve{
		model2d.XY(0, 0.6),
		model2d.XY(0.5, 1.1),
		model2d.XY(1.7, 1.1),
		model2d.XY(1.0, -0.9),
		model2d.XY(0, -0.9),
	}
	shapeMesh := model2d.NewMesh()
	shapeMesh.AddMesh(model2d.CurveMesh(shape, 100))
	shapeMesh.AddMesh(model2d.CurveMesh(shape, 100).MapCoords(model2d.XY(-1, 1).Mul))

	model2d.Rasterize("outline.png", shapeMesh.Scale(-1), 100.0)

	collider := model2d.MeshToCollider(shapeMesh)
	min, max := shapeMesh.Min(), shapeMesh.Max()
	center := model2d.Y(0.1)
	height := 0.5

	solid := model3d.CheckedFuncSolid(
		model3d.XYZ(min.X, min.Y, -1),
		model3d.XYZ(max.X, max.Y, 1),
		func(c model3d.Coord3D) bool {
			dir := c.XY().Sub(center).Normalize()
			if math.IsNaN(dir.Norm()) {
				// In the center.
				return math.Abs(c.Z) < height
			}
			rc, ok := collider.FirstRayCollision(&model2d.Ray{Origin: center, Direction: dir})
			if !ok {
				panic("no collision")
			}
			radius := rc.Scale
			centerDist := c.XY().Sub(center).Norm()
			maxHeight := height * math.Sqrt(1-centerDist*centerDist/(radius*radius))
			return math.Abs(c.Z) < maxHeight
		},
	)
	mesh := model3d.MarchingCubesSearch(solid, 0.03, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
	mesh.SaveGroupedSTL("out.stl")
}
