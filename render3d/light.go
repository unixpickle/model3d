package render3d

import (
	"math"

	"github.com/unixpickle/model3d"
)

// Color is a linear RGB color, where X, Y, and Z store R,
// G, and B respectively.
//
// Note that these colors are NOT sRGB (the standard),
// since sRGB values do not represent linear brightness.
//
// Colors should be positive, but they are not bounded on
// the positive side, since light isn't in the real world.
type Color = model3d.Coord3D

// ClampColor clamps the color into the range [0, 1].
func ClampColor(c Color) Color {
	return c.Max(Color{}).Min(Color{X: 1, Y: 1, Z: 1})
}

// NewColor creates a Color with a given brightness.
func NewColor(b float64) Color {
	return Color{X: b, Y: b, Z: b}
}

// NewColorRGB creates a Color from sRGB values.
func NewColorRGB(r, g, b float64) Color {
	return Color{X: gammaExpand(r), Y: gammaExpand(g), Z: gammaExpand(b)}
}

// RGB gets sRGB values for a Color.
func RGB(c Color) (float64, float64, float64) {
	return gammaCompress(c.X), gammaCompress(c.Y), gammaCompress(c.Z)
}

func gammaCompress(u float64) float64 {
	if u <= 0.0031308 {
		return 12.92 * u
	} else {
		return 1.055*math.Pow(u, 1/2.4) - 0.055
	}
}

func gammaExpand(u float64) float64 {
	if u <= 0.04045 {
		return u / 12.92
	} else {
		return math.Pow((u+0.055)/1.055, 2.4)
	}
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
	// In essence, when doing simple ray tracing, we want
	// the brightest part of a lambertian surface to have
	// the same brightness as the point light.
	density := 0.25 * math.Max(0, normal.Dot(pointToLight.Scale(1/dist)))

	return color.Scale(density)
}
