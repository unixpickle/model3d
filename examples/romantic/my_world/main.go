package main

import (
	"image"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	GlobeRadius = 0.7
	GlobeOutset = 0.03

	BoardSize      = 3.0
	BoardThickness = 0.2
	BoardInset     = 0.1
)

func main() {
	solid := model3d.JoinedSolid{NewBaseSolid(), NewGlobeSolid()}
	mesh := model3d.SolidToMesh(solid, 0.01, 1, -1, 5)
	ioutil.WriteFile("my_world.stl", mesh.EncodeSTL(), 0755)
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}

type GlobeSolid struct {
	Equirect *toolbox3d.Equirect
}

func NewGlobeSolid() *GlobeSolid {
	r, err := os.Open("map.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &GlobeSolid{Equirect: toolbox3d.NewEquirect(img)}
}

func (g *GlobeSolid) Min() model3d.Coord3D {
	d := GlobeRadius + GlobeOutset
	return model3d.Coord3D{X: -d, Y: -d, Z: 0}
}

func (g *GlobeSolid) Max() model3d.Coord3D {
	d := GlobeRadius + GlobeOutset
	return model3d.Coord3D{X: d, Y: d, Z: d}
}

func (g *GlobeSolid) Contains(c model3d.Coord3D) bool {
	if c.Z < 0 {
		return false
	}
	c.X, c.Y, c.Z = c.Y, c.Z, c.X
	geo := c.Geo()
	r, _, _, _ := g.Equirect.At(geo).RGBA()
	if r > 0xffff/2 {
		return c.Norm() < GlobeRadius+GlobeOutset
	} else {
		return c.Norm() < GlobeRadius
	}
}

type BaseSolid struct {
	Img image.Image
}

func NewBaseSolid() *BaseSolid {
	r, err := os.Open("base.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &BaseSolid{Img: img}
}

func (b *BaseSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -BoardSize / 2, Y: -BoardSize / 2, Z: -BoardThickness}
}

func (b *BaseSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BoardSize / 2, Y: BoardSize / 2, Z: 0}
}

func (b *BaseSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(b.Min()) != b.Min() || c.Max(b.Max()) != b.Max() {
		return false
	}
	if c.Z < -BoardInset {
		return true
	}

	x := int(float64(b.Img.Bounds().Dx()) * (c.X/BoardSize + 0.5))
	y := int(float64(b.Img.Bounds().Dy()) * (0.5 - c.Y/BoardSize))
	r, _, _, _ := b.Img.At(x, y).RGBA()
	return r > 0xffff/2
}
