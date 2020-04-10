package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Height = 4.0
	Radius = 0.3

	BaseWidth = 3.0
	BaseDepth = 2.0
)

func main() {
	log.Println("Creating solid...")
	solid := model3d.JoinedSolid{
		&model3d.Rect{
			MinVal: model3d.Coord3D{X: -BaseWidth / 2, Y: -BaseDepth / 2, Z: -Radius},
			MaxVal: model3d.Coord3D{X: BaseWidth / 2, Y: BaseDepth / 2, Z: Radius},
		},
		&TubeSolid{
			Curve: model2d.MeshToCollider(Mesh2D(PolyPiecewiseCurve)),
		},
		&TubeSolid{
			Curve: model2d.MeshToCollider(Mesh2D(SinusoidalCurve)),
		},
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)

	log.Println("Saving results...")
	mesh.SaveGroupedSTL("curvy_thing.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type TubeSolid struct {
	Curve model2d.Collider
}

func (t TubeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: t.Curve.Min().Y - Radius, Y: -Radius, Z: t.Curve.Min().X - Radius}
}

func (t TubeSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: t.Curve.Max().Y + Radius, Y: Radius, Z: t.Curve.Max().X}
}

func (t TubeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(t, c) {
		return false
	}
	c2d := model2d.Coord{X: c.Z, Y: c.X}
	if math.Abs(c.Y) > Radius {
		return false
	}
	radius := math.Sqrt(Radius*Radius - c.Y*c.Y)
	return t.Curve.CircleCollision(c2d, radius)
}

func Mesh2D(f func(float64) float64) *model2d.Mesh {
	res := model2d.NewMesh()
	for z := 0.0; z+0.01 < Height; z += 0.01 {
		p1 := model2d.Coord{X: z, Y: f(z)}
		p2 := model2d.Coord{X: z + 0.01, Y: f(z + 0.01)}
		res.Add(&model2d.Segment{p1, p2})
	}
	return res
}

func SinusoidalCurve(z float64) float64 {
	return -0.3*math.Sin(8-2*z)*math.Sqrt(z) + 0.6
}

func PolyPiecewiseCurve(z float64) float64 {
	if z < 2 {
		return 0.5*math.Pow(z, 2)*(z-2) - 0.4
	} else {
		return -PolyPiecewiseCurve(4-z) - 0.8
	}
}
