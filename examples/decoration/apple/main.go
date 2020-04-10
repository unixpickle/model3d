package main

import (
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const AppleHeight = 1.0

func main() {
	bite := &model3d.Torus{
		Center:      model3d.Coord3D{X: -1.7, Y: 0, Z: AppleHeight / 2},
		Axis:        model3d.Coord3D{Z: 1},
		OuterRadius: 1.0,
		InnerRadius: 0.3,
	}
	biggerBite := *bite
	biggerBite.InnerRadius += 0.01

	stem := &model3d.Cylinder{
		P1:     model3d.Coord3D{Z: AppleHeight / 2},
		P2:     model3d.Coord3D{Z: AppleHeight * 1.1},
		Radius: AppleHeight / 30,
	}
	biggerStem := *stem
	biggerStem.Radius += 0.005
	biggerStem.P2.Z += 0.005

	solid := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			NewAppleSolid(),
			stem,
		},
		Negative: bite,
	}

	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8).SmoothAreas(0.1, 5)

	sig := NewSignature()
	vertexColor := func(c model3d.Coord3D) [3]float64 {
		if biggerBite.Contains(c) {
			if sig.Contains(c) {
				return [3]float64{0, 0, 0}
			}
			return [3]float64{1, 1, 0.5}
		} else if biggerStem.Contains(c) {
			return [3]float64{0.27, 0.21, 0}
		} else {
			return [3]float64{1, 0, 0}
		}
	}
	colorFunc := model3d.VertexColorsToTriangle(vertexColor)

	ioutil.WriteFile("apple.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200,
		render3d.TriangleColorFunc(colorFunc))
}

type AppleSolid struct {
	Image image.Image
}

func NewAppleSolid() *AppleSolid {
	r, err := os.Open("half_apple.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &AppleSolid{Image: img}
}

func (a *AppleSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -a.width(), Y: -a.width(), Z: 0}
}

func (a *AppleSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: a.width(), Y: a.width(), Z: AppleHeight}
}

func (a *AppleSolid) Contains(c model3d.Coord3D) bool {
	min := a.Min()
	max := a.Max()
	if c.Min(min) != min || c.Max(max) != max {
		return false
	}

	radius := (model3d.Coord2D{X: c.X, Y: c.Y}).Dist(model3d.Coord2D{})
	imageX := int(math.Round(float64(a.Image.Bounds().Dx()) * (1 - radius/a.width())))
	imageY := int(math.Round((AppleHeight - c.Z) / a.scale()))

	_, _, _, alpha := a.Image.At(imageX, imageY).RGBA()
	return alpha > 0xffff/2
}

func (a *AppleSolid) scale() float64 {
	return AppleHeight / float64(a.Image.Bounds().Dy())
}

func (a *AppleSolid) width() float64 {
	return AppleHeight * float64(a.Image.Bounds().Dx()) / float64(a.Image.Bounds().Dy())
}

type Signature struct {
	Image image.Image
}

func NewSignature() *Signature {
	r, err := os.Open("turing_signature.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &Signature{Image: img}
}

func (s *Signature) Contains(c model3d.Coord3D) bool {
	if c.X > 0 {
		return false
	}
	scale := float64(s.Image.Bounds().Dx()) / 0.5
	imageX := s.Image.Bounds().Dx()/2 - int(math.Round(c.Y*scale))
	imageY := s.Image.Bounds().Dy()/2 - int(math.Round((c.Z-AppleHeight/2-0.05)*scale))
	if imageX < 0 || imageY < 0 || imageX >= s.Image.Bounds().Dx() ||
		imageY >= s.Image.Bounds().Dy() {
		return false
	}
	_, _, _, a := s.Image.At(imageX, imageY).RGBA()
	return a > 0xffff/2
}
