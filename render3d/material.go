package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
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
	// by the surface.
	Emission() Color

	// Ambient is the baseline color to use for all
	// collisions with this surface for rendering.
	// It ensures that every surface is rendered at least
	// some amount.
	Ambient() Color
}

// AsymMaterial is a specialized Material with a different
// sampling distribution for destination and source
// vectors.
//
// This is useful for transparent objects, but should not
// be used for typical materials.
type AsymMaterial interface {
	Material

	// SampleDest is like SampleSource, but it samples a
	// destination direction.
	SampleDest(gen *rand.Rand, normal, source model3d.Coord3D) model3d.Coord3D

	// DestDensity is like SourceDensity, but for
	// SampleDest rather than SampleSource.
	DestDensity(normal, source, dest model3d.Coord3D) float64
}

// SampleDest is like mat.SampleSource, but it samples a
// destination direction.
//
// If mat is an AsymMaterial, its SampleDest method is
// used.
func SampleDest(mat Material, gen *rand.Rand, normal, source model3d.Coord3D) model3d.Coord3D {
	if a, ok := mat.(AsymMaterial); ok {
		return a.SampleDest(gen, normal, source)
	} else {
		return mat.SampleSource(gen, normal, source.Scale(-1)).Scale(-1)
	}
}

// DestDensity is like mat.SourceDensity, but for
// SampleDest rather than SampleSource.
//
// If mat is an AsymMaterial, its SampleDest method is
// used.
func DestDensity(mat Material, normal, source, dest model3d.Coord3D) float64 {
	if a, ok := mat.(AsymMaterial); ok {
		return a.DestDensity(normal, source, dest)
	} else {
		return mat.SourceDensity(normal, dest.Scale(-1), source.Scale(-1))
	}
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
	// Multiply by another 2 so that the outgoing flux is
	// equal to the incoming flux (the cosine term drops
	// half the light).
	return l.DiffuseColor.Scale(4)
}

