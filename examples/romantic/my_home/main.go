package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
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
	text = text.Scale(1.0 / 256.0).MapCoords(model2d.XY(-2, -2).Add)
	textSolid := model2d.NewColliderSolid(model2d.MeshToCollider(text))
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
		TriangularPolygon(
			thickness, true,
			model3d.XZ(-size, -size),
			model3d.XZ(-size, size),
			model3d.XZ(size, size),
			model3d.XZ(size, -size),
		),
		TriangularLine(thickness, model3d.X(-size+thickness/2), model3d.X(size-thickness/2)),
		TriangularLine(thickness, model3d.Z(-size+thickness/2), model3d.Z(size-thickness/2)),
	}
}

func DoorSolid() model3d.Solid {
	const size = 0.15
	return model3d.JoinedSolid{
		TriangularPolygon(
			0.02, false, model3d.XYZ(-size, 0.5, 0), model3d.XYZ(-size, 0.5, 0.45),
			model3d.XYZ(size, 0.5, 0.45), model3d.XYZ(size, 0.5, 0),
		),
		TriangularBall(0.03, model3d.XYZ(-size+0.07, 0.5, 0.45/2)),
	}
}

func TriangularPolygon(thickness float64, close bool, p ...model3d.Coord3D) model3d.Solid {
	res := model3d.JoinedSolid{}
	for i := 0; i < len(p)-1; i++ {
		res = append(res, TriangularLine(thickness, p[i], p[i+1]))
		if i != 0 {
			res = append(res, TriangularBall(thickness, p[i]))
		}
	}
	if close {
		res = append(
			res,
			TriangularLine(thickness, p[len(p)-1], p[0]),
			TriangularBall(thickness, p[len(p)-1]),
			TriangularBall(thickness, p[0]),
		)
	}
	return res.Optimize()
}

func TriangularLine(thickness float64, p1, p2 model3d.Coord3D) model3d.Solid {
	dir := p1.Sub(p2)
	length := dir.Norm()
	dir = dir.Normalize()

	// Basis vectors will be axis-aligned if dir is.
	b1, b2 := dir.OrthoBasis()

	ball := model3d.XYZ(thickness, thickness, thickness)
	return model3d.CheckedFuncSolid(
		p1.Min(p2).Sub(ball),
		p1.Max(p2).Add(ball),
		func(c model3d.Coord3D) bool {
			subtracted := c.Sub(p2)
			dot := subtracted.Dot(dir)
			if dot < 0 || dot > length {
				return false
			}
			dot1, dot2 := b1.Dot(subtracted), b2.Dot(subtracted)
			return math.Abs(dot1)+math.Abs(dot2) < thickness
		},
	)
}

func TriangularBall(thickness float64, p model3d.Coord3D) model3d.Solid {
	ball := model3d.XYZ(thickness, thickness, thickness)
	return model3d.CheckedFuncSolid(
		p.Sub(ball),
		p.Add(ball),
		func(c model3d.Coord3D) bool {
			diff := c.Sub(p)
			return math.Abs(diff.X)+math.Abs(diff.Y)+math.Abs(diff.Z) < thickness
		},
	)
}
