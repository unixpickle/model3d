package main

import (
	"math"

	"github.com/unixpickle/model3d"
)

const NumSides = 12

func main() {
	system := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Coord3D{Z: -1},
			Max:    0.4,
		},
	}
	iAngle := math.Pi * 2 / NumSides
	for i := 0; i < NumSides; i++ {
		theta := float64(i) * iAngle
		p1 := model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta)}
		p2 := model3d.Coord3D{X: math.Cos(theta + iAngle/2), Y: math.Sin(theta + iAngle/2)}
		n1 := model3d.Coord3D{X: p1.X, Y: p1.Y, Z: -0.8}.Normalize()
		n2 := model3d.Coord3D{X: p2.X, Y: p2.Y, Z: -1}.Normalize()
		n3 := model3d.Coord3D{X: p1.X, Y: p1.Y, Z: 0.8}.Normalize()
		system = append(system,
			&model3d.LinearConstraint{
				Normal: n1,
				Max:    n1.Dot(p1),
			},
			&model3d.LinearConstraint{
				Normal: n2,
				Max:    n2.Dot(p2) + 0.03,
			},
			&model3d.LinearConstraint{
				Normal: n3,
				Max:    n3.Dot(p1),
			},
		)
	}
	mesh := system.Mesh()
	mesh.SaveGroupedSTL("diamond.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}
