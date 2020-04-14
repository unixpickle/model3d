package main

import (
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	AdditionSamples = 1000
	KeepSamples     = 990
	GridSize        = 10
)

func main() {
	log.Println("Creating rocks...")
	mesh := model3d.NewMesh()
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			rock := model3d.ConvexPolytope{
				&model3d.LinearConstraint{
					Normal: model3d.Coord3D{X: -1},
					Max:    0,
				},
				&model3d.LinearConstraint{
					Normal: model3d.Coord3D{X: 1},
					Max:    1,
				},
				&model3d.LinearConstraint{
					Normal: model3d.Coord3D{Y: -1},
					Max:    0,
				},
				&model3d.LinearConstraint{
					Normal: model3d.Coord3D{Y: 1},
					Max:    1,
				},
				&model3d.LinearConstraint{
					Normal: model3d.Coord3D{Z: -1},
					Max:    0,
				},
				&model3d.LinearConstraint{
					Normal: model3d.Coord3D{Z: 1},
					Max:    1,
				},
			}
			for i := 0; i < 100; i++ {
				rock = AddConstraint(rock)
			}
			min := model3d.Coord3D{X: float64(i)*5.0 + rand.Float64()*4, Y: float64(j)*5.0 + rand.Float64()*4, Z: 0}
			mesh.AddMesh(rock.Mesh().MapCoords(min.Add))
		}
	}

	log.Println("Saving results...")
	mesh.SaveGroupedSTL("rock.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func AddConstraint(r model3d.ConvexPolytope) model3d.ConvexPolytope {
	lc := &model3d.LinearConstraint{
		Normal: model3d.Coord3D{
			X: rand.NormFloat64(),
			Y: rand.NormFloat64(),
			Z: math.Abs(rand.NormFloat64()),
		},
	}
	var dotValues []float64
	for len(dotValues) < AdditionSamples {
		c := model3d.Coord3D{
			X: rand.Float64(),
			Y: rand.Float64(),
			Z: rand.Float64(),
		}
		if r.Contains(c) {
			dotValues = append(dotValues, c.Dot(lc.Normal))
		}
	}
	sort.Float64s(dotValues)
	lc.Max = dotValues[KeepSamples]
	return append(r, lc)
}
