package main

import (
	"log"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	wedge := Wedge{}

	createTopHole := func(c model3d.Coord3D, r float64) *model3d.Sphere {
		c.Z = wedge.ZForY(c.Y) + r/2
		return &model3d.Sphere{
			Center: c,
			Radius: r,
		}
	}

	createVerticalHole := func(c model3d.Coord3D, r float64) *model3d.Cylinder {
		return &model3d.Cylinder{
			P1:     model3d.XYZ(c.X, c.Y, -100),
			P2:     model3d.XYZ(c.X, c.Y, 100),
			Radius: r,
		}
	}

	createSideHole := func(c model3d.Coord3D, r float64) model3d.Solid {
		return model3d.JoinedSolid{
			&model3d.Sphere{
				Center: model3d.XYZ(-r*0.7, c.Y, c.Z),
				Radius: r,
			},
			&model3d.Sphere{
				Center: model3d.XYZ(wedge.Max().X+r*0.7, c.Y, c.Z),
				Radius: r,
			},
		}
	}

	createBackHole := func(c model3d.Coord3D, r float64) *model3d.Sphere {
		c.Y = wedge.YForZ(c.Z) + r*0.7
		return &model3d.Sphere{
			Center: c,
			Radius: r,
		}
	}

	solid := &model3d.SubtractedSolid{
		Positive: wedge,
		Negative: model3d.JoinedSolid{
			// Top holes (sorted by Y)
			createTopHole(model3d.XY(2.058, -0.1), 0.3),
			createVerticalHole(model3d.XY(2.5, 0.4), 0.2),
			createTopHole(model3d.XY(0.866, 0.6), 0.3),
			createTopHole(model3d.XY(1.5, 1.2), 0.4),
			createVerticalHole(model3d.XY(0.8, 1.21), 0.25),
			createTopHole(model3d.XY(1.4, 1.972), 0.3),
			createVerticalHole(model3d.XY(2.3, 2.2), 0.25),
			createTopHole(model3d.XY(0.7, 2.5), 0.3),
			createTopHole(model3d.XY(2.9, 3), 0.4),
			createTopHole(model3d.XY(1.640, 3.171), 0.3),
			createTopHole(model3d.XY(0.8, 3.3), 0.4),

			createSideHole(model3d.YZ(2, 0.5), 0.3),
			createSideHole(model3d.YZ(3.5, 1.2), 0.4),
			createSideHole(model3d.YZ(3.2, 0.6), 0.4),
			createSideHole(model3d.YZ(1, 0.7), 0.4),

			createBackHole(model3d.XZ(1.3, 0.8), 0.4),
			createBackHole(model3d.XZ(2.3, 1.2), 0.38),
		},
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("cheese.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type Wedge struct{}

func (w Wedge) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (w Wedge) Max() model3d.Coord3D {
	return model3d.XYZ(3, 4, 2)
}

func (w Wedge) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(w, c) {
		return false
	}
	return c.Z <= w.ZForY(c.Y)
}

func (w Wedge) ZForY(y float64) float64 {
	maxZ := 2.0          // sin(30 degrees) * 4
	maxY := 3.4641016151 // cos(30 degrees) * 4
	if y < maxY {
		yFrac := y / maxY
		return maxZ * yFrac
	} else {
		yFrac := (4.0 - y) / (4.0 - maxY)
		return maxZ * yFrac
	}
}

func (w Wedge) YForZ(z float64) float64 {
	// Return for the back, even though there's two
	// possible y values for each z.
	maxZ := 2.0          // sin(30 degrees) * 4
	maxY := 3.4641016151 // cos(30 degrees) * 4
	return maxY + (4-maxY)*(1-z/maxZ)
}
