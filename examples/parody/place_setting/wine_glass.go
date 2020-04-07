package main

import (
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/model2d"
)

const (
	WineGlassCupWidth     = 1.5
	WineGlassCupHeight    = 1.5
	WineGlassCupMinRadius = 0.2
	WineGlassCupTopRadius = 1.3 / 2
	WineGlassCupThickness = 0.05

	WineGlassBaseWidth      = 1.0
	WineGlassBaseHeight     = 0.05
	WineGlassStemRadius     = 0.1
	WineGlassStemHeight     = 1.2
	WineGlassStemTransition = 0.25
)

func CreateWineGlass() model3d.Solid {
	return model3d.StackedSolid{
		WineGlassStem{},
		WineGlassCup{},
	}
}

type WineGlassCup struct{}

func (w WineGlassCup) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -WineGlassCupWidth / 2, Y: -WineGlassCupWidth / 2}
}

func (w WineGlassCup) Max() model3d.Coord3D {
	return model3d.Coord3D{X: WineGlassCupWidth / 2, Y: WineGlassCupWidth / 2,
		Z: WineGlassCupHeight}
}

func (w WineGlassCup) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(w, c) {
		return false
	}
	r := c.Coord2D().Norm()

	curve := model2d.BezierCurve{
		{X: 0, Y: WineGlassCupMinRadius},
		{X: WineGlassCupHeight / 3.0, Y: WineGlassCupWidth / 2},
		{X: 2 * WineGlassCupHeight / 3.0, Y: WineGlassCupWidth / 2},
		{X: WineGlassCupHeight, Y: WineGlassCupTopRadius},
	}

	maxRadius := curve.EvalX(c.Z)
	if c.Z < WineGlassCupThickness {
		return r <= maxRadius
	}
	return r <= maxRadius && r >= maxRadius-WineGlassCupThickness
}

type WineGlassStem struct{}

func (w WineGlassStem) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -WineGlassBaseWidth / 2, Y: -WineGlassBaseWidth / 2}
}

func (w WineGlassStem) Max() model3d.Coord3D {
	return model3d.Coord3D{X: WineGlassBaseWidth / 2, Y: WineGlassBaseWidth / 2,
		Z: WineGlassStemHeight}
}

func (w WineGlassStem) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(w, c) {
		return false
	}

	r := c.Coord2D().Norm()
	radius := WineGlassStemRadius
	if c.Z < WineGlassStemTransition {
		// Transition to base.
		curve := model2d.BezierCurve{
			{X: 0, Y: WineGlassBaseWidth / 2},
			{X: WineGlassBaseHeight, Y: WineGlassStemRadius},
			{X: WineGlassStemTransition, Y: WineGlassStemRadius},
		}
		radius = curve.EvalX(c.Z)
	} else if c.Z > WineGlassStemHeight-WineGlassStemTransition {
		// Transition to cup.
		curve := model2d.BezierCurve{
			{X: WineGlassStemHeight - WineGlassStemTransition, Y: WineGlassStemRadius},
			{X: WineGlassStemHeight - WineGlassStemTransition/1.5, Y: WineGlassStemRadius},
			{X: WineGlassStemHeight, Y: WineGlassCupMinRadius},
		}
		radius = curve.EvalX(c.Z)
	}
	return r <= radius
}
