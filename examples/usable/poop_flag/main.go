package main

import (
	"log"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d"
)

const (
	Radius = 1.2
	Height = Radius

	ScrewRadius     = 0.3
	ScrewGrooveSize = 0.05
	ScrewSlack      = 0.04

	SteakHeight = 3.0
)

func main() {
	screw := &toolbox3d.ScrewSolid{
		P1:         model3d.Coord3D{Z: -SteakHeight},
		P2:         model3d.Coord3D{Z: Height / 2},
		Radius:     ScrewRadius,
		GrooveSize: ScrewGrooveSize,
	}
	pointedScrew := PointedSolid{
		Solid:  screw,
		Center: model3d.Coord3D{Z: Height / 2},
	}

	log.Println("Building poop solid...")
	poop := &model3d.SubtractedSolid{
		Positive: PoopSolid(),
		Negative: pointedScrew,
	}

	log.Println("Building poop mesh...")
	mesh := model3d.SolidToMesh(poop, 0.01, 0, -1, 10)
	mesh.SaveGroupedSTL("poop.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)

	log.Println("Building steak mesh...")
	screw.P1, screw.P2 = screw.P2, screw.P1
	screw.Radius -= ScrewSlack
	mesh = model3d.SolidToMesh(pointedScrew, 0.01, 0, -1, 10)
	mesh.SaveGroupedSTL("steak.stl")
	model3d.SaveRandomGrid("steak.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}

func PoopSolid() model3d.Solid {
	tip := &PointedSolid{
		Solid: &model3d.RectSolid{
			MinVal: model3d.Coord3D{Z: Radius * 0.85, X: -Radius * 0.3, Y: -Radius * 0.3},
			MaxVal: model3d.Coord3D{Z: Radius * 1.15, X: Radius * 0.3, Y: Radius * 0.3},
		},
		Center: model3d.Coord3D{Z: Radius * 1.15},
	}
	constructed := model3d.JoinedSolid{
		tip,
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: -Radius * 0.25},
			P2:     model3d.Coord3D{Z: Radius * 0.25},
			Radius: Radius,
		},
		&model3d.TorusSolid{
			Axis:        model3d.Coord3D{Z: 1},
			OuterRadius: Radius,
			InnerRadius: 0.25 * Radius,
		},
		&model3d.TorusSolid{
			Center:      model3d.Coord3D{Z: Radius * 0.35},
			Axis:        model3d.Coord3D{Z: 1},
			OuterRadius: Radius * 0.7,
			InnerRadius: 0.25 * Radius,
		},
		&model3d.TorusSolid{
			Center:      model3d.Coord3D{Z: Radius * 0.65},
			Axis:        model3d.Coord3D{Z: 1},
			OuterRadius: Radius * 0.35,
			InnerRadius: 0.25 * Radius,
		},
	}
	mesh := model3d.SolidToMesh(constructed, 0.011, 0, -1, 5)
	for i := 0; i < 5; i++ {
		log.Printf("- lasso solid %d/5...", i)
		mesh = mesh.LassoSolid(constructed, 0.005, 4, 100, 0.2)
	}

	// Remove internal gaps, leaving only the external shell.
	topTriangle := &model3d.Triangle{}
	mesh.Iterate(func(t *model3d.Triangle) {
		if t.Max().Z > topTriangle.Max().Z {
			topTriangle = t
		}
	})
	searched := map[*model3d.Triangle]bool{topTriangle: true}
	toSearch := []*model3d.Triangle{topTriangle}
	for len(toSearch) > 0 {
		t := toSearch[0]
		toSearch = toSearch[1:]
		for _, n := range mesh.Neighbors(t) {
			if !searched[n] {
				searched[n] = true
				toSearch = append(toSearch, n)
			}
		}
	}
	mesh.Iterate(func(t *model3d.Triangle) {
		if !searched[t] {
			mesh.Remove(t)
		}
	})

	// Clip off bottom to remove need for supports, and
	// to avoid a rough surface that occurs due to
	// rounding.
	return &model3d.SubtractedSolid{
		Positive: model3d.NewColliderSolid(model3d.MeshToCollider(mesh)),
		Negative: &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -Radius * 2, Y: -Radius * 2, Z: -Radius * 0.3},
			MaxVal: model3d.Coord3D{X: Radius * 2, Y: Radius * 2, Z: -Radius * 0.15},
		},
	}
}

type PointedSolid struct {
	model3d.Solid
	Center model3d.Coord3D
}

func (p PointedSolid) Contains(c model3d.Coord3D) bool {
	if !p.Solid.Contains(c) {
		return false
	}
	radius := p.Center.Z - c.Z
	if radius < 0 {
		return false
	}
	c1 := c.Sub(p.Center)
	c1.Z = 0
	return c1.Norm() <= radius
}
