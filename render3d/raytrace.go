package render3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
)

const DefaultEpsilon = 1e-8

// A RecursiveRayTracer renders objects using recursive
// tracing with random sampling.
type RecursiveRayTracer struct {
	Camera *Camera
	Lights []*PointLight

	// FocusPoints are functions which cause rays to
	// bounce more in certain directions, with the aim of
	// reducing variance with no bias.
	FocusPoints []FocusPoint

	// FocusPointProbs stores, for each FocusPoint, the
	// probability that this focus point is used to sample
	// a ray (rather than the BRDF).
	FocusPointProbs []float64

	// MaxDepth is the maximum number of recursions.
	// Setting to 0 is almost equivalent to RayCast, but
	// the ray tracer still checks for shadows.
	MaxDepth int

	// NumSamples is the number of rays to sample.
	NumSamples int

	// MinSamples and MaxStddev control early stopping for
	// pixel sampling. If they are both non-zero, then
	// MinSamples rays are sampled, and then more rays are
	// sampled until the pixel standard deviation goes
	// below MaxStddev, or NumSamples samples are taken.
	//
	// Additionally, see Convergence for to customize this
	// stopping criterion.
	MinSamples int
	MaxStddev  float64

	// OversaturatedStddevs controls how few samples are
	// taken at bright parts of the scene.
	//
	// If specified, a pixel may stop being sampled after
	// MinSamples samples if the brightness of that pixel
	// is more than OversaturatedStddevs standard
	// deviations above the maximum brightness (1.0).
	//
	// This can override MaxStddev, since bright parts of
	// the image may have high standard deviations despite
	// having uninteresting specific values.
	OversaturatedStddevs float64

	// Convergence implements a custom convergence
	// criterion.
	//
	// If specified, MaxStddev and OversaturatedStddevs
	// are not used.
	//
	// Convergence is called with the current mean and
	// variance of a pixel, and a return value of true
	// indicates that the ray has converged.
	Convergence func(mean, stddev Color) bool

	// Cutoff is the maximum brightness for which
	// recursion is performed. If small but non-zero, the
	// number of rays traced can be reduced.
	Cutoff float64

	// Antialias, if non-zero, specifies a fraction of a
	// pixel to perturb every ray's origin.
	// Thus, 1 is maximum, and 0 means no change.
	Antialias float64

	// Epsilon is a small distance used to move away from
	// surfaces before bouncing new rays.
	// If nil, DefaultEpsilon is used.
	Epsilon float64

	// LogFunc, if specified, is called periodically with
	// progress information.
	//
	// The frac argument specifies the fraction of pixels
	// which have been colored.
	//
	// The sampleRate argument specifies the mean number
	// of rays traced per pixel.
	LogFunc func(frac float64, sampleRate float64)
}

// Render renders the object to an image.
func (r *RecursiveRayTracer) Render(img *Image, obj Object) {
	r.rayRenderer().Render(img, obj)
}

// RayVariance estimates the variance of the color
// components in the rendered image for a single ray path.
// It is intended to be used to quickly judge how well
// importance sampling is working.
//
// The variance is averaged over every color component in
// the image.
func (r *RecursiveRayTracer) RayVariance(obj Object, width, height, samples int) float64 {
	return r.rayRenderer().RayVariance(obj, width, height, samples)
}

func (r *RecursiveRayTracer) rayRenderer() *rayRenderer {
	return &rayRenderer{
		RayColor: func(g *goInfo, obj Object, ray *model3d.Ray) Color {
			return r.recurse(g.Gen, obj, ray, 0, NewColor(1))
		},

		Camera:               r.Camera,
		NumSamples:           r.NumSamples,
		MinSamples:           r.MinSamples,
		MaxStddev:            r.MaxStddev,
		OversaturatedStddevs: r.OversaturatedStddevs,
		Convergence:          r.Convergence,
		Antialias:            r.Antialias,
		LogFunc:              r.LogFunc,
	}
}

func (r *RecursiveRayTracer) recurse(gen *rand.Rand, obj Object, ray *model3d.Ray,
	depth int, scale Color) Color {
	if scale.Sum()/3 < r.Cutoff {
		return Color{}
	}
	collision, material, ok := obj.Cast(ray)
	if !ok {
		return Color{}
	}
	point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))

	dest := ray.Direction.Normalize().Scale(-1)
	color := material.Emission()
	if depth == 0 {
		// Only add ambient light directly to object, not to
		// recursive rays.
		color = color.Add(material.Ambient())
	}
	for _, l := range r.Lights {
		lightDirection := l.Origin.Sub(point)

		shadowRay := r.bounceRay(point, lightDirection)
		shadowCollision, _, ok := obj.Cast(shadowRay)
		if ok && shadowCollision.Scale < 1 {
			continue
		}

		brdf := material.BSDF(collision.Normal, point.Sub(l.Origin).Normalize(), dest)
		color = color.Add(l.ShadeCollision(collision.Normal, lightDirection).Mul(brdf))
	}
	if depth >= r.MaxDepth {
		return color
	}
	nextSource := r.sampleNextSource(gen, point, collision.Normal, dest, material)
	weight := 1 / r.sourceDensity(point, collision.Normal, nextSource, dest, material)
	weight *= math.Abs(nextSource.Dot(collision.Normal))
	reflectWeight := material.BSDF(collision.Normal, nextSource, dest)
	nextRay := r.bounceRay(point, nextSource.Scale(-1))
	nextMask := reflectWeight.Scale(weight)
	nextScale := scale.Mul(nextMask)
	nextColor := r.recurse(gen, obj, nextRay, depth+1, nextScale)
	return color.Add(nextColor.Mul(nextMask))
}

func (r *RecursiveRayTracer) sampleNextSource(gen *rand.Rand, point, normal, dest model3d.Coord3D,
	mat Material) model3d.Coord3D {
	if len(r.FocusPoints) == 0 {
		return mat.SampleSource(gen, normal, dest)
	} else if len(r.FocusPoints) != len(r.FocusPointProbs) {
		panic("FocusPoints and FocusPointProbs must match in length")
	}

	p := gen.Float64()
	for i, prob := range r.FocusPointProbs {
		p -= prob
		if p < 0 {
			return r.FocusPoints[i].SampleFocus(gen, mat, point, normal, dest)
		}
	}

	return mat.SampleSource(gen, normal, dest)
}

func (r *RecursiveRayTracer) sourceDensity(point, normal, source, dest model3d.Coord3D,
	mat Material) float64 {
	if len(r.FocusPoints) == 0 {
		return mat.SourceDensity(normal, source, dest)
	}

	matProb := 1.0
	var prob float64
	for i, focusProb := range r.FocusPointProbs {
		prob += focusProb * r.FocusPoints[i].FocusDensity(mat, point, normal, source, dest)
		matProb -= focusProb
	}

	return prob + matProb*mat.SourceDensity(normal, source, dest)
}

func (r *RecursiveRayTracer) bounceRay(point model3d.Coord3D, dir model3d.Coord3D) *model3d.Ray {
	eps := r.Epsilon
	if eps == 0 {
		eps = DefaultEpsilon
	}
	return &model3d.Ray{
		// Prevent a duplicate collision from being
		// detected when bouncing off an existing
		// object.
		Origin:    point.Add(dir.Normalize().Scale(eps)),
		Direction: dir,
	}
}
