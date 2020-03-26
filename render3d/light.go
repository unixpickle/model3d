package render3d

import (
	"math"

	"github.com/unixpickle/model3d"
)

// Color is an RGB color, where components are X, Y, and Z
// respectively.
//
// Colors should be positive, but they are not bounded on
// the positive side, since light isn't in the real world.
type Color = model3d.Coord3D

// ClampColor clamps the color into the range [0, 1].
func ClampColor(c Color) Color {
	return c.Max(Color{}).Min(Color{X: 1, Y: 1, Z: 1})
}

// A PointLight is a light eminating from a point and
// going in all directions equally.
type PointLight struct {
	Origin model3d.Coord3D
	Color  Color

	// If true, the ray tracer should use an inverse
	// square relation to dim this light as it gets
	// farther from an object.
	QuadDropoff bool
}

// ColorAtDistance gets the Color produced by this light
// at some distance.
func (p *PointLight) ColorAtDistance(distance float64) Color {
	if !p.QuadDropoff {
		return p.Color
	}
	return p.Color.Scale(1 / (distance * distance))
}

// ShadeCollision determines a scaled color for a surface
// light collision.
func (p *PointLight) ShadeCollision(normal, pointToLight model3d.Coord3D) Color {
	dist := pointToLight.Norm()
	color := p.ColorAtDistance(dist)

	// Multiply by a density correction that comes from
	// lambertian shading.
	// The 0.5 comes from the fact that the light is
	// always sampled, while it should only be sampled on
	// one half of the hemisphere.
	density := 0.5 * math.Max(0, normal.Dot(pointToLight.Scale(1/dist)))

	return color.Scale(density)
}
