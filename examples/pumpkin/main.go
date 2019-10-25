package main

import (
	"io/ioutil"
	"log"
	"math"

	"github.com/unixpickle/model3d"
)

const Thickness = 0.9

func main() {
	pumpkin := PumpkinSolid()
	base := LidSolid{Solid: pumpkin}
	lid := LidSolid{IsLid: true, Solid: model3d.JoinedSolid{pumpkin, StemSolid{}}}

	log.Println("Creating base...")
	mesh1 := model3d.SolidToMesh(base, 0.05, 1, 0.8, 5)
	ioutil.WriteFile("base.stl", mesh1.EncodeSTL(), 0755)

	log.Println("Creating lid...")
	mesh2 := model3d.SolidToMesh(lid, 0.05, 1, 0.8, 5)
	ioutil.WriteFile("lid.stl", mesh2.EncodeSTL(), 0755)
}

func PumpkinSolid() model3d.Solid {
	mesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		return 1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat)
	}, 100)
	mesh.Iterate(func(t *model3d.Triangle) {
		t1 := *t
		for i, c := range t1 {
			t1[i] = c.Scale(Thickness)
		}
		t1[0], t1[1] = t1[1], t1[0]
		mesh.Add(&t1)
	})
	return model3d.NewColliderSolid(model3d.MeshToCollider(mesh))
}

type StemSolid struct{}

func (s StemSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{Y: 0.9, X: -0.3, Z: -0.3}
}

func (s StemSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{Y: 1.6, X: 0.3, Z: 0.3}
}

func (s StemSolid) Contains(c model3d.Coord3D) bool {
	if c.Max(s.Max()) != s.Max() || c.Min(s.Min()) != s.Min() {
		return false
	}
	c.X -= 0.15 * math.Pow(c.Y-s.Min().Y, 2)
	theta := math.Atan2(c.X, c.Z)
	radius := 0.05*math.Sin(theta*5) + 0.15
	return model3d.Coord2D{X: c.X, Y: c.Z}.Norm() < radius
}

type LidSolid struct {
	IsLid bool
	Solid model3d.Solid
}

func (l LidSolid) Min() model3d.Coord3D {
	return l.Solid.Min()
}

func (l LidSolid) Max() model3d.Coord3D {
	return l.Solid.Max()
}

func (l LidSolid) Contains(c model3d.Coord3D) bool {
	coneCenter := 0.0
	if l.IsLid {
		coneCenter += 0.1
	}
	inLid := model3d.Coord2D{X: c.X, Y: c.Z}.Norm() < 0.7*(c.Y-coneCenter)
	return inLid == l.IsLid && l.Solid.Contains(c)
}
