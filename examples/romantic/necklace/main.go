package main

import (
	"image"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	LinkWidth     = 0.5
	LinkHeight    = 0.8
	LinkThickness = 0.06
	LinkOddShift  = LinkWidth * 0.3

	HookOffset = 0.2
	HookLength = LinkHeight

	TotalLength = 20

	StartRadius = 3.0
	SpiralRate  = 0.4 / (math.Pi * 2)
	MoveRate    = 0.6 * LinkHeight
)

func main() {
	smoother := model3d.MeshSmoother{
		StepSize:           0.1,
		Iterations:         20,
		ConstraintDistance: 0.0025,
		ConstraintWeight:   0.3,
	}

	log.Println("Creating link mesh...")
	solid := LinkSolid{}
	link := model3d.SolidToMesh(solid, 0.005, 0, 0, 0)
	link = smoother.Smooth(link)
	link = link.FlattenBase(0)
	link = link.EliminateCoplanar(1e-8)
	VerifyMesh(link)

	log.Println("Creating hook mesh...")
	hook := model3d.SolidToMesh(HookSolid{}, 0.005, 0, 0, 0)
	hook = smoother.Smooth(hook)
	hook = hook.FlattenBase(0)
	hook = hook.EliminateCoplanar(1e-8)
	VerifyMesh(hook)

	log.Println("Creating full mesh...")
	m := model3d.NewMesh()
	manifold := NewSpiralManifold(StartRadius, SpiralRate)
	numLinks := int(TotalLength / LinkHeight)
	for i := 0; i < numLinks; i++ {
		offset := model3d.Coord3D{X: LinkOddShift / 2}
		if i%2 == 1 {
			offset = offset.Scale(-1)
		}
		subMesh := link
		if i == numLinks-1 {
			subMesh = hook
		}
		m.AddMesh(subMesh.MapCoords(offset.Add).MapCoords(manifold.Convert))
		manifold.Move(MoveRate)
	}

	log.Println("Saving mesh...")
	m.SaveGroupedSTL("necklace.stl")

	log.Println("Saving rendering...")
	img := image.NewGray(image.Rect(0, 0, 1000, 500))
	model3d.RenderRayCast(model3d.MeshToCollider(m), img, model3d.Coord3D{Z: 2, Y: -4},
		model3d.Coord3D{X: 1}, model3d.Coord3D{Y: -0.4, Z: -0.5},
		model3d.Coord3D{Y: 0.5, Z: -0.4}, math.Pi/2)
	f, err := os.Create("rendering.png")
	essentials.Must(err)
	defer f.Close()
	png.Encode(f, img)

	log.Println("Verifying mesh...")
	VerifyMesh(m)
}

func VerifyMesh(m *model3d.Mesh) {
	if m.SelfIntersections() > 0 {
		panic("self intersections detected")
	}
	if _, n := m.RepairNormals(1e-5); n != 0 {
		panic("incorrect normals")
	}
}

type LinkSolid struct{}

func (l LinkSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -LinkWidth / 2, Y: -LinkHeight / 2, Z: 0}
}

func (l LinkSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: LinkWidth / 2, Y: LinkHeight / 2,
		Z: LinkWidth/2 + LinkThickness*math.Sqrt2}
}

func (l LinkSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(l, c) {
		return false
	}
	if c.Z < LinkThickness &&
		(c.X < -LinkWidth/2+LinkThickness || c.X > LinkWidth/2-LinkThickness) {
		return true
	}
	if c.Y > -LinkHeight/2+LinkThickness && c.Y < LinkHeight/2-LinkThickness {
		return false
	}

	height := LinkWidth/2 - math.Abs(c.X)
	return c.Z >= height && c.Z <= height+LinkThickness*math.Sqrt2
}

type HookSolid struct{}

func (h HookSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -HookLength / 2, Y: -LinkHeight / 2, Z: 0}
}

func (h HookSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: HookLength / 2, Y: LinkHeight/2 + HookOffset + LinkThickness,
		Z: LinkWidth/2 + LinkThickness*math.Sqrt2}
}

func (h HookSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(h, c) {
		return false
	}
	if c.Z < LinkThickness {
		if c.Y < LinkHeight/2 {
			if math.Abs(c.X) > LinkWidth/2-LinkThickness && math.Abs(c.X) < LinkWidth/2 {
				return true
			}
			if c.Y > LinkHeight/2-LinkThickness && math.Abs(c.X) < LinkWidth/2 {
				return true
			}
		} else if c.Y < LinkHeight/2+HookOffset {
			return math.Abs(c.X) < LinkThickness/2
		} else {
			// Hook itself fills the max-y part of the solid.
			return true
		}
	}
	if c.Y < -LinkHeight/2+LinkThickness && math.Abs(c.X) < LinkWidth/2 {
		height := LinkWidth/2 - math.Abs(c.X)
		if c.Z >= height && c.Z <= height+LinkThickness*math.Sqrt2 {
			return true
		}
	}

	return false
}
