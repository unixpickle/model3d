package main

import (
	"image/color"
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateLetters() render3d.Object {
	log.Println("Creating letters...")
	defer log.Println("Done creating letters.")

	bmp := model2d.MustReadBitmap("assets/letters.png", func(c color.Color) bool {
		r, _, _, _ := c.RGBA()
		return r < 0x5000
	}).FlipY()
	mesh2d := bmp.Mesh().SmoothSq(50)
	scale := LettersWidth / (mesh2d.Max().X - mesh2d.Min().X)
	subX := (mesh2d.Max().X + mesh2d.Min().X) / 2
	mesh2d = mesh2d.MapCoords(func(c model2d.Coord) model2d.Coord {
		c.X -= subX
		return c.Scale(scale)
	})
	solid := &LettersSolid{Collider: model2d.MeshToCollider(mesh2d)}
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisY,
		Min:   LettersRounding + 0.01,
		Max:   LettersThickness - (LettersRounding + 0.01),
		Ratio: 0.01,
	}
	mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, solid), 0.01, 8)
	mesh = mesh.MapCoords(ax.Inverse().Apply)
	mesh = mesh.EliminateCoplanar(1e-8)

	return &render3d.ColliderObject{
		Collider: model3d.MeshToCollider(mesh),
		Material: &render3d.PhongMaterial{
			Alpha:         20,
			SpecularColor: render3d.NewColor(0.1),
			DiffuseColor:  render3d.NewColorRGB(0.48, 0, 0.56),
		},
	}
}

type LettersSolid struct {
	Collider model2d.Collider
}

func (l *LettersSolid) Min() model3d.Coord3D {
	min2 := l.Collider.Min()
	return model3d.Coord3D{X: min2.X, Y: 0, Z: min2.Y}
}

func (l *LettersSolid) Max() model3d.Coord3D {
	max2 := l.Collider.Max()
	return model3d.Coord3D{X: max2.X, Y: LettersThickness, Z: max2.Y}
}

func (l *LettersSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(l, c) {
		return false
	}
	edgeDist := math.Min(c.Y, l.Max().Y-c.Y)
	radius := 0.0
	if edgeDist < LettersRounding {
		frac := (LettersRounding - edgeDist) / LettersRounding
		radius = LettersRounding * (1 - math.Sqrt(1-frac*frac))
	}
	c2d := model2d.Coord{X: c.X, Y: c.Z}
	return model2d.ColliderContains(l.Collider, c2d, radius)
}
