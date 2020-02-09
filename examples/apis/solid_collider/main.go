package main

import (
	"github.com/unixpickle/model3d"
)

func main() {
	// A solid representing a sphere with a cylinder taken
	// out of the side.
	solid := &model3d.SubtractedSolid{
		Positive: &model3d.SphereSolid{
			Radius: 0.5,
		},
		Negative: &model3d.CylinderSolid{
			P1:     model3d.Coord3D{Y: 0.3, Z: -0.5},
			P2:     model3d.Coord3D{Y: 0.3, Z: 0.5},
			Radius: 0.3,
		},
	}

	// Create a collider from an arbitrary solid.
	collider := &model3d.SolidCollider{
		Solid: solid,

		// This resolution is good enough for our rendering
		// purposes. Making it larger makes the image look
		// blocky around the edges.
		Epsilon: 0.005,

		// Larger values make the lighting more accurate,
		// while smaller values result in some grain.
		NormalSamples: 500,
	}
	model3d.SaveRandomGrid("rendering.png", collider, 3, 3, 200, 200)
}
