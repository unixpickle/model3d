package main

import (
	"io/ioutil"
	"log"
	"math"

	"github.com/unixpickle/model3d"
)

const BagelInnerRadius = 0.2

func main() {
	log.Println("Creating lock...")
	m1 := CreateLock()
	log.Println("Creating bagel...")
	m2 := CreateBagel()

	log.Println("Creating colors...")
	triToColor := map[*model3d.Triangle][3]float64{}
	m1.Iterate(func(t *model3d.Triangle) {
		triToColor[t] = [3]float64{207 / 255, 144 / 255, 8 / 255}
	})
	m2.Iterate(func(t *model3d.Triangle) {
		var outside bool
		for _, p := range t {
			if (BagelSolid{}).InnerRadiusAt(p) > BagelInnerRadius {
				outside = true
				break
			}
		}
		if outside {
			triToColor[t] = [3]float64{1, 1, 1}
		} else {
			triToColor[t] = [3]float64{1, 234 / 255, 189 / 255}
		}
		m1.Add(t)
	})

	log.Println("Exporting model...")
	ioutil.WriteFile("mesh.stl", m1.EncodeSTL(), 0755)
}

func CreateLock() *model3d.Mesh {
	return model3d.SolidToMesh(LockSolid{}, 0.05, 3, 0.8, 3)
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
	if c.Y < 0 && c.Y > -0.5 {
		d1 := c.Dist(model3d.Coord3D{X: -0.35, Y: c.Y, Z: 0})
		d2 := c.Dist(model3d.Coord3D{X: 0.35, Y: c.Y, Z: 0})
		return math.Min(d1, d2) < 0.1
	}

	// Check the body of the lock.
	if c.Y > 0 && c.Y < 0.8 {
		return c.X > -0.6 && c.X < 0.6 && c.Z > -0.3 && c.Z < 0.3
	}

	// Check the top of the lock's hook.
	if c.Y < -0.5 && c.Y > -1 {
		theta := math.Atan2(c.Y+0.5, c.X)
		p := model3d.Coord3D{X: math.Cos(theta) * 0.35, Y: math.Sin(theta)*0.35 - 0.5, Z: 0}
		return p.Dist(c) < 0.1
	}

	return false
}

func CreateBagel() *model3d.Mesh {
	return model3d.SolidToMesh(BagelSolid{}, 0.025, 2, 0.8, 3)
}

type BagelSolid struct{}

func (b BagelSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -2, Y: -2, Z: -2}
}

func (b BagelSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: 2, Y: 2, Z: 2}
}

func (b BagelSolid) Contains(c model3d.Coord3D) bool {
	return b.InnerRadiusAt(c) < BagelInnerRadius
}

func (b BagelSolid) InnerRadiusAt(c model3d.Coord3D) float64 {
	theta := math.Atan2(c.Z, c.Y+0.85)
	p := model3d.Coord3D{X: 0, Y: math.Cos(theta)*0.5 - 0.85, Z: math.Sin(theta) * 0.5}
	return p.Dist(c)
}

func (b BagelSolid) At(outerTheta, innerTheta, radius float64) model3d.Coord3D {
	result := model3d.Coord3D{
		X: 0,
		Y: math.Cos(outerTheta)*0.5 - 0.85,
		Z: math.Sin(outerTheta) * 0.5,
	}
	result.X = math.Cos(innerTheta) * radius
	result.Y += math.Sin(innerTheta) * math.Cos(outerTheta) * radius
	result.Z += math.Sin(innerTheta) * math.Sin(outerTheta) * radius
	return result
}
