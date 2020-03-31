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
	// SampleFocus samples a source direction for a
	// collision, focusing on some aspect of a scene.
	SampleFocus(gen *rand.Rand, mat Material, point,
		normal, dest model3d.Coord3D) model3d.Coord3D

	// FocusDensity computes the probability density ratio
	// of sampling the source direction from a point,
	// relative to density on a unit sphere.
	FocusDensity(mat Material, point, normal, source,
		dest model3d.Coord3D) float64
}

// A PhongFocusPoint uses a distribution proportional to
// cos(theta)^alpha, just like phong shading.
type PhongFocusPoint struct {
	Target model3d.Coord3D

	// Alpha is the amount of focus to put on the target
	// direction.
	Alpha float64

	// MaterialFilter, if non-nil, is called to see if a
	// given material needs to be focused on a light.
	MaterialFilter func(m Material) bool
}

// SampleFocus samples a point that is more
// concentrated in the direction of Target.
func (p *PhongFocusPoint) SampleFocus(gen *rand.Rand, mat Material, point, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	if p.Target == point || !p.focusMaterial(mat) {
		return mat.SampleSource(gen, normal, dest)
	}
	direction := point.Sub(p.Target).Normalize()
	return sampleAroundDirection(gen, p.Alpha, direction)
}

// FocusDensity gives the probability density ratio for
// the given direction.
func (p *PhongFocusPoint) FocusDensity(mat Material, point, normal, source,
	dest model3d.Coord3D) float64 {
	if p.Target == point || !p.focusMaterial(mat) {
		return mat.SourceDensity(normal, source, dest)
	}
	direction := point.Sub(p.Target).Normalize()
	return densityAroundDirection(p.Alpha, direction, source)
}

func (p *PhongFocusPoint) focusMaterial(mat Material) bool {
	if p.MaterialFilter != nil {
		return p.MaterialFilter(mat)
	}
	return true
}
