package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	NumSides   = 12
	BaseHeight = 0.4
	TipHeight  = 1.2
)

func main() {
	log.Println("Creating diamond polytope...")
	system := CreateDiamondPolytope()
	log.Println("Exporting diamond...")
	mesh := system.Mesh()
	mesh.SaveGroupedSTL("diamond.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)

	CreateStand(mesh)
}

func CreateDiamondPolytope() model3d.ConvexPolytope {
	system := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Coord3D{Z: -1},
			Max:    BaseHeight,
		},
	}

	addTriangle := func(t *model3d.Triangle) {
		n := t.Normal()

		// Make sure the normal points outward.
		if n.Dot(t[0]) < 0 {
			t[0], t[1] = t[1], t[0]
		}

		system = append(system, &model3d.LinearConstraint{
			Normal: t.Normal(),
			Max:    t[0].Dot(t.Normal()),
		})
	}

	iAngle := math.Pi * 2 / NumSides
	rimPoint := func(i int) model3d.Coord3D {
		return model3d.Coord3D{
			X: math.Cos(float64(i) * iAngle),
			Y: math.Sin(float64(i) * iAngle),
		}
	}
	basePoint := func(i int) model3d.Coord3D {
		return model3d.Coord3D{
			X: math.Cos((float64(i) + 0.5) * iAngle),
			Y: math.Sin((float64(i) + 0.5) * iAngle),
		}.Scale(1 - BaseHeight).Sub(model3d.Z(BaseHeight))
	}
	tipPoint := model3d.Z(TipHeight)

	for i := 0; i < NumSides; i++ {
		addTriangle(&model3d.Triangle{
			rimPoint(i),
			rimPoint(i + 1),
			tipPoint,
		})
		addTriangle(&model3d.Triangle{
			rimPoint(i),
			rimPoint(i + 1),
			basePoint(i),
		})
		addTriangle(&model3d.Triangle{
			basePoint(i),
			basePoint(i + 1),
			rimPoint(i + 1),
		})
	}
	return system
}

func CreateStand(diamond *model3d.Mesh) {
	log.Println("Creating stand...")
	diamond = diamond.MapCoords(model3d.XYZ(-1, 1, -1).Mul)
	solid := diamond.Solid()

	standSolid := &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P1:     model3d.Coord3D{Z: solid.Min().Z},
			P2:     model3d.Coord3D{Z: solid.Min().Z + 0.5},
			Radius: 1.0,
		},
		Negative: solid,
	}
	mesh := model3d.MarchingCubesSearch(standSolid, 0.01, 8)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh = mesh.FlattenBase(0)

	log.Println("Saving stand mesh...")
	mesh.SaveGroupedSTL("stand.stl")

	log.Println("Rendering stand mesh...")
	render3d.SaveRandomGrid("rendering_stand.png", mesh, 3, 3, 300, nil)
}
