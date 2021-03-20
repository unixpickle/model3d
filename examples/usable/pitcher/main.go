package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	PitcherRadius = 1.8
	PitcherHeight = 5.0
	LipHeight     = 1.0
	LipRadius     = 0.5
	LipInset      = 0.1
	HandleHeight  = 2.5
	HandleWidth   = 0.3
	Thickness     = 0.2
)

func main() {
	log.Println("Creating solid...")
	model := PitcherExteriorSolid()

	log.Println("Creating mesh...")
	squeeze := toolbox3d.NewSmartSqueeze(toolbox3d.AxisZ, 0, 0.04, 0)
	squeeze.AddUnsqueezable(PitcherHeight-LipHeight-0.1, math.Inf(1))
	squeeze.AddUnsqueezable(PitcherHeight/2-HandleHeight/2-0.1, PitcherHeight/2+HandleHeight/2+0.1)
	squeeze.AddPinch(-Thickness / 2)
	mesh := squeeze.MarchingCubesSearch(model, 0.02, 8)

	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("pitcher.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func PitcherExteriorSolid() model3d.Solid {
	model := PitcherExterior()
	squeeze := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   0.1,
		Max:   PitcherHeight - LipHeight - 0.1,
		Ratio: 0.1,
	}
	mesh := model3d.MarchingCubesConj(model, 0.02, 16, squeeze)
	mesh = mesh.EliminateCoplanar(1e-5)

	// Fix rough edges around the rim
	mesh = FlattenTop(mesh)

	mesh.Iterate(func(t *model3d.Triangle) {
		if t.Normal().Z > 0.95 {
			mesh.Remove(t)
		}
	})
	rounded := model3d.NewColliderSolidHollow(model3d.MeshToCollider(mesh), Thickness/2)
	return model3d.JoinedSolid{
		rounded,
		PitcherHandle(),
		// Solid cylinder at the base instead of a
		// rounded base (which needs support).
		&model3d.Cylinder{
			P1:     model3d.Z(-Thickness / 2),
			P2:     model3d.Z(Thickness / 2),
			Radius: PitcherRadius + Thickness/2,
		},
	}
}

func PitcherExterior() model3d.Solid {
	base := &model3d.Cylinder{
		Radius: PitcherRadius,
		P2:     model3d.Z(PitcherHeight),
	}
	lip := Pyramid()
	lip = model3d.TransformSolid(
		&model3d.Matrix3Transform{
			Matrix: &model3d.Matrix3{
				LipRadius, 0, 0, 0, LipRadius, 0, 0, 0, LipHeight,
			},
		},
		lip,
	)
	lip = model3d.TransformSolid(
		&model3d.Translate{Offset: model3d.XZ(-PitcherRadius+LipInset, PitcherHeight-lip.Max().Z)},
		lip,
	)
	return model3d.JoinedSolid{
		base,
		lip,
	}
}

func Pyramid() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1, -1, 0),
		model3d.XYZ(1, 1, 1),
		func(c model3d.Coord3D) bool {
			return c.Z > math.Abs(c.X)+math.Abs(c.Y)
		},
	)
}

func PitcherHandle() model3d.Solid {
	x := PitcherRadius
	z := PitcherHeight / 2
	inset := Thickness * math.Sqrt(2)
	return model3d.CheckedFuncSolid(
		model3d.XYZ(x-Thickness/2, -HandleWidth/2, z-HandleHeight/2),
		model3d.XYZ(x+HandleHeight/2, HandleWidth/2, z+HandleHeight/2),
		func(c model3d.Coord3D) bool {
			zDist := math.Abs(c.Z - z)
			curX := x + HandleHeight/2 - zDist
			return c.X < curX && c.X > curX-inset
		},
	)
}

func FlattenTop(m *model3d.Mesh) *model3d.Mesh {
	rotation := model3d.Rotation(model3d.X(1), math.Pi)
	m = m.MapCoords(rotation.Apply)
	m = m.FlattenBase(0.99 * math.Pi / 2)
	m = m.MapCoords(rotation.Inverse().Apply)
	return m
}
