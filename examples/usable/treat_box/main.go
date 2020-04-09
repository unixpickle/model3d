package main

import (
	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/model3d/toolbox3d"

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
	mesh := model3d.MarchingCubesSearch(box, 0.01, 8)
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
	mesh = model3d.MarchingCubesSearch(lid, 0.0075, 8)
	mesh.SaveGroupedSTL("lid.stl")
	model3d.SaveRandomGrid("rendering_lid.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	handle := model3d.JoinedSolid{
		&model3d.Cylinder{
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
	mesh = model3d.MarchingCubesSearch(handle, 0.005, 8)
	mesh.SaveGroupedSTL("handle.stl")
	model3d.SaveRandomGrid("rendering_handle.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

}

type BoxSolid struct {
	Bitmap   *model2d.Bitmap
	Collider model2d.Collider
}

func NewBoxSolid() *BoxSolid {
	bmp := model2d.MustReadBitmap("bone.png", nil)
	collider := model2d.MeshToCollider(bmp.Mesh())
	return &BoxSolid{Bitmap: bmp, Collider: collider}
}

func (b *BoxSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (b *BoxSolid) Max() model3d.Coord3D {
	depth := BoxWidth * float64(b.Bitmap.Height) / float64(b.Bitmap.Width)
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
	x := int(float64(b.Bitmap.Width) * c.X / b.Max().X)
	y := int(float64(b.Bitmap.Height) * c.Y / b.Max().Y)
	return b.Bitmap.Get(x, y)
}

func (b *BoxSolid) containsInset(c model3d.Coord3D, thickness float64) bool {
	if !b.filledContains(c) {
		return false
	}
	scale := float64(b.Bitmap.Width) / b.Max().X
	x := float64(b.Bitmap.Width) * c.X / b.Max().X
	y := float64(b.Bitmap.Height) * c.Y / b.Max().Y
	return !b.Collider.CircleCollision(model2d.Coord{X: x, Y: y}, thickness*scale)
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
