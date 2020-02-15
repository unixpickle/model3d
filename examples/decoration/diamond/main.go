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
	for i := 0; i < NumSides; i++ {
		theta := float64(i) * math.Pi * 2 / NumSides
		n1 := model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta), Z: -1}.Normalize()
		n2 := model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta), Z: 0.8}.Normalize()
		p := model3d.Coord3D{X: math.Cos(theta), Y: math.Sin(theta)}
		system = append(system,
			&model3d.LinearConstraint{
				Normal: n1,
				Max:    n1.Dot(p),
			},
			&model3d.LinearConstraint{
				Normal: n2,
				Max:    n2.Dot(p),
			},
		)
	}
	mesh := system.Mesh()
	mesh.SaveGroupedSTL("diamond.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}