func (l *LambertMaterial) SampleSource(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	// Sample with probabilities proportional to the cosine
	// property (Lamert's law).
	u := gen.Float64()
	cosLat := math.Sqrt(u)
	sinLat := math.Sqrt(1 - u)
	lon := gen.Float64() * 2 * math.Pi

	xAxis, zAxis := normal.OrthoBasis()

	lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
	point := normal.Scale(-cosLat).Add(lonPoint.Scale(sinLat))

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
		// See LambertMaterial.BSDF() for scale.
		color = p.DiffuseColor.Scale(4)
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
	if (p.DiffuseColor == Color{}) || gen.Intn(2) == 0 {
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
	cosLat := math.Pow(v, 1/(alpha+1))
	sinLat := math.Sqrt(1 - cosLat*cosLat)

	lonPoint := xAxis.Scale(math.Cos(lon)).Add(zAxis.Scale(math.Sin(lon)))
	return direction.Scale(cosLat).Add(lonPoint.Scale(sinLat))
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
// asymmetric, since energy is concentrated and spread out
// due to refraction.
type RefractMaterial struct {
	// IndexOfRefraction is the index of refraction of
	// this material. Values greater than one simulate
	// materials like water or glass, where light travels
	// more slowly than in space.
	IndexOfRefraction float64

	// RefractColor is the mask used for refracted flux.
	RefractColor Color

	// SpecularColor, if specified, indicates that Fresnel
	// reflection should be used with the given color.
	// Typically, if specified, the color of 1's should be
	// used for a white reflection.
	SpecularColor Color
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

func (r *RefractMaterial) reflectAmount(normal, source model3d.Coord3D) float64 {
	// https://en.wikipedia.org/wiki/Schlick%27s_approximation
	x := (r.IndexOfRefraction - 1) / (r.IndexOfRefraction + 1)
	r0 := x * x
	return r0 * (1 - r0) * math.Pow(1-math.Abs(normal.Dot(source)), 5)
}

func (r *RefractMaterial) BSDF(normal, source, dest model3d.Coord3D) Color {
	if (r.SpecularColor == Color{}) {
		return r.RefractColor.Scale(r.refractBSDF(normal, source, dest))
	}
	reflectAmount := r.reflectAmount(normal, source)
	refract := (1 - reflectAmount) * r.refractBSDF(normal, source, dest)
	reflect := reflectAmount * r.reflectBSDF(normal, source, dest)
	return r.RefractColor.Scale(refract).Add(r.SpecularColor.Scale(reflect))
}

func (r *RefractMaterial) refractBSDF(normal, source, dest model3d.Coord3D) float64 {
	refracted := r.refract(normal, source)
	if dest.Dot(refracted) < 1-cosineEpsilon {
		return 0
	}

	// Correct for change in flux going out vs coming in,
	// making the integral of the outgoing flux approx 1.
	scale := 1 / math.Max(cosineEpsilon, math.Abs(dest.Dot(normal)))

	// eps/2 is the spanned fraction of the sphere
	// for which we return non-zero.
	return scale * 2 / cosineEpsilon
}

func (r *RefractMaterial) reflectBSDF(normal, source, dest model3d.Coord3D) float64 {
	reflected := normal.Reflect(source).Scale(-1)
	if dest.Dot(reflected) < 1-cosineEpsilon {
		return 0
	}
	scale := 1 / maximumCosine(dest.Dot(normal), source.Dot(normal))
	return scale * 2 / cosineEpsilon
}

func (r *RefractMaterial) SampleSource(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	if (r.SpecularColor == Color{}) {
		// Sample deterministically, since all vectors around
		// this neighborhood have the same BSDF.
		return r.refractInverse(normal, dest)
	}

	reflect := r.reflectAmount(normal, dest)
	if gen.Float64() > reflect {
		return r.refractInverse(normal, dest)
	} else {
		return normal.Reflect(dest).Scale(-1)
	}
}

func (r *RefractMaterial) SourceDensity(normal, source, dest model3d.Coord3D) float64 {
	if (r.SpecularColor == Color{}) {
		// Get the density, assuming we intended to sample
		// around a small section of source vectors.
		refracted := r.refractInverse(normal, dest)
		if source.Dot(refracted) < 1-cosineEpsilon {
			return 0
		}
		return 2 / cosineEpsilon
	}
	reflect := r.reflectAmount(normal, dest)
	reflected := normal.Reflect(dest).Scale(-1)
	refracted := r.refractInverse(normal, dest)
	var density float64
	if source.Dot(refracted) >= 1-cosineEpsilon {
		density += 1 - reflect
	}
	if source.Dot(reflected) >= 1-cosineEpsilon {
		density += reflect
	}
	return density * 2 / cosineEpsilon
}

func (r *RefractMaterial) SampleDest(gen *rand.Rand, normal,
	source model3d.Coord3D) model3d.Coord3D {
	return r.SampleSource(gen, normal.Scale(-1), source)
}

func (r *RefractMaterial) DestDensity(normal, source, dest model3d.Coord3D) float64 {
	return r.SourceDensity(normal.Scale(-1), dest, source)
}

func (r *RefractMaterial) Emission() Color {
	return Color{}
}

func (r *RefractMaterial) Ambient() Color {
	return Color{}
}

// HGMaterial implements the Henyey-Greenstein phase
// function for ray scattering.
//
// This material cancels out the normal cosine term that
// is introduced by the rendering equation.
// To ignore normals and not cancel, set IgnoreNormals.
type HGMaterial struct {
	// G is a control parameter in [-1, 1].
	// -1 is backscattering, 1 is forward scattering.
	G float64

	// ScatterColor controls how much light is actually
	// scattered vs. absorbed.
	ScatterColor Color

	IgnoreNormals bool
}

func (h *HGMaterial) BSDF(normal, source, dest model3d.Coord3D) Color {
	density := h.cosDensity(source.Dot(dest))
	if !h.IgnoreNormals {
		density /= math.Max(1e-5, math.Abs(source.Dot(normal)))
	}
	return h.ScatterColor.Scale(density)
}

func (h *HGMaterial) SampleSource(gen *rand.Rand, normal, dest model3d.Coord3D) model3d.Coord3D {
	// Based on notes from https://www.astro.umd.edu/~jph/HG_note.pdf.
	s := gen.Float64()*2 - 1
	g := h.numericalG()
	g2 := g * g
	powTerm := (1 - g2) / (1 + g*s)

	cosTheta := (1 + g2 - powTerm*powTerm) / (2 * g)
	sinTheta := math.Sqrt(1 - cosTheta*cosTheta)
	alpha := gen.Float64() * math.Pi * 2

	b1, b2 := dest.OrthoBasis()
	ortho := b1.Scale(math.Cos(alpha)).Add(b2.Scale(math.Sin(alpha)))
	return dest.Scale(cosTheta).Add(ortho.Scale(sinTheta))
}

func (h *HGMaterial) SourceDensity(normal, source, dest model3d.Coord3D) float64 {
	return h.cosDensity(source.Dot(dest))
}

func (h *HGMaterial) cosDensity(cos float64) float64 {
	g := h.numericalG()
	// Based on notes from https://www.astro.umd.edu/~jph/HG_note.pdf.
	g2 := g * g
	divisor := 1 + g2 - 2*g*cos
	return (1 - g2) / math.Pow(divisor, 3.0/2.0)
}

func (h *HGMaterial) Emission() Color {
	return Color{}
}

func (h *HGMaterial) Ambient() Color {
	return Color{}
}

func (h *HGMaterial) numericalG() float64 {
	if math.Abs(h.G) < 1e-5 {
		return 1e-5
	}
	return math.Max(math.Min(h.G, 1-1e-5), -(1 - 1e-5))
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
	p := gen.Float64()
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

func (j *JoinedMaterial) SampleDest(gen *rand.Rand, normal,
	dest model3d.Coord3D) model3d.Coord3D {
	if len(j.Probs) != len(j.Materials) {
		panic("mismatched probabilities and materials")
	}
	p := gen.Float64()
	for i, subProb := range j.Probs {
		p -= subProb
		if p < 0 || i == len(j.Probs)-1 {
			return SampleDest(j.Materials[i], gen, normal, dest)
		}
	}
	panic("unreachable")
}

func (j *JoinedMaterial) DestDensity(normal, source, dest model3d.Coord3D) float64 {
	var density float64
	for i, subProb := range j.Probs {
		density += subProb * DestDensity(j.Materials[i], normal, source, dest)
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
