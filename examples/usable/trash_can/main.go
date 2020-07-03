package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	BottomRadius = 2.0
	TopRadius    = 3.0
	Height       = 6.0
	Thickness    = 0.2

	Bulge     = 0.2
	BulgeRate = 4.2

	SwirlRate  = 0.2
	SwirlPower = 1.2
)

func main() {
	log.Println("Generating swirl design...")
	mesh := model3d.MarchingCubesSearch(TrashCanSolid{}, Height/200, 8)
	log.Println("Saving swirl mesh...")
	mesh.SaveGroupedSTL("trash_swirl.stl")
	log.Println("Rendering swirl...")
	render3d.SaveRandomGrid("rendering_swirl.png", mesh, 3, 3, 200, nil)

	log.Println("Generating bulge design...")
	mesh = model3d.MarchingCubesSearch(TrashCanSolid{Bulge: true}, Height/200, 8)
	log.Println("Saving bulge mesh...")
	mesh.SaveGroupedSTL("trash_bulge.stl")
	log.Println("Rendering bulge...")
	render3d.SaveRandomGrid("rendering_bulge.png", mesh, 3, 3, 200, nil)
}

type TrashCanSolid struct {
	Bulge bool
}

func (t TrashCanSolid) Min() model3d.Coord3D {
	r := TopRadius + Bulge
	return model3d.Coord3D{X: -r, Y: -r}
}

func (t TrashCanSolid) Max() model3d.Coord3D {
	r := TopRadius + Bulge
	return model3d.XYZ(r, r, Height)
}

func (t TrashCanSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(t, c) {
		return false
	}
	frac := c.Z / Height
	radius := TopRadius*frac + BottomRadius*(1-frac)
	theta := math.Atan2(c.Y, c.X)
	if !t.Bulge {
		theta += SwirlRate * math.Pow(c.Z, SwirlPower)
	}
	distAround := theta * (TopRadius + BottomRadius) / 2

	bulgeRadius := radius
	if t.Bulge {
		bulgeRadius += 0.5 * Bulge * (math.Abs(math.Cos(BulgeRate*c.Z)) +
			math.Abs(math.Sin(BulgeRate*distAround)))
	} else {
		bulgeRadius += Bulge * (1 - math.Abs(math.Cos(BulgeRate*distAround)))
	}

	centerDist := c.Coord2D().Norm()
	return (c.Z < Thickness || centerDist > radius-Thickness) && centerDist < bulgeRadius
}
