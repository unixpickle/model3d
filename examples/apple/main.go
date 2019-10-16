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

const AppleHeight = 1.0

func main() {
	r, err := os.Open("half_apple.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)

	bite := &model3d.TorusSolid{
		Center:      model3d.Coord3D{X: -1.5, Y: 0, Z: AppleHeight / 2},
		Axis:        model3d.Coord3D{Z: 1},
		OuterRadius: 1.0,
		InnerRadius: 0.2,
	}
	biggerBite := *bite
	biggerBite.InnerRadius = 0.21

	solid := &model3d.SubtractedSolid{
		Positive: &AppleSolid{Image: img},
		Negative: bite,
	}

	mesh := model3d.SolidToMesh(solid, 0.025, 2, 0.8, 5)
	colorFunc := func(t *model3d.Triangle) [3]float64 {
		if biggerBite.Contains(t[0]) {
			return [3]float64{1, 1, 0.5}
		} else {
			return [3]float64{1, 0, 0}
		}
	}
	ioutil.WriteFile("apple.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)
}

type AppleSolid struct {
	Image image.Image
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
