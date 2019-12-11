package main

import (
	"image"
	"image/png"
	"os"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	BoxWidth  = 4.0
	BoxHeight = 1.4
	Thickness = 0.2
	LidSlack  = 0.05

	ScrewRadius = 0.14
	ScrewGroove = 0.04
	ScrewSlack  = 0.03

	HandleRadius = 0.3
	HandleHeight = 0.6
)

func main() {
	box := NewBoxSolid()
	mesh := model3d.SolidToMesh(box, 0.01, 0, -1, 20)
	mesh.SaveGroupedSTL("box.stl")
	model3d.SaveRandomGrid("rendering_box.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	center := box.Min().Mid(box.Max())
	center.Z = 0
	lid := &model3d.SubtractedSolid{
		Positive: &LidSolid{BoxSolid: box},
		Negative: &toolbox3d.ScrewSolid{
			P1:         center,
			P2:         center.Add(model3d.Coord3D{Z: Thickness * 2}),
			Radius:     ScrewRadius,
			GrooveSize: ScrewGroove,
		},
	}
	mesh = model3d.SolidToMesh(lid, 0.0075, 0, -1, 20)
	mesh.SaveGroupedSTL("lid.stl")
	model3d.SaveRandomGrid("rendering_lid.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	handle := model3d.JoinedSolid{
		&model3d.CylinderSolid{
			P1:     model3d.Coord3D{},
			P2:     model3d.Coord3D{Z: HandleHeight},
			Radius: HandleRadius,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: HandleHeight},
			P2:         model3d.Coord3D{Z: HandleHeight + Thickness*2},
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGroove,
		},
	}
	mesh = model3d.SolidToMesh(handle, 0.005, 0, -1, 10)
	mesh.SaveGroupedSTL("handle.stl")
	model3d.SaveRandomGrid("rendering_handle.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

}

type BoxSolid struct {
	Img image.Image
}

func NewBoxSolid() *BoxSolid {
	f, err := os.Open("bone.png")
	essentials.Must(err)
	defer f.Close()
	img, err := png.Decode(f)
	essentials.Must(err)
	return &BoxSolid{Img: img}
}

func (b *BoxSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (b *BoxSolid) Max() model3d.Coord3D {
	depth := BoxWidth * float64(b.Img.Bounds().Dy()) / float64(b.Img.Bounds().Dx())
	return model3d.Coord3D{X: BoxWidth, Y: depth, Z: BoxHeight}
}

func (b *BoxSolid) Contains(c model3d.Coord3D) bool {
	if !b.filledContains(c) {
		return false
	}
	if c.Z <= Thickness {
		return true
	}
	return !b.containsInset(c, Thickness)
}

func (b *BoxSolid) filledContains(c model3d.Coord3D) bool {
	if c.Min(b.Min()) != b.Min() || c.Max(b.Max()) != b.Max() {
		return false
	}
	x := int(float64(b.Img.Bounds().Dx()) * c.X / b.Max().X)
	y := int(float64(b.Img.Bounds().Dy()) * c.Y / b.Max().Y)
	red, _, blue, _ := b.Img.At(x, y).RGBA()
	return blue > red*2
}

func (b *BoxSolid) containsInset(c model3d.Coord3D, thickness float64) bool {
	scale := BoxWidth / (BoxWidth - thickness*4)
	center := b.Min().Mid(b.Max())
	if c.X < center.X {
		center.X = b.Min().Scale(0.6).Add(center.Scale(0.4)).X
	} else {
		center.X = b.Max().Scale(0.6).Add(center.Scale(0.4)).X
	}
	outset := c.Sub(center).Scale(scale).Add(center)
	outset.Z = c.Z
	return b.filledContains(outset)
}

type LidSolid struct {
	BoxSolid *BoxSolid
}

func (l *LidSolid) Min() model3d.Coord3D {
	return l.BoxSolid.Min()
}

func (l *LidSolid) Max() model3d.Coord3D {
	res := l.BoxSolid.Max()
	res.Z = Thickness * 2
	return res
}

func (l *LidSolid) Contains(c model3d.Coord3D) bool {
	if c.Z > Thickness*2 {
		return false
	}
	if c.Z <= Thickness {
		return l.BoxSolid.Contains(c)
	}
	return l.BoxSolid.containsInset(c, Thickness+LidSlack)
}
