package render3d

import (
	"math/rand"

	"github.com/unixpickle/model3d"
)

// A FocusPoint is some part of the scene that warrants
// extra rays (e.g. part of a light).
//
// Focus points perform importance sampling such that the
// point or object of interest is sampled more.
type FocusPoint interface {
	// SampleFocus takes in a given point (i.e. a ray
	// origin) and samples a unit direction facing into
	// this point from some source.
	SampleFocus(gen *rand.Rand, point model3d.Coord3D) model3d.Coord3D

	// FocusDensity computes the probability density ratio
	// of sampling the given direction from a point,
	// relative to density on a unit sphere.
	FocusDensity(point, direction model3d.Coord3D) float64
}

// A PhongFocusPoint uses a distribution proportional to
// cos(theta)^alpha, just like phong shading.
type PhongFocusPoint struct {
	Target model3d.Coord3D

	// Alpha is the amount of focus to put on the target
	// direction.
	Alpha float64
}

// SampleFocus samples a point that is more
// concentrated in the direction of Target.
func (p *PhongFocusPoint) SampleFocus(gen *rand.Rand, point model3d.Coord3D) model3d.Coord3D {
	if p.Target == point {
		return model3d.NewCoord3DRandUnit()
	}
	normal := p.Target.Sub(point).Normalize()
	mat := PhongMaterial{Alpha: p.Alpha}
	return mat.SampleSource(gen, normal, normal)
}

// FocusDensity gives the probability density ratio for
// the given direction.
func (p *PhongFocusPoint) FocusDensity(point, direction model3d.Coord3D) float64 {
	if p.Target == point {
		return 1
	}
	normal := p.Target.Sub(point).Normalize()
	mat := PhongMaterial{Alpha: p.Alpha}
	return mat.SourceDensity(normal, direction, normal)
}
