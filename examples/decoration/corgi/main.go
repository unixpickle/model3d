package main

import (
	"log"

	"github.com/unixpickle/model3d"
)

const (
	BodyLength = 1.5
	BodyRadius = 0.3
	BodyToNeck = 0.2

	NeckLength = 0.5
	NeckRadius = 0.4

	HindLegWidth        = 0.5
	HindLegMuscleHeight = 0.6
	HindLegHeight       = 0.8
	HindLegThickness    = 0.4
	HindLegInset        = 0.15
	HindLegZ            = -0.02
	HindLegX            = -0.1
	HindLegRadius       = 0.04
)

func main() {
	log.Println("creating body solid...")
	model := SmoothJoin(0.1, MakeBody(), MakeHindLegs())
	log.Println("creating mesh...")
	mesh := model3d.SolidToMesh(model, 0.01, 0, -1, 5)
	log.Println("saving...")
	mesh.SaveGroupedSTL("corgi.stl")
	log.Println("rendering...")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

func MakeBody() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P2:     model3d.Coord3D{X: BodyLength},
			Radius: BodyRadius,
		},
		&model3d.SphereSolid{
			Radius: BodyRadius,
		},
		&model3d.SphereSolid{
			Center: model3d.Coord3D{X: BodyLength},
			Radius: BodyRadius,
		},
	}
}

func MakeHindLegs() model3d.Solid {
	return model3d.JoinedSolid{
		HindLegSolid{
			Center: model3d.Coord3D{X: HindLegX, Y: -BodyRadius + HindLegInset, Z: HindLegZ},
		},
		HindLegSolid{
			Center: model3d.Coord3D{X: HindLegX, Y: BodyRadius - HindLegInset, Z: HindLegZ},
		},
	}
}

type HindLegSolid struct {
	Center model3d.Coord3D
}

func (h HindLegSolid) Min() model3d.Coord3D {
	return h.Center.Add(model3d.Coord3D{X: -HindLegWidth / 2, Y: -HindLegThickness / 2,
		Z: HindLegMuscleHeight/2 - HindLegHeight})
}

func (h HindLegSolid) Max() model3d.Coord3D {
	return h.Center.Add(model3d.Coord3D{X: HindLegWidth / 2, Y: HindLegThickness / 2,
		Z: HindLegMuscleHeight / 2})
}

func (h HindLegSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(h, c) {
		return false
	}
	c = c.Sub(h.Center)
	muscleScale := model3d.Coord3D{X: 2 / HindLegWidth, Y: 2 / HindLegThickness,
		Z: 2 / HindLegMuscleHeight}
	if c.Mul(muscleScale).Norm() < 1 {
		return true
	}
	return c.Coord2D().Norm() < HindLegRadius
}
