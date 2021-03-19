package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	PitcherRadius = 1.8
	PitcherHeight = 5.0
	LipHeight     = 1.0
	LipRadius     = 0.5
	Thickness     = 0.2
)

func main() {
	model := PitcherExteriorSolid()
	mesh := model3d.MarchingCubesSearch(model, 0.02, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func PitcherExteriorSolid() model3d.Solid {
	model := PitcherExterior()
	squeeze := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   0,
		Max:   PitcherHeight - LipHeight - 0.1,
		Ratio: 0.1,
	}
	mesh := model3d.MarchingCubesConj(model, 0.02, 16, squeeze)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.Iterate(func(t *model3d.Triangle) {
		if t.Normal().Z > 0.95 {
			mesh.Remove(t)
		}
	})
	// TODO: figure out how to smooth out the top.
	// It is currently "rough" and jagged.
	return model3d.NewColliderSolidHollow(model3d.MeshToCollider(mesh), Thickness/2)
}

func PitcherExterior() model3d.Solid {
	base := &model3d.Cylinder{
		Radius: PitcherRadius,
		P2:     model3d.Z(PitcherHeight),
	}
	parab := Parabaloid()
	parab = model3d.TransformSolid(
		&model3d.Matrix3Transform{Matrix: &model3d.Matrix3{LipRadius, 0, 0, 0, LipRadius, 0, 0, 0, LipHeight}},
		parab,
	)
	parab = model3d.TransformSolid(
		&model3d.Translate{Offset: model3d.XZ(PitcherRadius, PitcherHeight-parab.Max().Z)},
		parab,
	)
	return model3d.JoinedSolid{
		base,
		parab,
	}
}

func Parabaloid() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1, -1, 0),
		model3d.XYZ(1, 1, 1),
		func(c model3d.Coord3D) bool {
			return c.Z > c.X*c.X+c.Y*c.Y
		},
	)
}
