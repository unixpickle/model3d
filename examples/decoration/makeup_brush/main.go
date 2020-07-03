package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	HandleHeight       = 0.7
	HandleBottomRadius = 0.1
	HandleTopRadius    = 0.2

	BrushHeight         = 0.3
	BrushTopRadius      = 0.4
	BrushRippleDepth    = 0.02
	BrushTopRippleDepth = 0.01
	BrushRippleFreq     = 25.0
	BrushTopRippleFreq  = 40.0
)

func main() {
	solid := model3d.JoinedSolid{
		BrushSolid{},
		// Rounded bottom for brush.
		&model3d.Sphere{
			Center: model3d.Coord3D{Z: -HandleHeight},
			Radius: HandleBottomRadius,
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.01, 8).Blur(-1)
	mesh.SaveGroupedSTL("brush.stl")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 200, nil)
}

type BrushSolid struct{}

func (b BrushSolid) Min() model3d.Coord3D {
	return model3d.XYZ(-BrushTopRadius, -BrushTopRadius, -HandleHeight)
}

func (b BrushSolid) Max() model3d.Coord3D {
	return model3d.XYZ(BrushTopRadius, BrushTopRadius, BrushHeight+BrushTopRadius/2)
}

func (b BrushSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}

	centerDist := c.XY().Norm()

	if c.Z < 0 {
		// Handle
		frac := (c.Z + HandleHeight) / HandleHeight
		radius := frac*HandleTopRadius + (1-frac)*HandleBottomRadius
		return centerDist <= radius
	}

	frac := math.Max(0, (BrushHeight-c.Z)/BrushHeight)
	radius := frac*HandleTopRadius + (1-frac)*BrushTopRadius
	theta := math.Atan2(c.Y, c.X)
	radius -= BrushRippleDepth * BrushRippleFunc(theta)
	if c.Z < BrushHeight {
		return centerDist <= radius
	}

	// Top of the brush.
	topRadius := BrushTopRadius - BrushRippleDepth*BrushRippleFunc(theta)
	if centerDist >= topRadius {
		return false
	}
	height := math.Sqrt(1-math.Pow((10+centerDist/topRadius)/11, 2)) * topRadius
	height += BrushTopRippleDepth * BrushTopRippleFunc((centerDist/topRadius)*BrushTopRippleFreq)
	return c.Z-BrushHeight < height
}

func BrushRippleFunc(theta float64) float64 {
	value := math.Sin(40*theta) + 0.2*math.Cos(54*theta) + 0.3*math.Sin(100*theta)
	return (value/1.5 + 1.0) / 2
}

func BrushTopRippleFunc(theta float64) float64 {
	value := math.Sin(40.0*theta) + 0.3*math.Cos(60*theta)
	return (value/1.3 + 1.0) / 2
}
