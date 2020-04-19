package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
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

// SphereFocusPoint uses a distribution that samples rays
// which hit a specific sphere.
type SphereFocusPoint struct {
	// Center and Radius controls the position and size of
	// the sphere to sample.
	Center model3d.Coord3D
	Radius float64

	// MaterialFilter, if non-nil, is called to see if a
	// given material needs to be focused on a light.
	MaterialFilter func(m Material) bool
}

// SampleFocus samples a point that is more
// concentrated in the direction of Target.
func (s *SphereFocusPoint) SampleFocus(gen *rand.Rand, mat Material, point, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	if s.Center.Dist(point) < s.Radius || !s.focusMaterial(mat) {
		return mat.SampleSource(gen, normal, dest)
	}
	minCos, dir := s.focusInfo(point)
	return sampleAroundUniform(gen, minCos, dir)
}

// FocusDensity gives the probability density ratio for
// the given direction.
func (s *SphereFocusPoint) FocusDensity(mat Material, point, normal, source,
	dest model3d.Coord3D) float64 {
	if s.Center.Dist(point) < s.Radius || !s.focusMaterial(mat) {
		return mat.SourceDensity(normal, source, dest)
	}
	minCos, dir := s.focusInfo(point)
	return densityAroundUniform(minCos, dir, source)
}

func (s *SphereFocusPoint) focusMaterial(mat Material) bool {
	if s.MaterialFilter != nil {
		return s.MaterialFilter(mat)
	}
	return true
}

func (s *SphereFocusPoint) focusInfo(point model3d.Coord3D) (minCos float64, dir model3d.Coord3D) {
	// If we are at a distance d from a sphere, and we shoot
	// a ray with an angle a (relative to the line going to
	// the center of the sphere), then this equation models
	// the distance from the center c to the ray at time t.
	//
	//     c = cos(a), s = sin(a)
	//     x(t) = c*t, y(t) = s*t
	//     d(t) = (c*t - d)^2 + (s*t)^2
	//
	// Now let's minimize d(t):
	//
	//     d'(t) = 0 = 2*(c*t - d)*c + 2*(s*t)*s
	//               = 2*c^2*t - 2*d*c + 2*s^2*t
	//               = (2*c^2 + 2*s^2)*t - 2*d*c
	//     t = 2*d*c / (2*c^2 + 2*s^2)
	//       = 2*d*c / 2 = d*cos(a)
	//     d(t) = (d*cos(a)^2 - d)^2 + (d*sin(a)*cos(a))^2
	//          = d^2*cos(a)^4 - 2*d^2*cos(a)^2 + d^2 + d^2*sin(a)^2*cos(a)^2
	//          = d^2 * (cos(a)^4 - 2*cos(a)^2 + 1 + sin(a)^2*cos(a)^2)
	//          = d^2 * (1 + cos(a)^2*(cos(a)^2 - 2 + sin(a)^2))
	//          = d^2 * (1 - cos(a)^2)
	//
	// If we want d(t) = r^2, then
	//
	//     r^2 = d^2 * (1 - cos(a)^2)
	//     cos(a)^2 = 1 - r^2 / d^2
	//     cos(a) = sqrt(1 - r^2/d^2)
	//

	direction := point.Sub(s.Center)
	dist := direction.Norm()
	if dist < s.Radius {
		// Shouldn't ever happen, but incase the check
		// in the BSDF fails due to rounding error, let's
		// have a solution here.
		return 0, direction.Scale(1 / dist)
	}
	ratio := s.Radius / dist
	return math.Sqrt(1 - ratio*ratio), direction.Scale(1 / dist)
}

func sampleAroundUniform(gen *rand.Rand, minCos float64,
	direction model3d.Coord3D) model3d.Coord3D {
	// p(theta) ~ sin(theta)
	// p(theta < x) = (1 - cos(theta)) / (1 - minCos)
	// let alpha = 1/(1-minCos)
	// u = (1 - cos(theta)) * alpha
	// theta = acos(1-u/alpha)

	lat := math.Acos(1 - gen.Float64()*(1-minCos))
	lon := gen.Float64() * 2 * math.Pi

	xAxis, zAxis := direction.OrthoBasis()

	lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
	return direction.Scale(math.Cos(lat)).Add(lonPoint.Scale(math.Sin(lat)))
}

func densityAroundUniform(minCos float64, direction, sample model3d.Coord3D) float64 {
	if direction.Dot(sample) < minCos {
		return 0
	}
	return 2 / (1 - minCos)
}
