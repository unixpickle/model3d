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

var Green = [3]float64{
	float64(0x1b) / 255.0,
	float64(0xad) / 255.0,
	float64(0x64) / 255.0,
}

const (
	PickleLength = 2.0
	PickleWidth  = PickleLength / 2
)

func main() {
	solid := &PickleSolid{F: NewPickleFunction()}
	mesh := model3d.SolidToMesh(solid, 0.1, 4, 0.8, 8)

	colorFunc := NewInscription().ColorAt
	ioutil.WriteFile("pickle.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
}

type PickleSolid struct {
	F *PickleFunction
}

func (p *PickleSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{Z: -PickleWidth}
}

func (p *PickleSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: PickleWidth, Y: PickleLength, Z: PickleWidth}
}

func (p *PickleSolid) Contains(c model3d.Coord3D) bool {
	radius := p.F.RadiusAt(c.Y)
	center := p.F.CenterAt(c.Y)
	dist := c.Dist(model3d.Coord3D{X: center, Y: c.Y})
	return dist < radius
}

type PickleFunction struct {
	image image.Image
	cache map[int][2]float64
}

func NewPickleFunction() *PickleFunction {
	r, err := os.Open("pickle.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &PickleFunction{
		image: img,
		cache: map[int][2]float64{},
	}
}

func (p *PickleFunction) RadiusAt(y float64) float64 {
	min, max := p.minMaxAt(y)
	return (max - min) / 2
}

func (p *PickleFunction) CenterAt(y float64) float64 {
	min, max := p.minMaxAt(y)
	return (min + max) / 2
}

func (p *PickleFunction) minMaxAt(y float64) (float64, float64) {
	scale := float64(p.image.Bounds().Dy()) / PickleLength
	idx := p.image.Bounds().Dy() - (int(math.Round(y*scale)) + 1)
	if val, ok := p.cache[idx]; ok {
		return val[0], val[1]
	}

	min := 0
	max := 0
	for x := 0; x < p.image.Bounds().Dx(); x++ {
		_, _, _, alpha := p.image.At(x, idx).RGBA()
		if alpha > 0xffff/2 {
			if min == 0 {
				min = x
			}
			max = x
		}
	}

	p.cache[idx] = [2]float64{
		float64(min) / scale,
		float64(max) / scale,
	}

	return p.minMaxAt(y)
}

type Inscription struct {
	image image.Image
}

func NewInscription() *Inscription {
	r, err := os.Open("inscription.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &Inscription{image: img}
}

func (i *Inscription) ColorAt(t *model3d.Triangle) [3]float64 {
	c := t[0].Add(t[1]).Add(t[2]).Scale(1.0 / 3.0)
	if c.Z < 0 {
		return Green
	}
	scale := float64(i.image.Bounds().Dy()) / PickleLength
	x := int(math.Round(c.X * scale))
	y := i.image.Bounds().Dy() - (int(math.Round(c.Y*scale)) + 1)
	r, g, b, a := i.image.At(x, y).RGBA()
	if a < 0xffff/2 {
		return Green
	}
	return [3]float64{float64(r) / 0xffff, float64(g) / 0xffff, float64(b) / 0xffff}
}
