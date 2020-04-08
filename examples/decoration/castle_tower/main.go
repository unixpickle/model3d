package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d"
)

const (
	BrickZSpace     = 0.4
	BrickThetaSpace = 0.4
	BrickDivot      = 0.03

	TopBlockCount     = 10
	TopBlockThickness = 0.2

	WallHeight      = 3.0
	WallThickness   = 0.4
	WallBrickXSpace = 0.6
)

func main() {
	tower := model3d.StackedSolid{
		model3d.JoinedSolid{
			&BrickCylinder{Height: 3, Radius: 1},
			&toolbox3d.Ramp{
				P1: model3d.Coord3D{Z: 2},
				P2: model3d.Coord3D{Z: 3.2},
				Solid: &model3d.CylinderSolid{
					P1:     model3d.Coord3D{Z: 2},
					P2:     model3d.Coord3D{Z: 3.2},
					Radius: 1.2,
				},
			},
		},
		&BrickCylinder{Height: 1.2, Radius: 1.2},
		&TopBlocks{Height: 0.3, Radius: 1.2},
	}
	solid := model3d.JoinedSolid{
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -3.7, Y: -1.2, Z: -0.2},
			MaxVal: model3d.Coord3D{X: 3.7, Y: 1.2, Z: 0},
		},
		&XBlock{X: -2.5, Solid: tower},
		&XBlock{X: 2.5, Solid: tower},
		&Wall{X1: -2, X2: 2},
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)

	log.Println("Saving STL...")
	mesh.SaveGroupedSTL("castle.stl")

	log.Println("Saving rendering...")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 500, 500)
}

type XBlock struct {
	Solid model3d.Solid
	X     float64
}

func (x *XBlock) Min() model3d.Coord3D {
	res := x.Solid.Min()
	res.X += x.X
	return res
}

func (x *XBlock) Max() model3d.Coord3D {
	res := x.Solid.Max()
	res.X += x.X
	return res
}

func (x *XBlock) Contains(c model3d.Coord3D) bool {
	return x.Solid.Contains(c.Sub(model3d.Coord3D{X: x.X}))
}

type BrickCylinder struct {
	Height float64
	Radius float64
}

func (b *BrickCylinder) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -b.Radius, Y: -b.Radius}
}

func (b *BrickCylinder) Max() model3d.Coord3D {
	return model3d.Coord3D{X: b.Radius, Y: b.Radius, Z: b.Height}
}

func (b *BrickCylinder) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}

	effectiveRadius := b.Radius

	thetaDist := math.Atan2(c.Y, c.X) + math.Pi
	dist1 := math.Mod(c.Z, BrickZSpace) - BrickZSpace/2
	dist2 := b.Radius * (math.Mod(thetaDist, BrickThetaSpace) - BrickThetaSpace/2)

	for _, dist := range []float64{dist1, dist2} {
		effectiveRadius = math.Min(effectiveRadius, b.Radius-BrickDivot+math.Abs(dist))
	}

	return c.Coord2D().Norm() < effectiveRadius
}

type TopBlocks struct {
	Height float64
	Radius float64
}

func (t *TopBlocks) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -t.Radius, Y: -t.Radius}
}

func (t *TopBlocks) Max() model3d.Coord3D {
	return model3d.Coord3D{X: t.Radius, Y: t.Radius, Z: t.Height}
}

func (t *TopBlocks) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(t, c) {
		return false
	}
	r := c.Coord2D().Norm()
	if r > t.Radius || r < t.Radius-TopBlockThickness {
		return false
	}
	thetaDist := math.Atan2(c.Y, c.X) + math.Pi
	spaceTheta := math.Pi * 2 / TopBlockCount
	return math.Mod(thetaDist, spaceTheta) < spaceTheta/2
}

type Wall struct {
	X1 float64
	X2 float64
}

func (w *Wall) Min() model3d.Coord3D {
	return model3d.Coord3D{X: w.X1, Y: -WallThickness / 2}
}

func (w *Wall) Max() model3d.Coord3D {
	return model3d.Coord3D{X: w.X2, Y: WallThickness / 2, Z: WallHeight}
}

func (w *Wall) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(w, c) {
		return false
	}
	dist1 := math.Mod(c.X-w.X1, WallBrickXSpace) - WallBrickXSpace/2
	dist2 := math.Mod(c.Z, BrickZSpace) - BrickZSpace/2
	thickness := WallThickness / 2
	for _, dist := range []float64{dist1, dist2} {
		thickness = math.Min(thickness, thickness-BrickDivot+math.Abs(dist))
	}
	return math.Abs(c.Y) < thickness
}
