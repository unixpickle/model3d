package main

import (
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

	createSideHole := func(c model3d.Coord3D, r float64) model3d.Solid {
		return model3d.JoinedSolid{
			&model3d.Sphere{
				Center: model3d.XYZ(-r*0.7, c.Y, c.Z),
				Radius: r,
			},
			&model3d.Sphere{
				Center: model3d.XYZ(wedge.Max().Z+r*0.7, c.Y, c.Z),
				Radius: r,
			},
		}
	}

	solid := &model3d.SubtractedSolid{
		Positive: wedge,
		Negative: model3d.JoinedSolid{
			createTopHole(model3d.XY(0.8, 4), 0.5),
			createTopHole(model3d.XY(2.3, 2), 0.3),
			createTopHole(model3d.XY(1.5, 1.2), 0.4),
			createTopHole(model3d.XY(2.9, 3), 0.4),
			createTopHole(model3d.XY(3.1, 4), 0.3),
			createTopHole(model3d.XY(0.5, 3.2), 0.35),
			createTopHole(model3d.XY(0.7, 2.5), 0.15),
			createTopHole(model3d.XY(1.640, 3.171), 0.3),
			createTopHole(model3d.XY(2.796, 1.213), 0.4),
			createTopHole(model3d.XY(0.866, 0.6), 0.3),
			createTopHole(model3d.XY(2.058, -0.1), 0.3),
			createTopHole(model3d.XY(1.842, 1.972), 0.2),
			createSideHole(model3d.YZ(2, 0.5), 0.3),
			createSideHole(model3d.YZ(3.5, 1.5), 0.4),
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
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
	yFrac := y / w.Max().Y
	return yFrac * w.Max().Z
}
