package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d"
)

const cosineEpsilon = 1e-8

// A Material determines how light bounces off a locally
// flat surface.
type Material interface {
	// BSDF gets the amount of light that bounces off the
	// surface into a given direction.
	// It differs slightly from the usual meaning of BRDF,
	// since it may include refractions into the surface.
	// Thus, the BSDF is a function on the entire sphere,
	// not just a hemisphere.
	//
	// Both arguments should be unit vectors.
	//
	// The source argument specifies the direction that
	// light is coming in and hitting the surface.
	//
	// The dest argument specifies the direction in which
	// the light is to reflect or refract, and where we
	// would like to know the intensity.
	//
	// Returns a multiplicative mask for incoming light.
	//
	// The outgoing flux should be less than or equal to
	// the incoming flux.
	// Thus, the outgoing Color should be, on expectation
	// over random unit dest vectors weighted by the
	// cosine of the outgoing angle, less than 1 in all
	// components.
	BSDF(normal, source, dest model3d.Coord3D) Color

	// SampleSource samples a random source vector for a
	// given dest vector, possibly with a non-uniform
	// distribution.
	//
	// The main purpose of SampleSource is to compute a
	// the mean outgoing light using importance sampling.
	//
	// The densities returned by SourceDensity correspond
	// to this sampling distribution.
	//
	// The gen argument is used for sampling.
	// This is more efficient than using a shared RNG for
	// multithreaded rendering.
	SampleSource(gen *rand.Rand, normal, dest model3d.Coord3D) model3d.Coord3D

	// SourceDensity computes the density ratio of
	// arbitrary source directions under the distribution
	// used by SampleSource(). These ratios measure the
	// density divided by the density on a unit sphere.
	// Thus, they measure density per steradian.
	SourceDensity(normal, source, dest model3d.Coord3D) float64

	// Emission is the amount of light directly given off
	// by the surface in the normal direction.
	Emission() Color

	// Ambient is the baseline color to use for all
	// collisions with this surface for rendering.
	// It ensures that every surface is rendered at least
	// some amount.
	Ambient() Color
}

// LambertMaterial is a completely matte material.
type LambertMaterial struct {
	DiffuseColor  Color
	AmbientColor  Color
	EmissionColor Color
}

func (l *LambertMaterial) BSDF(normal, source, dest model3d.Coord3D) Color {
	if dest.Dot(normal) < 0 || source.Dot(normal) > 0 {
		return Color{}
	}
	// Multiply by 2 since half the sphere is zero.
	return l.DiffuseColor.Scale(2)
}

