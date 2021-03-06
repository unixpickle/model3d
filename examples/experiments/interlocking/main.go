package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	Size     = 0.1
	Slope    = 0.7
	Offset   = 0.4
	SideSize = 2.0
)

func main() {
	solid := model3d.JoinedSolid{
		SlopedTriangle{
			Base1: model3d.Coord3D{},
			Base2: model3d.X(SideSize),
			Tip:   model3d.XYZ(SideSize/2, Slope, SideSize*math.Sqrt(3)/2),
		},
		SlopedTriangle{
			Base1: model3d.Coord3D{X: Offset, Y: Slope},
			Base2: model3d.Coord3D{X: SideSize + Offset, Y: Slope},
			Tip: model3d.Coord3D{
				X: SideSize/2 + Offset,
				Y: 0,
				Z: SideSize * math.Sqrt(3) / 2,
			},
		},
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8)
	mesh = mesh.EliminateCoplanar(1e-5)

	log.Println("Saving...")
	mesh.SaveGroupedSTL("mesh.stl")

	log.Println("Rendering...")
	cameraOrigin := mesh.Max().Mid(mesh.Min()).Add(model3d.Coord3D{Y: -3, Z: 2})
	render3d.SaveRendering("rendering.png", mesh, cameraOrigin, 300, 300, nil)
}

type SlopedTriangle struct {
	Base1 model3d.Coord3D
	Base2 model3d.Coord3D
	Tip   model3d.Coord3D
}

func (s SlopedTriangle) Min() model3d.Coord3D {
	return s.Base1.Min(s.Base2).Min(s.Tip).Sub(model3d.XYZ(Size, Size, Size))
}

func (s SlopedTriangle) Max() model3d.Coord3D {
	return s.Base1.Max(s.Base2).Max(s.Tip).Add(model3d.XYZ(Size, Size, Size))
}

func (s SlopedTriangle) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}
	return checkInPole(s.Base1, s.Base2, c, 0.1) ||
		checkInPole(s.Base1, s.Tip, c, 0) ||
		checkInPole(s.Base2, s.Tip, c, 0)
}

func checkInPole(c1, c2, c model3d.Coord3D, slack float64) bool {
	v := c2.Sub(c1)
	baseLen := v.Norm()
	frac := v.Dot(c.Sub(c1)) / (baseLen * baseLen)
	if frac < -slack || frac > 1+slack {
		return false
	}
	basis1, basis2 := v.OrthoBasis()
	return math.Abs(basis1.Dot(c.Sub(c1))) < Size*(1+slack) &&
		math.Abs(basis2.Dot(c.Sub(c1))) < Size*(1+slack)
}
