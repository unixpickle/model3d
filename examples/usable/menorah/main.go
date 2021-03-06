package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	HolderBottomThickness = 0.05
	HolderTopThickness    = 0.1
	HolderHeight          = 0.7
	HolderRadius          = 0.18

	Thickness = HolderRadius + HolderBottomThickness
)

func main() {
	solid := model3d.JoinedSolid{
		&model3d.SubtractedSolid{
			Positive: &model3d.Sphere{
				Center: model3d.Coord3D{Z: -0.2},
				Radius: 0.9,
			},
			Negative: &model3d.Rect{
				MinVal: model3d.XYZ(-1, -1, -2),
				MaxVal: model3d.XYZ(1, 1, 0),
			},
		},
		&model3d.Rect{
			MinVal: model3d.XYZ(-2, -0.9, -0.3),
			MaxVal: model3d.XYZ(2, 0.9, 0),
		},
		&CandleHolder{Center: model3d.Z(0.6)},
	}
	for _, pole := range GeneratePoles() {
		solid = append(solid, pole, &CandleHolder{
			Center: pole.P2,
		})
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.0125, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("menorah.stl")

	log.Println("Saving rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.Coord3D{Z: 6, Y: -10}, 500, 400, nil)
}

func GeneratePoles() []*ScewedPole {
	minError := 10000.0
	bestResult := []*ScewedPole{}

	for spacing := 0.001; spacing < 1.0; spacing += 0.001 {
		poles, curError := GeneratePolesSpacing(spacing)
		if curError < minError {
			minError = curError
			bestResult = poles
		}
	}

	return bestResult
}

func GeneratePolesSpacing(spacing float64) ([]*ScewedPole, float64) {
	pole1 := &ScewedPole{P2: model3d.Coord3D{X: -3, Z: 3.5}, Radius: Thickness}
	pole2 := pole1.Mirror()

	pole3 := pole1.MidwayUp(spacing).Mirror()
	pole4 := pole2.MidwayUp(spacing).Mirror()

	pole5 := pole1.MidwayUp(spacing).MidwayUp(spacing).Mirror()
	pole6 := pole3.MidwayUp(spacing).Mirror()
	pole7 := pole2.MidwayUp(spacing).MidwayUp(spacing).Mirror()
	pole8 := pole4.MidwayUp(spacing).Mirror()

	return []*ScewedPole{
		pole1,
		pole2,
		pole3,
		pole4,
		pole5,
		pole6,
		pole7,
		pole8,
	}, math.Abs(pole1.P2.Dist(pole5.P2) - pole5.P2.Dist(pole6.P2))
}

type ScewedPole struct {
	P1     model3d.Coord3D
	P2     model3d.Coord3D
	Radius float64
}

func (s *ScewedPole) Min() model3d.Coord3D {
	return s.P1.Min(s.P2).Sub(model3d.XYZ(s.Radius, s.Radius, s.Radius))
}

func (s *ScewedPole) Max() model3d.Coord3D {
	return s.P1.Max(s.P2).Add(model3d.XYZ(s.Radius, s.Radius, 0))
}

func (s *ScewedPole) Contains(c model3d.Coord3D) bool {
	fracP2 := (c.Z - s.P1.Z) / (s.P2.Z - s.P1.Z)
	if fracP2 < 0 || fracP2 > 1 {
		return false
	}
	centerPoint := s.P1.Scale(1 - fracP2).Add(s.P2.Scale(fracP2))
	return centerPoint.Dist(c) <= s.Radius
}

func (s *ScewedPole) Mirror() *ScewedPole {
	return &ScewedPole{
		P1:     s.P1,
		P2:     model3d.XYZ(s.P1.X*2-s.P2.X, s.P2.Y, s.P2.Z),
		Radius: s.Radius,
	}
}

func (s *ScewedPole) MidwayUp(spacing float64) *ScewedPole {
	frac := spacing / s.P2.Sub(s.P1).Norm()
	newP1 := s.P2.Scale(0.5 + frac).Add(s.P1.Scale(0.5 - frac))
	return &ScewedPole{
		P1:     newP1,
		P2:     s.P2,
		Radius: s.Radius,
	}
}

type CandleHolder struct {
	Center model3d.Coord3D
}

func (c *CandleHolder) Min() model3d.Coord3D {
	dx := HolderRadius + HolderTopThickness
	return c.Center.Sub(model3d.XYZ(dx, dx, 0))
}

func (c *CandleHolder) Max() model3d.Coord3D {
	dx := HolderRadius + HolderTopThickness
	return c.Center.Add(model3d.XYZ(dx, dx, HolderHeight))
}

func (c *CandleHolder) Contains(coord model3d.Coord3D) bool {
	coord = coord.Sub(c.Center)
	zFrac := coord.Z / HolderHeight
	if zFrac < 0 || zFrac > 1 {
		return false
	}
	thickness := zFrac*HolderTopThickness + (1-zFrac)*HolderBottomThickness
	radius := coord.XY().Norm()
	return radius >= HolderRadius && radius <= HolderRadius+thickness
}
