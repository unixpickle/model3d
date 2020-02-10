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

		// Use an accurate bisection method to compute
		// normals.
		//
		// This method uses very small vectors to tell if
		// if a given direction has a positive dot product
		// with surface normals.
		// Therefore, if this epsilon is too large, it
		// will yield totally incorrect results.
		// If the epsilon is too small (i.e. 1e-20), then
		// the probe vectors may be too small to pass the
		// surface boundary at all, and the results will
		// also become inaccurate.
		NormalBisectEpsilon: 1e-5,
	}
	model3d.SaveRandomGrid("rendering.png", collider, 4, 4, 300, 300)
}
