package render3d

import (
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/unixpickle/model3d"
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

	// Cutoff is the maximum brightness for which
	// recursion is performed. If small but non-zero, the
	// number of rays traced can be reduced.
	Cutoff float64

	// Epsilon is a small distance used to move away from
	// surfaces before bouncing new rays.
	// If nil, DefaultEpsilon is used.
	Epsilon float64
}

// Render renders the object to an image.
func (r *RecursiveRayTracer) Render(img *Image, obj Object) {
	if r.NumSamples == 0 {
		panic("must set NumSamples to non-zero for RecursiveRayTracer")
	}
	maxX := float64(img.Width) - 1
	maxY := float64(img.Height) - 1
	caster := r.Camera.Caster(maxX, maxY)

	coords := make(chan [3]int, img.Width*img.Height)
	var idx int
	for y := 0; y < img.Width; y++ {
		for x := 0; x < img.Height; x++ {
			coords <- [3]int{x, y, idx}
			idx++
		}
	}
	close(coords)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ray := model3d.Ray{Origin: r.Camera.Origin}
			for c := range coords {
				ray.Direction = caster(float64(c[0]), float64(c[1]))
				var color Color
				for i := 0; i < r.NumSamples; i++ {
					color = color.Add(r.castRay(obj, &ray, 0, 1))
				}
				img.Data[c[2]] = color.Scale(1 / float64(r.NumSamples))
			}
		}()
	}

	wg.Wait()
}

func (r *RecursiveRayTracer) recurse(obj Object, point model3d.Coord3D, ray *model3d.Ray,
	coll model3d.RayCollision, mat Material, depth int, scale float64) Color {
	dest := ray.Direction.Normalize().Scale(-1)
	color := mat.Emission()
	if depth == 0 {
		// Only add ambient light directly to object, not to
		// recursive rays.
		color = color.Add(mat.Ambient())
	}
	for _, l := range r.Lights {
		lightDirection := l.Origin.Sub(point)

		shadowRay := r.bounceRay(point, lightDirection)
		collision, _, ok := obj.Cast(shadowRay)
		if ok && collision.Scale < 1 {
			continue
		}

		brdf := mat.BSDF(coll.Normal, point.Sub(l.Origin).Normalize(), dest)
		color = color.Add(l.ShadeCollision(coll.Normal, lightDirection).Mul(brdf))
	}
	if depth >= r.MaxDepth {
		return color
	}
	nextSource := r.sampleNextSource(point, coll.Normal, dest, mat)
	weight := 1 / r.sourceDensity(point, coll.Normal, nextSource, dest, mat)
	weight *= math.Abs(nextSource.Dot(coll.Normal))
	reflectWeight := mat.BSDF(coll.Normal, nextSource, dest)
	nextRay := r.bounceRay(point, nextSource.Scale(-1))
	nextScale := scale * reflectWeight.Sum() * weight
	nextColor := r.castRay(obj, nextRay, depth+1, nextScale)
	return color.Add(nextColor.Mul(reflectWeight).Scale(weight))
}

func (r *RecursiveRayTracer) sampleNextSource(point, normal, dest model3d.Coord3D,
	mat Material) model3d.Coord3D {
	if len(r.FocusPoints) == 0 {
		return mat.SampleSource(normal, dest)
	}

	p := rand.Float64()
	for i, prob := range r.FocusPointProbs {
		p -= prob
		if p < 0 {
			return r.FocusPoints[i].SampleFocus(point)
		}
	}

	return mat.SampleSource(normal, dest)
}

func (r *RecursiveRayTracer) sourceDensity(point, normal, source, dest model3d.Coord3D,
	mat Material) float64 {
	if len(r.FocusPoints) == 0 {
		return mat.SourceDensity(normal, source, dest)
	}

	matProb := 1.0
	var prob float64
	for i, focusProb := range r.FocusPointProbs {
		prob += focusProb * r.FocusPoints[i].FocusDensity(point, source)
		matProb -= focusProb
	}

	return prob + matProb*mat.SourceDensity(normal, source, dest)
}

func (r *RecursiveRayTracer) castRay(obj Object, ray *model3d.Ray, depth int, scale float64) Color {
	if scale < r.Cutoff {
		return Color{}
	}
	collision, material, ok := obj.Cast(ray)
	if !ok {
		return Color{}
	}
	point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))
	return r.recurse(obj, point, ray, collision, material, depth, scale)
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
