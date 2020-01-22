package main

import (
	"log"
	"math"
	"math/rand"
	"sort"

	"github.com/unixpickle/model3d"
)

const (
	AdditionSamples = 1000
	KeepSamples     = 990
	GridSize        = 3
)

func main() {
	log.Println("Creating rocks...")
	var rocks model3d.JoinedSolid
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			min := model3d.Coord3D{X: float64(i) * 1.1, Y: float64(j) * 1.1, Z: 0}
			rock := &RockSolid{
				Solid: &model3d.RectSolid{
					MinVal: min,
					MaxVal: min.Add(model3d.Coord3D{X: 1, Y: 1, Z: rand.Float64()*0.1 + 0.95}),
				},
			}
			for i := 0; i < 100; i++ {
				rock.AddConstraint()
			}
			rocks = append(rocks, rock)
		}
	}

	log.Println("Creating mesh...")
	mesh := model3d.SolidToMesh(rocks, 0.01, 0, -1, 5)
	log.Println("Lasso mesh...")
	for i := 0; i < 5; i++ {
		mesh = mesh.LassoSolid(rocks, 0.01, 3, 0, 0.2)
	}
	log.Println("Saving results...")
	mesh.SaveGroupedSTL("rock.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

type RockSolid struct {
	model3d.Solid
	Constraints []*LinearConstraint
}

func (r *RockSolid) AddConstraint() {
	lc := &LinearConstraint{
		Normal: model3d.Coord3D{
			X: rand.NormFloat64(),
			Y: rand.NormFloat64(),
			Z: math.Abs(rand.NormFloat64()),
		},
	}
	var dotValues []float64
	for len(dotValues) < AdditionSamples {
		min := r.Min()
		scale := r.Max().Sub(min)
		c := model3d.Coord3D{
			X: rand.Float64()*scale.X + min.X,
			Y: rand.Float64()*scale.Y + min.Y,
			Z: rand.Float64()*scale.Z + min.Z,
		}
		if r.Contains(c) {
			dotValues = append(dotValues, c.Dot(lc.Normal))
		}
	}
	sort.Float64s(dotValues)
	lc.MaxDot = dotValues[KeepSamples]
	r.Constraints = append(r.Constraints, lc)
}

func (r *RockSolid) Contains(c model3d.Coord3D) bool {
	if !r.Solid.Contains(c) {
		return false
	}
	for _, constraint := range r.Constraints {
		if !constraint.Satisfies(c) {
			return false
		}
	}
	return true
}

type LinearConstraint struct {
	Normal model3d.Coord3D
	MaxDot float64
}

func (l *LinearConstraint) Satisfies(c model3d.Coord3D) bool {
	return l.Normal.Dot(c) <= l.MaxDot
}
