package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	CreamRadius  = 0.15
	CreamHeight  = 0.8
	CreamZOffset = 0.05
)

func CreamSolid() (model3d.Solid, toolbox3d.CoordColorFunc) {
	mesh := CreamMesh()
	return model3d.NewColliderSolid(model3d.MeshToCollider(mesh)),
		toolbox3d.ConstantCoordColorFunc(render3d.NewColor(1.0))
}

func CreamMesh() *model3d.Mesh {
	baseShape := model2d.CheckedFuncSolid(
		model2d.Ones(-CreamRadius),
		model2d.Ones(CreamRadius),
		func(c model2d.Coord) bool {
			norm := c.Norm()
			if norm < 1e-8 {
				return true
			}
			normed := c.Scale(1 / norm)
			pow := 2.5
			r := math.Pow(math.Pow(math.Abs(normed.X), pow)+math.Pow(math.Abs(normed.Y), pow), pow)
			return norm/CreamRadius < r
		},
	)
	length := 10.0
	solid3d := model3d.ProfileSolid(baseShape, 0, length)
	mesh3d := model3d.DualContour(solid3d, 0.03, true, false)

	// Deform along a curve.
	return mesh3d.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		centerCoord := creamCurve(c.Z / length)
		c2 := creamCurve(c.Z/length + 1e-5)
		derivative := c2.Sub(centerCoord).Normalize()
		x1 := model3d.Z(1).ProjectOut(derivative)
		x2 := x1.Cross(derivative)

		// Apply a small twist along the curve.
		rotation := c.Z / length * 25
		x1, x2 = x1.Scale(math.Cos(rotation)).Add(x2.Scale(math.Sin(rotation))),
			x1.Scale(-math.Sin(rotation)).Add(x2.Scale(math.Cos(rotation)))
		return centerCoord.Add(x1.Scale(c.Y)).Add(x2.Scale(c.X))
	})
}

func creamCurve(t float64) model3d.Coord3D {
	theta := t * 30
	// We want the start to naturally "tuck into" the spiral.
	radius := (CupTopRadius - CreamRadius) * (1 - math.Abs(t*1.1-0.1))
	return model3d.XYZ(
		math.Cos(theta)*radius,
		math.Sin(theta)*radius,
		CupHeight+t*CreamHeight-CreamZOffset,
	)
}
