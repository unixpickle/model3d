package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d"
)

// A SampleFunc generates random unit directions along
// with a weight specifying some relative probability
// density.
type SampleFunc func() (model3d.Coord3D, float64)

// A Material determines how light bounces off a locally
// flat surface.
type Material interface {
	// Reflect gets the amount of light that bounces off
	// the surface into a given direction.
	//
	// Both arguments should be unit vectors.
	//
	// The source argument specifies the direction that
	// light is coming in and hitting the surface.
	//
	// The dest argument specifies the direction in which
	// the light is to bounce, and where we would like to
	// know the intensity.
	//
	// Returns a multiplicative mask for incoming light.
	//
	// The outgoing flux should be less than or equal to
	// the incoming flux.
	// Thus, the outgoing Color should be, on expectation
	// over random unit source vectors, less than 1 in all
	// components.
	Reflect(normal, source, dest model3d.Coord3D) Color

	// BackSampler creates a function that samples random
	// source vectors for a given dest vector.
	//
	// The main purpose of BackSampler is to provide a
	// lower-variance means of integrating over the
	// incoming light using importance sampling.
	// The weight of a direction should be equal to the
	// ratio of the sampling density to the uniform
	// density over the unit sphere.
	BackSampler(normal, dest model3d.Coord3D) SampleFunc

	// Luminance is the amount of light directly given off
	// by the surface in the normal direction.
	Luminance() Color

	// Ambience is the baseline color to use for all
	// collisions with this surface for rendering.
	// It ensures that every surface is rendered at least
	// some amount.
	Ambience() Color
}

// LambertMaterial is a completely matte material.
type LambertMaterial struct {
	ReflectColor   Color
	LuminanceColor Color
	AmbienceColor  Color
}

func (l *LambertMaterial) Reflect(normal, source, dest model3d.Coord3D) Color {
	if dest.Dot(normal) < 0 {
		return Color{}
	}
	return l.DiffuseColor.Scale(math.Max(0, -normal.Dot(source)))
}

func (l *LambertMaterial) BackSampler(normal, dest model3d.Coord3D) SampleFunc {
	return func() (model3d.Coord3D, float64) {
		c := model3d.NewCoord3DRandUnit()
		if c.Dot(normal) > 0 {
			return c.Scale(-1), 0.5
		}
		return c, 0.5
	}
}

func (l *LambertMaterial) Luminance() Color {
	return l.LuminanceColor
}

func (l *LambertMaterial) Ambience() Color {
	return l.AmbienceColor
}

type PhongMaterial struct {
	// Alpha is the exponent for the reflection term.
	//
	// A value of 0 makes PhongMaterial similar to a
	// lambert material, except that the viewer cannot be
	// at a negative dot product to the reflected rays.
	//
	// Higher values result in more reflectiveness.
	Alpha float64

	ReflectColor   Color
	LuminanceColor Color
	AmbienceColor  Color
}

func (p *PhongMaterial) Reflect(normal, source, dest model3d.Coord3D) Color {
	if dest.Dot(normal) < 0 {
		return Color{}
	}
	reflection := normal.Reflect(source)
	refDot := reflection.Dot(dest)
	if refDot < 0 {
		return Color{}
	}
	intensity := -normal.Dot(source) * math.Pow(refDot, p.Alpha)
	return p.ReflectColor.Scale(math.Max(0, intensity))
}

func (p *PhongMaterial) BackSampler(normal, dest model3d.Coord3D) SampleFunc {
	// Create polar coordinates around the reflection, and
	// use alpha to decide the concentration.
	//
	// p(lat) ~ sin(lat) * cos(lat)^alpha
	// P(lat<x) ~ 1 - cos(lat)^(alpha+1)
	//
	// to sample from lat using uniform v, we can do:
	// v = 1 - cos(lat)^(alpha+1)
	// lat = acos((1 - v)^(1/(alpha+1)))
	// let 1 - v be a new random variable v:
	// lat = acos(v^(1/(alpha+1)))

	// Let's do a change of variables to figure out the
	// proper weights:
	//
	// u and v are random uniform variables.
	// lon = 2 * pi * u
	// lat = acos(v^(1/(alpha+1)))
	// dx = sin(lat) * d(lon)
	//    = sin(lat) * 2 * pi * du
	//    = sqrt(1 - v^(2/(alpha+1))) * 2 * pi * du
	// dy = d(lat)
	//    = -(v^(1/(alpha+1)-1)) / ((alpha+1)*sin(lat)) * dv
	//
	// The jacobian is diagonal, so the determinant is:
	// dx dy = 2 * pi * v^(1/(alpha+1)-1) / (alpha + 1) * du dv
	//
	// Dividing by the entire area of the sphere gives:
	//
	//     1/2 * v^(1/(alpha+1)-1) / (alpha + 1)
	//

	reflection := normal.Reflect(dest)
	xAxis, zAxis := reflection.OrthoBasis()
	return func() (model3d.Coord3D, float64) {
		u := rand.Float64()
		v := rand.Float64()

		lon := 2 * math.Pi * u
		lat := math.Acos(math.Pow(v, 1/(p.Alpha+1)))

		lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
		point := reflection.Scale(math.Cos(lat)).Add(lonPoint.Scale(math.Sin(lat)))
		weight := math.Pow(v, 1/(p.Alpha+1)-1) / (2 * (p.Alpha + 1))

		return point, weight
	}
}

func (p *PhongMaterial) Luminance() Color {
	return p.LuminanceColor
}

func (p *PhongMaterial) Ambience() Color {
	return p.AmbienceColor
}
