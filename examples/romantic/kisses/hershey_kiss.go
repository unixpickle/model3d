package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	HersheyKissRadius       = 0.4
	HersheyKissHeight       = 0.7
	HersheyKissMinThickness = 0.05
	HersheyKissCurlStart    = 0.9
)

type HersheyKissSolid struct {
	Center model3d.Coord3D
}

func (h HersheyKissSolid) Min() model3d.Coord3D {
	return h.Center.Add(model3d.Coord3D{X: -HersheyKissRadius, Y: -HersheyKissRadius})
}

func (h HersheyKissSolid) Max() model3d.Coord3D {
	return h.Center.Add(model3d.Coord3D{
		X: HersheyKissRadius,
		Y: HersheyKissRadius,
		Z: HersheyKissHeight,
	})
}

func (h HersheyKissSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(h, c) {
		return false
	}
	y := (c.Z - h.Center.Z) / HersheyKissHeight
	if y > HersheyKissCurlStart {
		c.X += 4 * math.Pow(y-HersheyKissCurlStart, 2)
	}
	x := c.Sub(h.Center).Coord2D().Norm() / HersheyKissRadius
	if x > 1 {
		return false
	} else if x < HersheyKissMinThickness {
		// Solid cylinder center.
		return true
	}
	x = 1 - x
	hForX := 0.5 + math.Asin(2*math.Pow(x, 0.8)-1)/math.Pi
	return y <= hForX
}
