package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const AddEars = false

func main() {
	pumpkin := model3d.JoinedSolid{
		PumpkinSolid{},
		StemSolid{},
	}

	if AddEars {
		for _, sign := range []float64{-1, 1} {
			pumpkin = append(pumpkin, CreateEar(0.7*sign))
		}
	}

	log.Println("Creating and clipping base...")
	lowResMesh := model3d.MarchingCubesSearch(pumpkin, 0.05, 8)
	minY := lowResMesh.Min().Y + 0.05
	pumpkin = append(pumpkin, &model3d.CylinderSolid{
		P1:     model3d.Y(minY),
		P2:     model3d.Y(minY + 0.3),
		Radius: 0.4,
	})
	min := pumpkin.Min()
	min.Y = minY
	clippedSolid := model3d.ForceSolidBounds(pumpkin, min, pumpkin.Max())

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(clippedSolid, 0.01, 8)
	mesh = mesh.EliminateCoplanar(1e-5)
	mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		c.Z, c.Y = c.Y, -c.Z
		return c
	})

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("pumpkin.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type PumpkinSolid struct{}

func (p PumpkinSolid) Min() model3d.Coord3D {
	return model3d.XYZ(-2.0, -1.6, -1.6)
}

func (p PumpkinSolid) Max() model3d.Coord3D {
	return p.Min().Scale(-1)
}

func (p PumpkinSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	c.Y *= 0.9
	g := c.Geo()
	r := (1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat))
	return c.Norm() <= r
}

type StemSolid struct{}

func (s StemSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{Y: 1.0, X: -0.8, Z: -0.8}
}

func (s StemSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{Y: 1.9, X: 0.8, Z: 0.8}
}

func (s StemSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}
	c.X -= 0.15 * math.Pow(c.Y-s.Min().Y, 2)
	theta := math.Atan2(c.X, c.Z)
	radius := 0.05*math.Sin(theta*5) + 0.15
	radius += 0.7 * math.Pow(s.Max().Y-c.Y, 2)
	return c.XZ().Norm() < radius
}

func CreateEar(angle float64) model3d.Solid {
	curve := func(x float64) float64 {
		return -0.5*math.Pow(math.Abs(x), 2.5) - 0.2*x*x
	}
	point := func(x float64) model2d.Coord {
		return model2d.XY(x, curve(x))
	}
	mesh := model2d.NewMesh()
	var lastPoint model2d.Coord
	for x := -2.8; x <= 2.8; x += 0.001 {
		mesh.Add(&model2d.Segment{point(x), point(x + 0.001)})
		lastPoint = point(x + 0.001)
	}
	mesh.Add(&model2d.Segment{lastPoint, point(-2.8)})

	mesh = mesh.Translate(model2d.Y(-lastPoint.Y))
	mesh = mesh.Scale(-1 / lastPoint.Y).Scale(0.9)
	mesh = mesh.Translate(model2d.Y(1.3)).Rotate(angle)
	collider2d := model2d.MeshToCollider(mesh)
	collider3d := model3d.ProfileCollider(collider2d, -0.001, 0.001)
	return model3d.NewColliderSolidHollow(collider3d, 0.05)
}
