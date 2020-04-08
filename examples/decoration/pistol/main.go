package main

import "github.com/unixpickle/model3d"

const (
	Width           = 0.8
	HandleThickness = 1.2
	HandleOffset    = 0.3
	Length          = 4.0
	HandleLength    = 2.0
	HandleSlope     = 0.2
	GrooveSize      = 0.05
	GrooveLength    = 1.0
)

func main() {
	solid := model3d.JoinedSolid{
		Body{},
		&model3d.TorusSolid{
			Center:      model3d.Coord3D{X: Length - HandleThickness - 0.2, Y: Width / 2, Z: Width + 0.2},
			Axis:        model3d.Coord3D{Y: 1},
			InnerRadius: 0.1,
			OuterRadius: 0.8,
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	mesh.SaveGroupedSTL("pistol.stl")
}

type Body struct{}

func (b Body) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (b Body) Max() model3d.Coord3D {
	return model3d.Coord3D{X: Length + HandleLength*HandleSlope, Y: Width, Z: HandleLength + Width}
}

func (b Body) Contains(c model3d.Coord3D) bool {
	if c.Min(b.Min()) != b.Min() || c.Max(b.Max()) != b.Max() {
		return false
	}
	if c.Z < GrooveSize && c.X > Length-GrooveLength && c.X < Length {
		if int(c.X/GrooveSize)%2 == 0 {
			return false
		}
	}
	if c.Z < Width && c.X < Length+HandleOffset {
		holeRadius := 0.2 - c.X/5
		if holeRadius < 0 {
			return true
		}
		distFromCenter := c.Dist(model3d.Coord3D{X: c.X, Y: Width / 2, Z: Width / 2})
		if distFromCenter < holeRadius {
			return false
		}
		return true
	}
	handleStart := Length - HandleThickness + (c.Z-Width)*HandleSlope
	handleEnd := handleStart + HandleThickness
	return c.X >= handleStart && c.X <= handleEnd
}
