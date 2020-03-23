package main

import (
	"image/png"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Radius = 2.0
	Outset = 0.1

	DowelSize   = 0.2
	DowelSlack  = 0.03
	DowelHeight = 0.5
)

func main() {
	dowel := &model3d.RectSolid{
		MinVal: model3d.Coord3D{
			X: -DowelSize,
			Y: -DowelSize,
			Z: -DowelHeight,
		},
		MaxVal: model3d.Coord3D{
			X: DowelSize,
			Y: DowelSize,
			Z: DowelHeight,
		},
	}

	solid := &model3d.SubtractedSolid{
		Positive: GlobeSolid{
			Collider: RawGlobeCollider(),
		},
		Negative: &toolbox3d.Ramp{
			Solid: &toolbox3d.Ramp{
				Solid: dowel,
				P1:    model3d.Coord3D{Z: DowelHeight},
				P2:    model3d.Coord3D{Z: DowelHeight - DowelSize},
			},
			P1: model3d.Coord3D{Z: -DowelHeight},
			P2: model3d.Coord3D{Z: -(DowelHeight - DowelSize)},
		},
	}
	split := &SplitSolid{Solid: solid, Top: true}
	topMesh := model3d.SolidToMesh(split, 0.01, 0, -1, 20)
	topMesh.SaveGroupedSTL("top.stl")
	model3d.SaveRandomGrid("top.png", model3d.MeshToCollider(topMesh), 3, 3, 300, 300)

	split.Top = false
	bottomMesh := model3d.SolidToMesh(split, 0.01, 0, -1, 20)
	bottomMesh = bottomMesh.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		c.Z, c.X = -c.Z, -c.X
		return c
	})
	bottomMesh.SaveGroupedSTL("bottom.stl")
	model3d.SaveRandomGrid("bottom.png", model3d.MeshToCollider(bottomMesh), 3, 3, 300, 300)

	dowel.MinVal.X += DowelSlack / 2
	dowel.MinVal.Y += DowelSlack / 2
	dowel.MaxVal.X -= DowelSlack / 2
	dowel.MaxVal.Y -= DowelSlack / 2
	// Accommodate for pointed tip.
	dowel.MaxVal.Z -= DowelSize*2 + DowelSlack
	mesh := model3d.SolidToMesh(dowel, 0.01, 0, -1, 10)
	mesh = mesh.FlattenBase(math.Pi * 0.49)
	mesh.SaveGroupedSTL("dowel.stl")
}

func RawGlobeCollider() model3d.Collider {
	f, err := os.Open("map.png")
	essentials.Must(err)
	defer f.Close()
	img, err := png.Decode(f)
	essentials.Must(err)
	eq := toolbox3d.NewEquirect(img)

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
	norm := c.Norm()
	if norm < Radius {
		return true
	} else if norm > Radius+Outset {
		return false
	}
	dist := norm - Radius
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
