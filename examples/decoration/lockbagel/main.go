package main

import (
	"io/ioutil"
	"log"
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const BagelInnerRadius = 0.25

func main() {
	log.Println("Creating lock...")
	m1 := CreateLock()
	log.Println("Creating bagel...")
	m2 := CreateBagel()

	log.Println("Creating colors...")
	triToColor := map[*model3d.Triangle][3]float64{}
	m1.Iterate(func(t *model3d.Triangle) {
		if t[0].Y > 0.8 || t[1].Y > 0.8 || t[2].Y > 0.8 {
			triToColor[t] = [3]float64{0.1, 0.1, 0.1}
		} else {
			triToColor[t] = [3]float64{1, 234.0 / 255, 189.0 / 255}
		}
	})
	m2.Iterate(func(t *model3d.Triangle) {
		triToColor[t] = [3]float64{235.0 / 255, 168.0 / 255, 52.0 / 255}
		m1.Add(t)
	})

	log.Println("Exporting model...")
	colorFunc := func(t *model3d.Triangle) [3]float64 {
		return triToColor[t]
	}
	ioutil.WriteFile("mesh.zip", m1.EncodeMaterialOBJ(colorFunc), 0755)

	render3d.SaveRandomGrid("rendering.png", m1, 3, 3, 300, render3d.TriangleColorFunc(colorFunc))
}

func CreateLock() *model3d.Mesh {
	return model3d.MarchingCubesSearch(LockSolid{}, 0.006, 8)
}

type LockSolid struct{}

func (l LockSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -1, Y: -1, Z: -1}
}

func (l LockSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: 1, Y: 1, Z: 1}
}

func (l LockSolid) Contains(c model3d.Coord3D) bool {
	// Check the sides of the lock's hook.
	if c.Y < 0 && c.Y > -0.4 {
		d1 := c.Dist(model3d.Coord3D{X: -0.45, Y: c.Y, Z: 0})
		d2 := c.Dist(model3d.Coord3D{X: 0.45, Y: c.Y, Z: 0})
		return math.Min(d1, d2) < 0.1
	}

	// Check the body of the lock.
	if c.Y > 0 && c.Y < 1.0 {
		inset := 0.0
		if c.Y > 0.1 && c.Y < 0.8 && int(math.Round(c.Y*100))%10 < 5 {
			inset = 0.025
		}
		return c.X > -0.7 && c.X < 0.7 && c.Z > -0.3+inset && c.Z < 0.3-inset
	}

	// Check the top of the lock's hook.
	if c.Y < -0.4 && c.Y > -0.95 {
		theta := math.Atan2(c.Y+0.4, c.X)
		p := model3d.Coord3D{X: math.Cos(theta) * 0.45, Y: math.Sin(theta)*0.45 - 0.4, Z: 0}
		return p.Dist(c) < 0.1
	}

	return false
}

func CreateBagel() *model3d.Mesh {
	torus := &model3d.TorusSolid{
		Center:      model3d.Coord3D{Y: -0.85},
		Axis:        model3d.Coord3D{X: 1},
		InnerRadius: BagelInnerRadius,
		OuterRadius: 0.5,
	}
	return model3d.MarchingCubesSearch(torus, 0.005, 8)
}
