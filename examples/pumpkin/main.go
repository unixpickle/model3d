package main

import (
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const Thickness = 0.9

func main() {
	r, err := os.Open("etching.png")
	essentials.Must(err)
	defer r.Close()
	etching, err := png.Decode(r)
	essentials.Must(err)

	pumpkin := &model3d.SubtractedSolid{
		Positive: PumpkinSolid{Scale: 1},
		Negative: PumpkinSolid{Scale: Thickness},
	}
	base := &model3d.SubtractedSolid{
		Positive: LidSolid{Solid: pumpkin},
		Negative: EtchSolid{
			Radius: 1.6,
			Height: 1.5,
			Image:  etching,
		},
	}
	lid := LidSolid{IsLid: true, Solid: model3d.JoinedSolid{pumpkin, StemSolid{}}}

	colorFunc := func(t *model3d.Triangle) [3]float64 {
		c := t[0]
		if (PumpkinSolid{Scale: 1.025}).Contains(c) {
			expectedNormal := t[0].Geo().Coord3D()
			if math.Abs(expectedNormal.Dot(t.Normal())) > 0.5 {
				return [3]float64{214.0 / 255, 143.0 / 255, 0}
			} else {
				return [3]float64{255.0 / 255, 206.0 / 255, 107.0 / 255}
			}
		}
		return [3]float64{79.0 / 255, 53.0 / 255, 0}
	}

	mesh := model3d.SolidToMesh(base, 0.05, 2, 0.8, 5)
	mesh.AddMesh(model3d.SolidToMesh(lid, 0.05, 2, 0.8, 5))
	ioutil.WriteFile("pumpkin.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
}

type PumpkinSolid struct {
	Scale float64
}

func (p PumpkinSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -p.Scale * 1.6, Y: -p.Scale * 1.6, Z: -p.Scale * 1.6}
}

func (p PumpkinSolid) Max() model3d.Coord3D {
	return p.Min().Scale(-1)
}

func (p PumpkinSolid) Contains(c model3d.Coord3D) bool {
	g := c.Geo()
	r := p.Scale * (1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat))
	return c.Norm() <= r
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

type EtchSolid struct {
	Image  image.Image
	Radius float64
	Height float64
}

func (e EtchSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -e.Radius, Y: -e.Height, Z: -e.Radius}
}

func (e EtchSolid) Max() model3d.Coord3D {
	return e.Min().Scale(-1)
}

func (e EtchSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(e.Min()) != e.Min() || c.Max(e.Max()) != e.Max() {
		return false
	}

	xFrac := c.Geo().Lon/(math.Pi*2) + 0.5
	yFrac := 1 - (c.Y+e.Height)/(e.Height*2)

	x := int(math.Round(xFrac * float64(e.Image.Bounds().Dx())))
	y := int(math.Round(yFrac * float64(e.Image.Bounds().Dy())))
	r, _, _, _ := e.Image.At(x, y).RGBA()
	return r < 0xffff/2
}