func (l *LambertMaterial) SampleSource(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	// Sample with probabilities proportional to the cosine
	// property (Lamert's law).
	u := gen.Float64()
	lat := math.Acos(math.Sqrt(u))
	lon := gen.Float64() * 2 * math.Pi

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

func (l *LambertMaterial) Emission() Color {
	return l.EmissionColor
}

func (l *LambertMaterial) Ambient() Color {
	return l.AmbientColor
}

// PhongMaterial implements the Phong reflection model.
//
// https://en.wikipedia.org/wiki/Phong_reflection_model.
type PhongMaterial struct {
	// Alpha controls the specular light, where 0 means
	// unconcentrated, and higher values mean more
	// concentrated.
	Alpha float64

	SpecularColor Color
	DiffuseColor  Color
	EmissionColor Color
	AmbientColor  Color

	// NoFluxCorrection can be set to true to disable
	// the max-Phong denominator.
	NoFluxCorrection bool
}

func (p *PhongMaterial) BSDF(normal, source, dest model3d.Coord3D) Color {
	destDot := dest.Dot(normal)
	sourceDot := -source.Dot(normal)
	if destDot < 0 || sourceDot < 0 {
		return Color{}
	}

	color := Color{}
	if p.DiffuseColor != color {
		// Scale by 2 because of hemisphere restriction.
		color = p.DiffuseColor.Scale(2)
	}

	reflection := normal.Reflect(source).Scale(-1)
	refDot := reflection.Dot(dest)
	if refDot < 0 {
		return color
	}
	intensity := math.Pow(refDot, p.Alpha)

	// Divide by (integral from x=0 to pi/2 of sin(x)*cos(x)^alpha)
	intensity *= (1 + p.Alpha)

	if !p.NoFluxCorrection {
		intensity /= maximumCosine(sourceDot, destDot)
	}

	// Scale by 2 because of hemisphere restriction.
	return color.Add(p.SpecularColor.Scale(2 * intensity))
}

func maximumCosine(cos1, cos2 float64) float64 {
	cos1 = math.Abs(cos1)
	cos2 = math.Abs(cos2)
	res := math.Max(cos1, cos2)
	return math.Max(res, cosineEpsilon)
}

// SampleSource uses importance sampling to sample in
// proportion to the reflection weight of a direction.
//
// If there is a diffuse lighting term, it is mixed in for
// some fraction of the samples.
func (p *PhongMaterial) SampleSource(gen *rand.Rand, normal, dest model3d.Coord3D) model3d.Coord3D {
	if (p.DiffuseColor == Color{}) || rand.Intn(2) == 0 {
		return p.sampleSpecular(gen, normal, dest)
	} else {
		return (&LambertMaterial{}).SampleSource(gen, normal, dest)
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
func (p *PhongMaterial) sampleSpecular(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	reflection := normal.Reflect(dest).Scale(-1)
	return sampleAroundDirection(gen, p.Alpha, reflection)
}

func (p *PhongMaterial) specularDensity(normal, source, dest model3d.Coord3D) float64 {
	reflection := normal.Reflect(dest).Scale(-1)
	return densityAroundDirection(p.Alpha, reflection, source)
}

func (p *PhongMaterial) Emission() Color {
	return p.EmissionColor
}

func (p *PhongMaterial) Ambient() Color {
	return p.AmbientColor
}

// sampleAroundDirection samples directions pointing near
// direction, with nearness having more weight for higher
// alpha.
func sampleAroundDirection(gen *rand.Rand, alpha float64,
	direction model3d.Coord3D) model3d.Coord3D {
	// Create a probability density matching the
	// specular part of the BSDF.
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

	xAxis, zAxis := direction.OrthoBasis()

	u := gen.Float64()
	v := gen.Float64()

	lon := 2 * math.Pi * u
	lat := math.Acos(math.Pow(v, 1/(alpha+1)))

	lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
	return direction.Scale(math.Cos(lat)).Add(lonPoint.Scale(math.Sin(lat)))
}

// densityAroundDirection gets the density for
// sampleAroundDirection.
func densityAroundDirection(alpha float64, direction, sample model3d.Coord3D) float64 {
	dot := direction.Dot(sample)
	if dot < 0 {
		return 0
	}
	v := math.Pow(dot, alpha+1)
	return 2 * (alpha + 1) / math.Pow(v, 1/(alpha+1)-1)
}

// RefractMaterial is an approximate refraction material
// based on a delta function.
//
// Unlike other BSDFs, the BSDF of RefractMaterial is
// asymmetric, since energy is concentrated due to
// refraction.
type RefractMaterial struct {
	// IndexOfRefraction is the index of refraction of
	// this material. Values greater than one simulate
	// materials like water or glass, where light travels
	// more slowly than in space.
	IndexOfRefraction float64

	// RefractColor is the mask used for refracted flux.
	RefractColor Color
}

func (r *RefractMaterial) refract(normal, source model3d.Coord3D) model3d.Coord3D {
	sinePart := source.ProjectOut(normal)

	sineScale := r.IndexOfRefraction
	cosinePart := normal
	if normal.Dot(source) < 0 {
		sineScale = 1 / sineScale
		cosinePart = cosinePart.Scale(-1)
	}

	sinePart = sinePart.Scale(sineScale)
	sineNorm := sinePart.Norm()
	if math.Abs(sineNorm) > 1 {
		// Total internal reflection.
		return normal.Reflect(source).Scale(-1)
	}
	cosinePart = cosinePart.Scale(math.Sqrt(1 - sineNorm*sineNorm))
	return sinePart.Add(cosinePart)
}

func (r *RefractMaterial) refractInverse(normal, dest model3d.Coord3D) model3d.Coord3D {
	return r.refract(normal, dest.Scale(-1)).Scale(-1)
}

func (r *RefractMaterial) BSDF(normal, source, dest model3d.Coord3D) Color {
	return r.RefractColor.Scale(r.bsdfScale(normal, source, dest))
}

func (r *RefractMaterial) bsdfScale(normal, source, dest model3d.Coord3D) float64 {
	refracted := r.refract(normal, source)
	delta := r.cosineDelta(normal, source, dest)
	if dest.Dot(refracted) < 1-delta {
		return 0
	}
	scale := 1 / maximumCosine(source.Dot(normal), dest.Dot(normal))

	// delta/2 is the spanned fraction of the sphere
	// for which we return non-zero.
	return scale * 2 / delta
}

func (r *RefractMaterial) cosineDelta(normal, source, dest model3d.Coord3D) float64 {
	delta := cosineEpsilon

	// Area of the patch is changed for energy focus.
	sourceSign := source.Dot(normal)
	destSign := dest.Dot(normal)
	if sourceSign != destSign {
		// Reflection, no change in delta.
	} else if sourceSign < 0 {
		// Entering solid, area is concentrated.
		delta /= r.IndexOfRefraction
	} else {
		// Leaving solid, area is spread out.
		delta *= r.IndexOfRefraction
	}

	return delta
}

func (r *RefractMaterial) SampleSource(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	// Sample deterministically, since all vectors around
	// this neighborhood have the same BRDF.
	return r.refractInverse(normal, dest)
}

func (r *RefractMaterial) SourceDensity(normal, source, dest model3d.Coord3D) float64 {
	// Get the density, assuming we intended to sample
	// around a small section of source vectors.
	refracted := r.refractInverse(normal, dest)
	delta := r.cosineDelta(normal, source.Scale(-1), dest.Scale(-1))
	if source.Dot(refracted) < 1-delta {
		return 0
	}
	return 2 / delta
}

func (r *RefractMaterial) Emission() Color {
	return Color{}
}

func (r *RefractMaterial) Ambient() Color {
	return Color{}
}

// A JoinedMaterial adds the BSDFs of multiple materials.
//
// It also importance samples from each BSDF according to
// pre-determined probabilities.
type JoinedMaterial struct {
	Materials []Material

	// Probs contains probabilities for importance
	// sampling each material.
	// The probabilities should sum to 1.
	Probs []float64
}

func (j *JoinedMaterial) BSDF(normal, source, dest model3d.Coord3D) Color {
	var res Color
	for _, m := range j.Materials {
		res = res.Add(m.BSDF(normal, source, dest))
	}
	return res
}

func (j *JoinedMaterial) SampleSource(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	if len(j.Probs) != len(j.Materials) {
		panic("mismatched probabilities and materials")
	}
	p := rand.Float64()
	for i, subProb := range j.Probs {
		p -= subProb
		if p < 0 || i == len(j.Probs)-1 {
			return j.Materials[i].SampleSource(gen, normal, dest)
		}
	}
	panic("unreachable")
}

func (j *JoinedMaterial) SourceDensity(normal, source, dest model3d.Coord3D) float64 {
	var density float64
	for i, subProb := range j.Probs {
		density += subProb * j.Materials[i].SourceDensity(normal, source, dest)
	}
	return density
}

func (j *JoinedMaterial) Emission() Color {
	var res Color
	for _, m := range j.Materials {
		res = res.Add(m.Emission())
	}
	return res
}

func (j *JoinedMaterial) Ambient() Color {
	var res Color
	for _, m := range j.Materials {
		res = res.Add(m.Ambient())
	}
	return res
}
