package render3d

import (
	"github.com/unixpickle/model3d"
)

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
		lightDist := lightDirection.Norm()

		shadowRay := bounceRay(point, lightDirection, lightDist)
		collision, _, ok := obj.Cast(shadowRay)
		if ok && collision.Scale < 1 {
			continue
		}

		scale := mat.Reflect(coll.Normal, point.Sub(l.Origin).Normalize(),
			ray.Origin.Sub(point).Normalize())
		color = color.Add(l.ColorAtDistance(lightDist).Mul(scale))
	}
	if depth >= r.MaxDepth {
		// Only add ambient light directly to object, not to
		// recursive rays.
		return color.Add(mat.Ambience())
	}
	nextDest := ray.Direction.Normalize().Scale(-1)
	nextSource, weight := mat.SampleSource(coll.Normal, nextDest)
	reflectWeight := mat.Reflect(coll.Normal, nextSource, nextDest)
	nextRay := bounceRay(point, nextSource, coll.Scale)
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

func bounceRay(point model3d.Coord3D, newDirection model3d.Coord3D, scale float64) *model3d.Ray {
	return &model3d.Ray{
		// Prevent a duplicate collision from being
		// detected when bouncing off an existing
		// object.
		Origin:    point.Add(newDirection.Scale(scale * 1e-8)),
		Direction: newDirection,
	}
}
