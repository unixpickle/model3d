package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/model3d"
)

const (
	// Between a size 7.5 and a size 8.
	RingDiameter = 0.7
	RingRadius   = RingDiameter / 2

	RingLength    = 0.1
	RingThickness = 0.06

	// Controls how curved the ring is.
	// Lower values mean more curvature.
	RingCurveRadius = 0.5

	HeartSpacing = 0.01
	HeartWidth   = 0.3
	HeartHeight  = HeartWidth * 456.0 / 512.0
	EngraveDepth = 0.015
)

func main() {
	solid := model3d.JoinedSolid{
		RingSolid{},
		NewHeartSolid(),
	}
	m := model3d.MarchingCubesSearch(solid, 0.0015, 8).Blur(-1, -1, -1, -1, -1)
	m.SaveGroupedSTL("ring.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(m), 3, 3, 300, 300)
}

type RingSolid struct{}

func (r RingSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -(RingRadius + RingThickness), Y: -(RingRadius + RingThickness),
		Z: -RingLength / 2}
}

func (r RingSolid) Max() model3d.Coord3D {
	return r.Min().Scale(-1)
}

func (r RingSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(r, c) {
		return false
	}
	rad := c.Coord2D().Norm()
	if rad < RingRadius || rad > RingRadius+RingThickness {
		return false
	}
	thicknessDelta := RingCurveRadius * (1 - math.Sqrt(1-math.Pow(c.Z/RingCurveRadius, 2)))
	if rad < RingRadius+thicknessDelta || rad > RingRadius+RingThickness-thicknessDelta {
		return false
	}
	return true
}

type HeartSolid struct {
	Outline   model2d.Collider
	Engraving model2d.Collider
}

func NewHeartSolid() *HeartSolid {
	m := model2d.MustReadBitmap("heart.png", nil).FlipY().Mesh()
	m = m.Blur(0.25).Blur(0.25)
	outline := model2d.MeshToCollider(m)

	m = model2d.MustReadBitmap("letters.png", nil).FlipY().FlipX().Mesh()
	m = m.Blur(0.25).Blur(0.25)
	engraving := model2d.MeshToCollider(m)

	return &HeartSolid{
		Outline:   outline,
		Engraving: engraving,
	}
}

func (h *HeartSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -HeartWidth / 2, Y: RingRadius + HeartSpacing, Z: -HeartHeight / 2}
}

func (h *HeartSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: HeartWidth / 2, Y: RingRadius + RingThickness + EngraveDepth,
		Z: HeartHeight / 2}
}

func (h *HeartSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(h, c) {
		return false
	}
	localCoord := model3d.Coord2D{
		X: h.Outline.Max().X * (c.X - h.Min().X) / (h.Max().X - h.Min().X),
		Y: h.Outline.Max().Y * (c.Z - h.Min().Z) / (h.Max().Z - h.Min().Z),
	}
	if !model2d.ColliderContains(h.Outline, localCoord, 0) {
		return false
	}
	if c.Y < h.Max().Y-EngraveDepth {
		return true
	}
	return !model2d.ColliderContains(h.Engraving, localCoord, 0)
}
