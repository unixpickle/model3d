package main

import (
	"image/png"
	"os"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	Radius = 2.0
	Outset = 0.1

	GrooveSize  = 0.05
	ScrewRadius = 0.2
	ScrewSlack  = 0.04

	BaseHeight = 0.2
	BaseRadius = 0.5
)

func main() {
	solid := &model3d.SubtractedSolid{
		Positive: GlobeSolid{
			Collider: RawGlobeCollider(),
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: -(Radius + Outset)},
			P2:         model3d.Coord3D{Z: Radius / 2},
			GrooveSize: GrooveSize,
			Radius:     ScrewRadius,
			Pointed:    true,
		},
	}
	split := &SplitSolid{Solid: solid, Top: true}
	topMesh := model3d.SolidToMesh(split, 0.01, 0, -1, 10)
	topMesh.SaveGroupedSTL("top.stl")
	model3d.SaveRandomGrid("top.png", model3d.MeshToCollider(topMesh), 3, 3, 300, 300)

	split.Top = false
	bottomMesh := model3d.SolidToMesh(split, 0.01, 0, -1, 10)
	bottomMesh = bottomMesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		c.Z, c.X = -c.Z, -c.X
		return c
	})
	bottomMesh.SaveGroupedSTL("bottom.stl")
	model3d.SaveRandomGrid("bottom.png", model3d.MeshToCollider(bottomMesh), 3, 3, 300, 300)

	base := model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: -(Radius + Outset + BaseHeight)},
			P2:     model3d.Coord3D{Z: -(Radius + Outset)},
			Radius: BaseRadius,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: -(Radius + Outset + BaseHeight)},
			P2:         model3d.Coord3D{Z: Radius/2 - ScrewSlack},
			GrooveSize: GrooveSize,
			Radius:     ScrewRadius - ScrewSlack,
			Pointed:    true,
		},
	}
	mesh := model3d.SolidToMesh(base, 0.01, 0, -1, 10)
	mesh.SaveGroupedSTL("base.stl")
}

func RawGlobeCollider() model3d.Collider {
	f, err := os.Open("map.png")
	essentials.Must(err)
	defer f.Close()
	img, err := png.Decode(f)
	essentials.Must(err)
	eq := model3d.NewEquirect(img)

	mesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		r, _, _, _ := eq.At(g).RGBA()
		if r > 0xffff/2 {
			return Radius + Outset
		}
		return Radius
	}, 300)
	mesh = mesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		c.Z, c.Y = c.Y, -c.Z
		return c
	})

	return model3d.MeshToCollider(mesh)
}

type GlobeSolid struct {
	Collider model3d.Collider
}

func (g GlobeSolid) Min() model3d.Coord3D {
	return g.Collider.Min()
}

func (g GlobeSolid) Max() model3d.Coord3D {
	return g.Collider.Max()
}

func (g GlobeSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(g.Min()) != g.Min() || c.Max(g.Max()) != g.Max() {
		return false
	}
	if c.Norm() < Radius {
		return true
	}
	dist := c.Norm() - Radius
	return g.Collider.RayCollisions(&model3d.Ray{
		Origin:    c,
		Direction: model3d.Coord3D{X: 1, Y: 1, Z: 1},
	})%2 != 0 && !g.Collider.SphereCollision(c, dist)
}

type SplitSolid struct {
	model3d.Solid
	Top bool
}

func (s *SplitSolid) Min() model3d.Coord3D {
	m := s.Solid.Min()
	if s.Top {
		m.Z = 0
	}
	return m
}

func (s *SplitSolid) Max() model3d.Coord3D {
	m := s.Solid.Max()
	if !s.Top {
		m.Z = 0
	}
	return m
}

func (s *SplitSolid) Contains(c model3d.Coord3D) bool {
	if s.Top {
		return c.Z >= 0 && s.Solid.Contains(c)
	} else {
		return c.Z < 0 && s.Solid.Contains(c)
	}
}
