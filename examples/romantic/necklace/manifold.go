package main

import (
	"math"

	"github.com/unixpickle/model3d"
)

type Manifold interface {
	Move(length float64)
	Convert(c model3d.Coord3D) model3d.Coord3D
}

type SpiralManifold struct {
	theta       float64
	startRadius float64
	spiralRate  float64
}

func NewSpiralManifold(startRadius, spiralRate float64) *SpiralManifold {
	return &SpiralManifold{
		startRadius: startRadius,
		spiralRate:  spiralRate,
	}
}

func (s *SpiralManifold) Move(length float64) {
	if length < 0 {
		panic("cannot go backwards")
	}
	for i := 0; i < 100; i++ {
		l := length / 100
		s.theta += l / s.currentRadius()
	}
}

func (s *SpiralManifold) Convert(c model3d.Coord3D) model3d.Coord3D {
	r := s.currentRadius()
	xBasis := model3d.Coord3D{X: math.Cos(s.theta), Y: math.Sin(s.theta)}
	yBasis := model3d.Coord3D{X: -xBasis.Y, Y: xBasis.X}
	center := xBasis.Scale(r)
	z := model3d.Coord3D{Z: c.Z}
	return center.Add(z).Add(xBasis.Scale(c.X)).Add(yBasis.Scale(c.Y))
}

func (s *SpiralManifold) currentRadius() float64 {
	return s.theta*s.spiralRate + s.startRadius
}
