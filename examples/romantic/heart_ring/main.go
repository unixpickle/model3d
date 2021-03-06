package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
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
		RingSolid(),
		NewHeartSolid(),
	}
	log.Println("Creating mesh...")
	m := model3d.MarchingCubesSearch(solid, 0.0015, 8).Blur(-1, -1, -1, -1, -1)
	log.Println("Eliminating co-planar...")
	m = m.EliminateCoplanar(1e-8)
	log.Println("Saving mesh...")
	m.SaveGroupedSTL("ring.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", m, 3, 3, 300, nil)
}

func RingSolid() model3d.Solid {
	min := model3d.XYZ(-(RingRadius + RingThickness), -(RingRadius + RingThickness), -RingLength/2)
	return model3d.CheckedFuncSolid(min, min.Scale(-1), func(c model3d.Coord3D) bool {
		rad := c.XY().Norm()
		if rad < RingRadius || rad > RingRadius+RingThickness {
			return false
		}
		thicknessDelta := RingCurveRadius * (1 - math.Sqrt(1-math.Pow(c.Z/RingCurveRadius, 2)))
		if rad < RingRadius+thicknessDelta || rad > RingRadius+RingThickness-thicknessDelta {
			return false
		}
		return true
	})
}

type HeartSolid struct {
	Outline   model2d.Collider
	Engraving model2d.Collider
}

func NewHeartSolid() *HeartSolid {
	m := model2d.MustReadBitmap("heart.png", nil).FlipY().Mesh()
	m = m.SmoothSq(10)
	outline := model2d.MeshToCollider(m)

	m = model2d.MustReadBitmap("letters.png", nil).FlipY().FlipX().Mesh()
	m = m.SmoothSq(10)
	engraving := model2d.MeshToCollider(m)

	return &HeartSolid{
		Outline:   outline,
		Engraving: engraving,
	}
}

func (h *HeartSolid) Min() model3d.Coord3D {
	return model3d.XYZ(-HeartWidth/2, RingRadius+HeartSpacing, -HeartHeight/2)
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
