package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d"
)

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

	// SampleSource samples a random source vector for a
	// given dest vector, possibly with a non-uniform
	// distribution.
	//
	// The main purpose of SampleSource is to compute a
	// the mean outgoing light using importance sampling.
	//
	// The densities returned by SourceDensity correspond
	// to this sampling distribution.
	SampleSource(normal, dest model3d.Coord3D) model3d.Coord3D

	// SourceDensity computes the density ratio of
	// arbitrary source directions under the distribution
	// used by SampleSource(). These ratios measure the
	// density divided by the density on a unit sphere.
	SourceDensity(normal, source, dest model3d.Coord3D) float64

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
	DiffuseColor   Color
	AmbienceColor  Color
	LuminanceColor Color
}

func (l *LambertMaterial) Reflect(normal, source, dest model3d.Coord3D) Color {
	if dest.Dot(normal) < 0 {
		return Color{}
	}
	return l.DiffuseColor.Scale(math.Max(0, -normal.Dot(source)))
}

func (l *LambertMaterial) SampleSource(normal, dest model3d.Coord3D) model3d.Coord3D {
	// Sample with probabilities proportional to the BRDF.
	u := rand.Float64()
	lat := math.Acos(math.Sqrt(u))
	lon := rand.Float64() * 2 * math.Pi

	xAxis, zAxis := normal.OrthoBasis()

	lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
	point := normal.Scale(-math.Cos(lat)).Add(lonPoint.Scale(math.Sin(lat)))

	return point
}

func (l *LambertMaterial) SourceDensity(normal, source, dest model3d.Coord3D) float64 {
	normalDot := -normal.Dot(source)
	if normalDot < 0 {
		return 0
	}
	return 4 * normalDot
}

func (l *LambertMaterial) Luminance() Color {
	return l.LuminanceColor
}

func (l *LambertMaterial) Ambience() Color {
	return l.AmbienceColor
}

// PhongMaterial implements the Phong reflection model.
//
// https://en.wikipedia.org/wiki/Phong_reflection_model.
type PhongMaterial struct {
	// Alpha controls the specular light, where 0 means
	// unconcentrated, and higher values mean more
	// concentrated.
	Alpha float64

	SpecularColor  Color
	DiffuseColor   Color
	LuminanceColor Color
	AmbienceColor  Color
}

func (p *PhongMaterial) Reflect(normal, source, dest model3d.Coord3D) Color {
	destDot := dest.Dot(normal)
	sourceDot := -source.Dot(normal)
	if destDot < 0 || sourceDot < 0 {
		return Color{}
	}

	color := Color{}
	if p.DiffuseColor != color {
		color = p.DiffuseColor.Scale(sourceDot)
	}

	reflection := normal.Reflect(source).Scale(-1)
	refDot := reflection.Dot(dest)
	if refDot < 0 {
		return color
	}
	intensity := sourceDot * math.Pow(refDot, p.Alpha)
	return color.Add(p.SpecularColor.Scale(intensity))
}

// SampleSource uses importance sampling to sample in
// proportion to the reflection weight of a direction.
//
// If there is a diffuse lighting term, it is mixed in for
// some fraction of the samples.
func (p *PhongMaterial) SampleSource(normal, dest model3d.Coord3D) model3d.Coord3D {
	if (p.DiffuseColor == Color{}) || rand.Intn(2) == 0 {
		return p.sampleSpecular(normal, dest)
	} else {
		return (&LambertMaterial{}).SampleSource(normal, dest)
	}
}

// SourceDensity gets the density of the SampleSource
// distribution.
func (p *PhongMaterial) SourceDensity(normal, source, dest model3d.Coord3D) float64 {
	phongWeight := p.specularDensity(normal, source, dest)
	if (p.DiffuseColor == Color{}) {
		return phongWeight
	}
	lambertWeight := (&LambertMaterial{}).SourceDensity(normal, source, dest)
	return (phongWeight + lambertWeight) / 2
}

// sampleSpecular samples source vectors weighted to
// emphasize specular reflections.
func (p *PhongMaterial) sampleSpecular(normal, dest model3d.Coord3D) model3d.Coord3D {
	// Create a probability density matching the
	// specular part of the BRDF.
	//
	//     p(cos(lat)=x) = x^alpha * (alpha + 1)
	//     p(cos(lat)<x) = x^(alpha+1)
	//     p(lat<t) = p(cos(lat)>cos(t)) = 1 - cos(t)^(alpha+1)
	//
	// Now we can convert this distribution into a func of
	// a uniform random variable, v:
	//
	//     lat = acos((1-v)^(1/(alpha+1)))
	//
	// Since 1-v is also a uniform random variable, we
	// will simply use:
	//
	//     lat = acos(v^(1/(alpha+1)))
	//
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

	reflection := normal.Reflect(dest).Scale(-1)
	xAxis, zAxis := reflection.OrthoBasis()

	u := rand.Float64()
	v := rand.Float64()

	lon := 2 * math.Pi * u
	lat := math.Acos(math.Pow(v, 1/(p.Alpha+1)))

	lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
	return reflection.Scale(math.Cos(lat)).Add(lonPoint.Scale(math.Sin(lat)))
}

func (p *PhongMaterial) specularDensity(normal, source, dest model3d.Coord3D) float64 {
	reflection := normal.Reflect(source).Scale(-1)
	reflectionDot := reflection.Dot(dest)
	if reflectionDot < 0 {
		return 0
	}
	v := math.Pow(reflectionDot, p.Alpha+1)
	return (2 * (p.Alpha + 1)) / math.Pow(v, 1/(p.Alpha+1)-1)
}

func (p *PhongMaterial) Luminance() Color {
	return p.LuminanceColor
}

func (p *PhongMaterial) Ambience() Color {
	return p.AmbienceColor
}
