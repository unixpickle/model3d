package main

import (
	"log"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Thickness          = 0.2
	Width              = 5.0
	Height             = 1.0
	MarkSize           = 0.03
	MarkSmallestHeight = 0.05
	MarkGap            = 1.0 / 16.0

	// Epsilon to prevent zero-area triangles on the
	// bottom (flattened) face.
	MarkEpsilon = 0.005
)

func main() {
	x := -1.0 / 16.0
	midY := Height / 2

	mesh := model3d.NewMesh()

	for i := 0; i <= int(Width*16); i++ {
		newX := x + MarkGap - MarkSize/2
		newMidY := MarkSmallestHeight
		for j := uint(0); j <= 4; j++ {
			if i&((1<<j)-1) == 0 {
				newMidY *= 1.5
			}
		}
		newMidY = Height - newMidY

		CreateQuad(mesh, FlatPoint(x, 0), FlatPoint(x, midY), FlatPoint(newX, newMidY),
			FlatPoint(newX, 0))
		CreateQuad(mesh, FlatPoint(x, midY), FlatPoint(x, Height), FlatPoint(newX, Height),
			FlatPoint(newX, newMidY))

		midY = newMidY
		x = newX

		raiseX1 := x
		raiseX2 := x + MarkEpsilon
		raiseX3 := x + MarkSize - MarkEpsilon
		raiseX4 := x + MarkSize

		raiseY1 := midY
		raiseY2 := midY + MarkEpsilon
		raiseY3 := Height - MarkEpsilon
		raiseY4 := Height

		// Left and right sides of marking.
		CreateQuad(mesh, FlatPoint(raiseX1, raiseY4), FlatPoint(raiseX1, raiseY1),
			RaisedPoint(raiseX2, raiseY2), RaisedPoint(raiseX2, raiseY3))
		CreateQuad(mesh, FlatPoint(raiseX4, raiseY1), FlatPoint(raiseX4, raiseY4),
			RaisedPoint(raiseX3, raiseY3), RaisedPoint(raiseX3, raiseY2))

		// Top and bottom sides of marking.
		CreateQuad(mesh, FlatPoint(raiseX4, raiseY4), FlatPoint(raiseX1, raiseY4),
			RaisedPoint(raiseX2, raiseY3), RaisedPoint(raiseX3, raiseY3))
		CreateQuad(mesh, FlatPoint(raiseX1, raiseY1), FlatPoint(raiseX4, raiseY1),
			RaisedPoint(raiseX3, raiseY2), RaisedPoint(raiseX2, raiseY2))

		// Top face of marking.
		CreateQuad(mesh, RaisedPoint(raiseX2, raiseY2), RaisedPoint(raiseX2, raiseY3),
			RaisedPoint(raiseX3, raiseY3), RaisedPoint(raiseX3, raiseY2))

		// Flat part of ruler below marking.
		CreateQuad(mesh, FlatPoint(raiseX1, 0), FlatPoint(raiseX1, raiseY1),
			FlatPoint(raiseX4, raiseY1), FlatPoint(raiseX4, 0))

		x = raiseX4
	}

	mesh.Iterate(func(t *model3d.Triangle) {
		// Create sides of ruler.
		for i := 0; i < 3; i++ {
			p := t[i]
			p1 := t[(i+1)%3]
			if len(mesh.Find(p, p1)) == 1 {
				p2 := p1
				p3 := p
				p2.Z = -Thickness
				p3.Z = -Thickness
				CreateQuad(mesh, p, p1, p2, p3)
			}
		}

		// Create bottom face.
		t1 := *t
		for i := range t1 {
			t1[i].Z = -Thickness
		}
		mesh.Add(&t1)
	})

	// I was too lazy to figure out the right normals, so
	// we automatically calculate them so that they face
	// towards the outside of the object.
	log.Println("Calculating normals...")
	mesh, _ = mesh.RepairNormals(1e-8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("ruler.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func FlatPoint(x, y float64) model3d.Coord3D {
	return model3d.Coord3D{X: x, Y: y}
}

func RaisedPoint(x, y float64) model3d.Coord3D {
	return model3d.Coord3D{X: x, Y: y, Z: MarkSize}
}

func CreateQuad(m *model3d.Mesh, p1, p2, p3, p4 model3d.Coord3D) {
	m.Add(&model3d.Triangle{p1, p2, p3})
	m.Add(&model3d.Triangle{p1, p3, p4})
}
