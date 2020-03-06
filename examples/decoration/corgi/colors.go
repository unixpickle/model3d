package main

import "github.com/unixpickle/model3d"

type Colorer struct {
	Solids []model3d.Solid
	Colors [][3]float64
}

func (c *Colorer) Add(s model3d.Solid, color [3]float64) {
	c.Solids = append(c.Solids, s)
	c.Colors = append(c.Colors, color)
}

func (c *Colorer) VertexColor(coord model3d.Coord3D) [3]float64 {
	for i, s := range c.Solids {
		if s.Contains(coord) {
			return c.Colors[i]
		}
	}
	return [3]float64{0, 0, 0}
}

func (c *Colorer) TriangleColor(t *model3d.Triangle) [3]float64 {
	var avg [3]float64
	for _, coord := range t {
		color := c.VertexColor(coord)
		for i, x := range color {
			avg[i] += x / 3
		}
	}
	return avg
}
