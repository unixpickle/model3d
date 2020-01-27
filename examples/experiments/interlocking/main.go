package main

import (
	"math"

	"github.com/unixpickle/model3d"
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
			Base2: model3d.Coord3D{X: SideSize},
			Tip:   model3d.Coord3D{X: SideSize / 2, Y: Slope, Z: SideSize * math.Sqrt(3) / 2},
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
	mesh := model3d.SolidToMesh(solid, 0.01, 0, -1, 5)
	mesh.SaveGroupedSTL("mesh.stl")
}

type SlopedTriangle struct {
	Base1 model3d.Coord3D
	Base2 model3d.Coord3D
	Tip   model3d.Coord3D
}

func (s SlopedTriangle) Min() model3d.Coord3D {
	return s.Base1.Min(s.Base2).Min(s.Tip).Sub(model3d.Coord3D{X: Size, Y: Size, Z: Size})
}

func (s SlopedTriangle) Max() model3d.Coord3D {
	return s.Base1.Max(s.Base2).Max(s.Tip).Add(model3d.Coord3D{X: Size, Y: Size, Z: Size})
}

func (s SlopedTriangle) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(s, c) {
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
