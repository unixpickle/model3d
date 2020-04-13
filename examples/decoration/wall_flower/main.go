package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Radius        = 1.0
	LengthRadians = 1.0
	Thickness     = 0.1
	BaseThickness = 0.1
	BaseRadius    = 0.5
)

func main() {
	log.Println("Wrapping flower shape...")
	sphereMesh := WrappedFlowerShape()

	log.Println("Creating solid...")
	solid := model3d.JoinedSolid{
		model3d.NewColliderSolidHollow(model3d.MeshToCollider(sphereMesh), Thickness),
		BaseSolid{},
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("flower.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func WrappedFlowerShape() *model3d.Mesh {
	shape := NewFlowerShape()
	mesh := model3d.NewMeshPolar(func(g model3d.GeoCoord) float64 {
		if shape.Get(g.Lat, g.Lon) {
			return Radius
		} else {
			return 0
		}
	}, 2000)
	mesh.Iterate(func(t *model3d.Triangle) {
		if t[0].Norm() == 0 || t[1].Norm() == 0 || t[2].Norm() == 0 {
			mesh.Remove(t)
		}
	})
	mesh = mesh.MapCoords(func(f model3d.Coord3D) model3d.Coord3D {
		f.Z *= -1
		f.Z += Radius
		return f
	})
	return mesh
}

type FlowerShape struct {
	Solid model2d.Solid
}

func NewFlowerShape() *FlowerShape {
	bitmap := model2d.MustReadBitmap("shape.png", nil)
	mesh := bitmap.Mesh().SmoothSq(200)
	mesh = mesh.MapCoords(model2d.Coord{
		X: 1 / float64(bitmap.Width),
		Y: 1 / float64(bitmap.Height),
	}.Mul)
	collider := model2d.MeshToCollider(mesh)
	solid := model2d.NewColliderSolid(collider)
	return &FlowerShape{
		Solid: solid,
	}
}

func (f *FlowerShape) Get(x, y float64) bool {
	return f.Solid.Contains(model2d.Coord{
		X: (x + LengthRadians) / (LengthRadians * 2),
		Y: (y + LengthRadians) / (LengthRadians * 2),
	})
}

type BaseSolid struct{}

func (b BaseSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -BaseRadius, Y: -BaseRadius, Z: -BaseThickness}
}

func (b BaseSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BaseRadius, Y: BaseRadius, Z: Radius}
}

func (b BaseSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	cylinderDist := c.Coord2D().Norm()
	sphereDist := c.Dist(model3d.Coord3D{Z: Radius})
	return cylinderDist < BaseRadius && sphereDist >= Radius
}
