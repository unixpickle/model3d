package main

import (
	"flag"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var outFile string
	var renderFile string
	var minHeight float64
	var maxHeight float64
	var radius float64
	var template string

	flag.StringVar(&outFile, "out", "coin.stl", "output file name")
	flag.StringVar(&renderFile, "render", "rendering.png", "rendered output file name")
	flag.Float64Var(&minHeight, "min-height", 0.1, "minimum height")
	flag.Float64Var(&maxHeight, "max-height", 0.13, "maximum height")
	flag.Float64Var(&radius, "radius", 0.5, "radius of coin")
	flag.StringVar(&template, "template", "example.png", "coin depth map image")

	flag.Parse()

	f, err := os.Open(template)
	essentials.Must(err)
	templateImg, _, err := image.Decode(f)
	f.Close()
	essentials.Must(err)

	solid := &CoinSolid{
		Img:       templateImg,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
		Radius:    radius,
	}

	m := model3d.MarchingCubesSearch(solid, radius/200, 8)

	essentials.Must(ioutil.WriteFile(outFile, m.EncodeSTL(), 0755))
	essentials.Must(render3d.SaveRandomGrid(renderFile, m, 4, 4, 200, nil))
}

type CoinSolid struct {
	Img       image.Image
	MinHeight float64
	MaxHeight float64
	Radius    float64
}

func (c *CoinSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -c.Radius, Y: -c.Radius}
}

func (c *CoinSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: c.Radius, Y: c.Radius, Z: c.MaxHeight}
}

func (c *CoinSolid) Contains(coord model3d.Coord3D) bool {
	theta := math.Atan2(coord.Y, coord.X)
	radius := (model3d.Coord2D{X: coord.X, Y: coord.Y}).Norm() / c.Radius
	if radius > 1 {
		return false
	}
	return coord.Z < c.height(theta, radius) && coord.Z >= 0
}

func (c *CoinSolid) height(theta, radius float64) float64 {
	x := math.Round((math.Cos(theta)*radius + 1) * float64(c.Img.Bounds().Dx()) / 2)
	y := math.Round((1 - math.Sin(theta)*radius) * float64(c.Img.Bounds().Dy()) / 2)
	rawColor := c.Img.At(int(x), int(y))
	r, g, b, _ := color.RGBAModel.Convert(rawColor).RGBA()
	relHeight := 1 - float64(r+g+b)/(3*0xffff)
	return c.MinHeight + (c.MaxHeight-c.MinHeight)*relHeight
}
