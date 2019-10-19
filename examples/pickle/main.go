package main

import (
	"image"
	"image/color"
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

const Color = false

func main() {
	var solid model3d.Solid
	solid = &PickleSolid{F: NewPickleFunction()}
	inscription := NewInscription()

	if !Color {
		solid = &model3d.SubtractedSolid{
			Positive: solid,
			Negative: inscription,
		}
	}

	mesh := model3d.SolidToMesh(solid, 0.025, 2, 0.8, 8)

	if !Color {
		ioutil.WriteFile("pickle.stl", mesh.EncodeSTL(), 0755)
	} else {
		colorFunc := model3d.VertexColorsToTriangle(inscription.ColorAt)
		ioutil.WriteFile("pickle.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
	}
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

	// Perform linear interpolation between two y values.
	idx1 := p.image.Bounds().Dy() - (int(math.Floor(y*scale)) + 1)
	idx2 := p.image.Bounds().Dy() - (int(math.Ceil(y*scale)) + 1)
	frac1 := math.Ceil(y*scale) - y*scale
	frac2 := 1 - frac1
	min1, max1 := p.getCache(idx1)
	min2, max2 := p.getCache(idx2)
	return min1*frac1 + min2*frac2, max1*frac1 + max2*frac2
}

func (p *PickleFunction) getCache(idx int) (float64, float64) {
	if idx < 0 || idx >= p.image.Bounds().Dy() {
		return 0, 0
	}
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

	scale := float64(p.image.Bounds().Dy()) / PickleLength
	p.cache[idx] = [2]float64{
		float64(min) / scale,
		float64(max) / scale,
	}

	return p.cache[idx][0], p.cache[idx][1]
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

func (i *Inscription) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -PickleLength, Y: -PickleLength, Z: -PickleLength}
}

func (i *Inscription) Max() model3d.Coord3D {
	return model3d.Coord3D{X: PickleLength, Y: PickleLength, Z: PickleLength}
}

func (i *Inscription) Contains(c model3d.Coord3D) bool {
	if i.Max().Max(c) != i.Max() || i.Min().Min(c) != i.Min() {
		return false
	}
	_, _, _, a := i.projectedColor(c).RGBA()
	return a >= 0xffff/2
}

func (i *Inscription) ColorAt(c model3d.Coord3D) [3]float64 {
	if c.Z < 0 {
		return Green
	}
	r, g, b, a := i.projectedColor(c).RGBA()
	if a < 0xffff/2 {
		return Green
	}
	return [3]float64{float64(r) / 0xffff, float64(g) / 0xffff, float64(b) / 0xffff}
}

func (i *Inscription) projectedColor(c model3d.Coord3D) color.Color {
	scale := float64(i.image.Bounds().Dy()) / PickleLength
	x := int(math.Round(c.X * scale))
	y := i.image.Bounds().Dy() - (int(math.Round(c.Y*scale)) + 1)
	return i.image.At(x, y)
}
