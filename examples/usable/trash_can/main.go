package main

import (
	"io/ioutil"
	"log"
	"math"

	"github.com/unixpickle/model3d"
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
	mesh := model3d.SolidToMesh(TrashCanSolid{}, Height/100, 1, 0.8, 5)
	ioutil.WriteFile("trash_swirl.stl", mesh.EncodeSTL(), 0755)
	model3d.SaveRandomGrid("rendering_swirl.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)

	log.Println("Generating bulge design...")
	mesh = model3d.SolidToMesh(TrashCanSolid{Bulge: true}, Height/100, 1, 0.8, 5)
	ioutil.WriteFile("trash_bulge.stl", mesh.EncodeSTL(), 0755)
	model3d.SaveRandomGrid("rendering_bulge.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
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
	return model3d.Coord3D{X: r, Y: r, Z: Height}
}

func (t TrashCanSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(t.Min()) != t.Min() || c.Max(t.Max()) != t.Max() {
		return false
	}
	frac := c.Z / Height
	radius := TopRadius*frac + BottomRadius*(1-frac)
	theta := math.Atan2(c.Y, c.X)
	if !t.Bulge {
		theta += SwirlRate * math.Pow(c.Z, SwirlPower)
	}
	distAround := theta * (TopRadius + BottomRadius) / 2

	centerDist := (model3d.Coord2D{X: c.X, Y: c.Y}).Norm()
	bulgeRadius := radius
	if t.Bulge {
		bulgeRadius += 0.5 * Bulge * (math.Abs(math.Cos(BulgeRate*c.Z)) +
			math.Abs(math.Sin(BulgeRate*distAround)))
	} else {
		bulgeRadius += Bulge * (1 - math.Abs(math.Cos(BulgeRate*distAround)))
	}
	return (c.Z < Thickness || centerDist > radius-Thickness) && centerDist < bulgeRadius
}
