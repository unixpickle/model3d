package main

import (
	"log"
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	BranchFactor = 5
	BranchDepth  = 3

	SameAngleFrac = 1.0
	HeightFrac    = 0.7

	RadiusFactor = 0.5
	LengthFactor = 0.6

	MinRadius = 0.01
)

func main() {
	trunk := &model3d.Cylinder{
		P1:     model3d.Coord3D{},
		P2:     model3d.Y(2),
		Radius: 0.15,
	}
	branches := CreateBranches(trunk, BranchDepth)
	solid := make(model3d.JoinedSolid, 0, len(branches)+1)
	solid = append(solid, trunk)
	for _, branch := range branches {
		solid = append(solid, branch)
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("tree.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func CreateBranches(branch *model3d.Cylinder, depthRemaining int) []*model3d.Cylinder {
	if depthRemaining == 0 {
		return nil
	}

	v := branch.P2.Sub(branch.P1)
	branchLen := v.Norm()
	v = v.Normalize()

	minDist := branch.Radius * RadiusFactor
	maxDist := branchLen - minDist
	minDist = maxDist - (maxDist-minDist)*HeightFrac

	basis1 := model3d.Coord3D{X: -v.Y, Y: v.X}
	if math.Abs(v.Z) > math.Abs(v.X) && math.Abs(v.Z) > math.Abs(v.Y) {
		basis1 = model3d.Coord3D{X: -v.Z, Z: v.X}
	}
	basis1 = basis1.Normalize()
	basis2 := (&model3d.Triangle{model3d.Coord3D{}, v, basis1}).Normal()

	initAngle := rand.Float64() * math.Pi * 2

	result := []*model3d.Cylinder{}
	for i := 0; i < BranchFactor; i++ {
		origin := branch.P1.Add(v.Scale(minDist + (maxDist-minDist)*rand.Float64()))
		theta := initAngle + math.Pi*2*float64(i)/BranchFactor
		theta += rand.Float64() / BranchFactor
		direction := basis1.Scale(math.Cos(theta)).Add(basis2.Scale(math.Sin(theta)))
		direction = direction.Add(v.Scale(SameAngleFrac * rand.Float64()))
		direction = direction.Normalize()

		sizeFrac := 0.5 * (rand.Float64() + 1)
		newLen := branchLen * LengthFactor * sizeFrac
		newBranch := &model3d.Cylinder{
			P1:     origin,
			P2:     origin.Add(direction.Scale(newLen)),
			Radius: branch.Radius * RadiusFactor * sizeFrac,
		}
		if newBranch.Radius < MinRadius {
			continue
		}
		result = append(result, newBranch)
		result = append(result, CreateBranches(newBranch, depthRemaining-1)...)
	}
	return result
}
