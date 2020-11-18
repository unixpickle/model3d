package main

import (
	"log"
	"math"
	"sort"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	BallRadius = 0.2
	RungRadius = 0.3
	NumRungs   = 4
	MaxRadius  = NumRungs*RungRadius*2 + RungRadius
	Slope      = 0.15
)

func main() {
	log.Println("Creating ball...")
	ballMesh := model3d.MarchingCubesSearch(&model3d.Sphere{
		Radius: BallRadius,
	}, 0.01, 8)
	ballMesh.SaveGroupedSTL("ball.stl")

	log.Println("Creating solid...")
	solid := NewSpiralSolid()
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 16)
	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("spiral.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type SpiralSolid struct {
	MaxHeight float64
	Dists     *SpiralDistances
}

func NewSpiralSolid() *SpiralSolid {
	dists := NewSpiralDistances()
	return &SpiralSolid{
		MaxHeight: dists.Distances[len(dists.Distances)-1] * Slope,
		Dists:     dists,
	}
}

func (s *SpiralSolid) Min() model3d.Coord3D {
	return model3d.XY(-MaxRadius, -MaxRadius)
}

func (s *SpiralSolid) Max() model3d.Coord3D {
	return model3d.XYZ(MaxRadius, MaxRadius, s.MaxHeight)
}

func (s *SpiralSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}
	theta := SpiralTheta(c.XY())
	if theta > math.Pi*2*NumRungs {
		return false
	}
	dist := s.Dists.Lookup(theta)
	maxZ := s.MaxHeight - dist*Slope

	// Inset to make a track for the ball to roll.
	center := SpiralCenter(theta)
	centerDist := c.XY().Dist(center)
	if centerDist < BallRadius {
		maxZ -= (BallRadius*BallRadius - centerDist*centerDist) / BallRadius
	}

	return c.Z < maxZ
}

func SpiralTheta(c model2d.Coord) float64 {
	theta := math.Atan2(c.Y, c.X)
	if theta < 0 {
		theta += math.Pi * 2
	}
	for SpiralCenter(theta).Norm()+RungRadius < c.Norm() {
		theta += math.Pi * 2
	}
	return theta
}

func SpiralCenter(theta float64) model2d.Coord {
	return model2d.NewCoordPolar(theta, RungRadius*theta/math.Pi)
}

type SpiralDistances struct {
	Thetas    []float64
	Distances []float64
}

func NewSpiralDistances() *SpiralDistances {
	res := &SpiralDistances{
		Thetas:    []float64{},
		Distances: []float64{},
	}
	theta := 0.0
	dist := 0.0
	for theta < NumRungs*math.Pi*2 {
		res.Thetas = append(res.Thetas, theta)
		res.Distances = append(res.Distances, dist)
		dist += SpiralCenter(theta).Dist(SpiralCenter(theta + 0.01))
		theta += 0.01
	}
	return res
}

func (s *SpiralDistances) Lookup(theta float64) float64 {
	idx := sort.SearchFloat64s(s.Thetas, theta)
	if idx == 0 {
		return s.Distances[0]
	} else if idx >= len(s.Distances)-1 {
		return s.Distances[len(s.Distances)-1]
	}
	// Linearly interpolate.
	d1 := s.Distances[idx-1]
	d2 := s.Distances[idx]
	t1 := s.Thetas[idx-1]
	t2 := s.Thetas[idx]
	frac := (theta - t1) / (t2 - t1)
	return frac*d2 + (1-frac)*d1
}
