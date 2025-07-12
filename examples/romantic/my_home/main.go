package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	solid := model3d.JoinedSolid{
		BaseSolid(),
		HouseSolid(),
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh.SaveGroupedSTL("house.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func BaseSolid() model3d.Solid {
	text := model2d.MustReadBitmap("text.png", nil).FlipX().Mesh().SmoothSq(30)
	text = text.Scale(1.0 / 256.0).Translate(model2d.XY(-2, -2))
	textSolid := text.Solid()
	return model3d.JoinedSolid{
		model3d.NewRect(model3d.XYZ(-2, -2, -0.1), model3d.XYZ(2, 2, 0)),
		model3d.ProfileSolid(textSolid, 0, 0.1),
	}
}

func HouseSolid() model3d.Solid {
	window := WindowSolid()
	windows := model3d.JoinedSolid{}
	for _, y := range []float64{-0.5, 0.5} {
		windows = append(windows, model3d.TranslateSolid(window, model3d.YZ(y, 0.7)))
		for _, z := range []float64{0.25, 0.7} {
			for _, x := range []float64{-0.6, 0.6} {
				windows = append(windows, model3d.TranslateSolid(window, model3d.XYZ(x, y, z)))
			}
		}
	}

	return model3d.JoinedSolid{
		// Body of house.
		model3d.NewRect(model3d.XYZ(-1, -0.5, 0), model3d.XYZ(1, 0.5, 1)),
		RoofSolid(),

		DoorSolid(),
		ChimneySolid(),
		windows,
	}
}

func RoofSolid() model3d.Solid {
	prism := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.YZ(-1, 1).Normalize(),
			Max:    math.Sqrt2 / 4,
		},
		&model3d.LinearConstraint{
			Normal: model3d.YZ(1, 1).Normalize(),
			Max:    math.Sqrt2 / 4,
		},
		&model3d.LinearConstraint{
			Normal: model3d.Z(-1),
			Max:    0,
		},
		&model3d.LinearConstraint{
			Normal: model3d.X(-1),
			Max:    1,
		},
		&model3d.LinearConstraint{
			Normal: model3d.X(1),
			Max:    1,
		},
	}
	return model3d.TranslateSolid(prism.Solid(), model3d.Z(1))
}

func ChimneySolid() model3d.Solid {
	return model3d.NewRect(model3d.XY(0.5, 0.1), model3d.XYZ(0.65, 0.25, 1.6))
}

func WindowSolid() model3d.Solid {
	const size = 0.15
	const thickness = 0.015

	return model3d.JoinedSolid{
		toolbox3d.TriangularPolygon(
			thickness, true,
			model3d.XZ(-size, -size),
			model3d.XZ(-size, size),
			model3d.XZ(size, size),
			model3d.XZ(size, -size),
		),
		toolbox3d.TriangularLine(thickness, model3d.X(-size+thickness/2), model3d.X(size-thickness/2)),
		toolbox3d.TriangularLine(thickness, model3d.Z(-size+thickness/2), model3d.Z(size-thickness/2)),
	}
}

func DoorSolid() model3d.Solid {
	const size = 0.15
	return model3d.JoinedSolid{
		toolbox3d.TriangularPolygon(
			0.02, false, model3d.XYZ(-size, 0.5, 0), model3d.XYZ(-size, 0.5, 0.45),
			model3d.XYZ(size, 0.5, 0.45), model3d.XYZ(size, 0.5, 0),
		),
		toolbox3d.TriangularBall(0.03, model3d.XYZ(-size+0.07, 0.5, 0.45/2)),
	}
}
