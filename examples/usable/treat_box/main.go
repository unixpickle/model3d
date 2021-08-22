package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
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
	bone := ReadBone()
	CreateBox(bone)
	CreateLid(bone)
	CreateHandle()
}

func ReadBone() model2d.Collider {
	bmp := model2d.MustReadBitmap("bone.png", nil)
	scale := BoxWidth / float64(bmp.Width)
	boneMesh := bmp.Mesh().Smooth(20).MapCoords(func(c model2d.Coord) model2d.Coord {
		return c.Scale(scale)
	})
	return model2d.MeshToCollider(boneMesh)
}

func CreateBox(bone model2d.Collider) {
	ax := &toolbox3d.AxisSqueeze{
		Axis:  toolbox3d.AxisZ,
		Min:   Thickness,
		Max:   BoxHeight * 0.9,
		Ratio: 0.1,
	}
	box := NewBoxSolid(bone)

	log.Println("Creating box mesh...")
	mesh := model3d.MarchingCubesConj(box, 0.01, 8, ax)

	log.Println("Saving box mesh...")
	mesh.SaveGroupedSTL("box.stl")

	log.Println("Rendering box...")
	render3d.SaveRandomGrid("rendering_box.png", mesh, 3, 3, 300, nil)
}

func CreateLid(bone model2d.Collider) {
	mid := bone.Min().Mid(bone.Max())
	center := model3d.Coord3D{X: mid.X, Y: mid.Y}
	lid := &model3d.SubtractedSolid{
		Positive: NewLidSolid(bone),
		Negative: &toolbox3d.ScrewSolid{
			P1:         center,
			P2:         center.Add(model3d.Coord3D{Z: Thickness * 2}),
			Radius:     ScrewRadius,
			GrooveSize: ScrewGroove,
		},
	}

	log.Println("Creating lid mesh...")
	mesh := model3d.MarchingCubesSearch(lid, 0.0075, 8)

	log.Println("Saving lid mesh...")
	mesh.SaveGroupedSTL("lid.stl")

	log.Println("Rendering lid...")
	render3d.SaveRandomGrid("rendering_lid.png", mesh, 3, 3, 300, nil)
}

func CreateHandle() {
	handle := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Z(HandleHeight),
			Radius: HandleRadius,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Z(HandleHeight),
			P2:         model3d.Coord3D{Z: HandleHeight + Thickness*2},
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGroove,
		},
	}

	log.Println("Creating handle mesh...")
	mesh := model3d.MarchingCubesSearch(handle, 0.005, 8)

	log.Println("Saving handle mesh...")
	mesh.SaveGroupedSTL("handle.stl")

	log.Println("Rendering handle...")
	render3d.SaveRandomGrid("rendering_handle.png", mesh, 3, 3, 300, nil)
}

type BoxSolid struct {
	Outside model2d.Solid
	Inside  model2d.Solid
}

func NewBoxSolid(bone model2d.Collider) *BoxSolid {
	return &BoxSolid{
		Outside: model2d.NewColliderSolid(bone),
		Inside:  model2d.NewColliderSolidInset(bone, Thickness),
	}
}

func (b *BoxSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (b *BoxSolid) Max() model3d.Coord3D {
	return model3d.XYZ(b.Outside.Max().X, b.Outside.Max().Y, BoxHeight)
}

func (b *BoxSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	c2 := c.XY()
	return b.Outside.Contains(c2) && (c.Z <= Thickness || !b.Inside.Contains(c2))
}

type LidSolid struct {
	Outside model2d.Solid
	Inside  model2d.Solid
}

func NewLidSolid(bone model2d.Collider) *LidSolid {
	return &LidSolid{
		Outside: model2d.NewColliderSolid(bone),
		Inside:  model2d.NewColliderSolidInset(bone, Thickness+LidSlack),
	}
}

func (l *LidSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (l *LidSolid) Max() model3d.Coord3D {
	return model3d.XYZ(l.Outside.Max().X, l.Outside.Max().Y, Thickness*2)
}

func (l *LidSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(l, c) {
		return false
	}
	c2 := c.XY()
	return l.Inside.Contains(c2) || (c.Z < Thickness && l.Outside.Contains(c2))
}
