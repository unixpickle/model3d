package render3d

import (
	"github.com/unixpickle/model3d"
)

const DefaultEpsilon = 1e-8

// A RecursiveRayTracer renders objects using recursive
// tracing with random sampling.
type RecursiveRayTracer struct {
	Camera *Camera
	Lights []*PointLight

	// MaxDepth is the maximum number of recursions.
	// Setting to 0 is almost equivalent to RayCast, but
	// the ray tracer still checks for shadows.
	MaxDepth int

	// NumSamples is the number of rays to sample.
	NumSamples int

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
	ray := model3d.Ray{Origin: r.Camera.Origin}
	var idx int
	for y := 0.0; y <= maxY; y++ {
		for x := 0.0; x <= maxX; x++ {
			ray.Direction = caster(x, y)
			var color Color
			for i := 0; i < r.NumSamples; i++ {
				color = color.Add(r.castRay(obj, &ray, 0))
			}
			img.Data[idx] = color.Scale(1 / float64(r.NumSamples))
			idx++
		}
	}
}

func (r *RecursiveRayTracer) recurse(obj Object, point model3d.Coord3D, ray *model3d.Ray,
	coll model3d.RayCollision, mat Material, depth int) Color {
	color := mat.Luminance()
	for _, l := range r.Lights {
		lightDirection := l.Origin.Sub(point)

		shadowRay := r.bounceRay(point, lightDirection)
		collision, _, ok := obj.Cast(shadowRay)
		if ok && collision.Scale < 1 {
			continue
		}

		scale := mat.Reflect(coll.Normal, point.Sub(l.Origin).Normalize(),
			ray.Origin.Sub(point).Normalize())
		lightDist := lightDirection.Norm()
		color = color.Add(l.ColorAtDistance(lightDist).Mul(scale))
	}
	if depth >= r.MaxDepth {
		// Only add ambient light directly to object, not to
		// recursive rays.
		return color.Add(mat.Ambience())
	}
	nextDest := ray.Direction.Normalize().Scale(-1)
	nextSource := mat.SampleSource(coll.Normal, nextDest)
	weight := 1 / mat.SourceDensity(coll.Normal, nextSource, nextDest)
	reflectWeight := mat.Reflect(coll.Normal, nextSource, nextDest)
	nextRay := r.bounceRay(point, nextSource.Scale(-1))
	nextColor := r.castRay(obj, nextRay, depth+1)
	return color.Add(nextColor.Mul(reflectWeight).Scale(weight))
}

func (r *RecursiveRayTracer) castRay(obj Object, ray *model3d.Ray, depth int) Color {
	collision, material, ok := obj.Cast(ray)
	if !ok {
		return Color{}
	}
	point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))
	return r.recurse(obj, point, ray, collision, material, depth)
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
