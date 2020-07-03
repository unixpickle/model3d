package main

import (
	"flag"
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var radius float64
	var depth float64
	var ripple float64
	var pointRadius float64
	var pointDepth float64
	var extraPoints bool
	flag.Float64Var(&radius, "radius", 2.5, "radius of the sun ball")
	flag.Float64Var(&depth, "depth", 1.5, "depth of the sun ball (no larger than radius)")
	flag.Float64Var(&ripple, "ripple", 0.05, "relative height of sphere ripples")
	flag.Float64Var(&pointRadius, "point-radius", 0.7, "outward radius of pointed edges")
	flag.Float64Var(&pointDepth, "point-depth", 0.1, "depth of pointed edges")
	flag.BoolVar(&extraPoints, "extra-points", false, "add an extra layer of points")
	flag.Parse()

	if depth > radius {
		essentials.Die("depth must not exceed radius")
	}
	solid := model3d.JoinedSolid{
		NewSunBall(radius, depth, ripple),
		NewPointedEdges(radius, pointRadius, pointDepth, false),
	}
	if extraPoints {
		solid = append(solid, NewPointedEdges(radius, pointRadius, pointDepth*1.5, true))
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("sun.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type SunBall struct {
	Radius float64
	Depth  float64

	SphereCenter model3d.Coord3D
	SphereRadius float64
	RippleHeight float64
}

func NewSunBall(radius, depth, ripple float64) *SunBall {
	// radius^2 + (sphereRadius - depth)^2 = sphereRadius^2
	// radius^2 + sphereRadius^2 - 2*sphereRadius*depth + depth^2 = sphereRadius^2
	// radius^2 + depth^2 = 2*sphereRadius*depth
	// (radius^2 + depth^2)/(2*depth) = sphereRadius
	sphereRadius := (radius*radius + depth*depth) / (2 * depth)
	return &SunBall{
		Radius: radius,
		Depth:  depth,

		SphereCenter: model3d.Z(depth - sphereRadius),
		SphereRadius: sphereRadius,
		RippleHeight: ripple * sphereRadius,
	}
}

func (s *SunBall) Min() model3d.Coord3D {
	return model3d.XY(-s.Radius, -s.Radius)
}

func (s *SunBall) Max() model3d.Coord3D {
	return model3d.XYZ(s.Radius, s.Radius, s.Depth)
}

func (s *SunBall) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}

	geo := c.Sub(s.SphereCenter).Geo()
	r := c.Dist(s.SphereCenter)

	rippleInset := math.Pow(math.Cos(geo.Lat*20+math.Sin(geo.Lon*5)), 2)
	rippleInset *= s.RippleHeight
	rippleInset *= c.Z / s.Depth // smaller ripples at sides

	return r < s.SphereRadius-rippleInset
}

type PointedEdges struct {
	MinRadius float64
	MaxRadius float64
	Depth     float64

	Phase     float64
	Frequency float64
}

func NewPointedEdges(baseRadius, extraRadius, depth float64, phase bool) *PointedEdges {
	circum := math.Pi * 2 * (baseRadius + extraRadius/2)
	sideLength := 2 * extraRadius / math.Sqrt(3)
	res := &PointedEdges{
		MinRadius: baseRadius,
		MaxRadius: baseRadius + extraRadius,
		Depth:     depth,

		Frequency: math.Floor(circum / sideLength),
	}
	if phase {
		res.Phase += 0.5 / res.Frequency
	}
	return res
}

func (p *PointedEdges) Min() model3d.Coord3D {
	return model3d.XY(-p.MaxRadius, -p.MaxRadius)
}

func (p *PointedEdges) Max() model3d.Coord3D {
	return model3d.XYZ(p.MaxRadius, p.MaxRadius, p.Depth)
}

func (p *PointedEdges) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	r := c.XY().Norm()
	if r < p.MinRadius {
		return false
	}
	theta := math.Atan2(c.Y, c.X) + 2*math.Pi
	modulo := math.Mod(theta/(2*math.Pi)+p.Phase, 1/p.Frequency) * p.Frequency
	radiusFrac := (p.MaxRadius - r) / (p.MaxRadius - p.MinRadius)
	return math.Abs(modulo-0.5)*2 < radiusFrac
}
