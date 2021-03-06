package main

import (
	"image"
	"image/png"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
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
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
	log.Println("Eliminating co-planar...")
	mesh = mesh.EliminateCoplanar(1e-8)
	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("my_world.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200, nil)
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
	return model3d.XYZ(-d, -d, 0)
}

func (g *GlobeSolid) Max() model3d.Coord3D {
	d := GlobeRadius + GlobeOutset
	return model3d.XYZ(d, d, d)
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
	return model3d.XYZ(-BoardSize/2, -BoardSize/2, -BoardThickness)
}

func (b *BaseSolid) Max() model3d.Coord3D {
	return model3d.XYZ(BoardSize/2, BoardSize/2, 0)
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
