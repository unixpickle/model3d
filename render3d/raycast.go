package render3d

import (
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
)

// A RayCaster renders objects using simple one-step ray
// tracing with no recursion.
type RayCaster struct {
	Camera *Camera
	Lights []*PointLight
}

// Render renders the object to an image.
func (r *RayCaster) Render(img *Image, obj Object) {
	maxX := float64(img.Width) - 1
	maxY := float64(img.Height) - 1
	caster := r.Camera.Caster(maxX, maxY)

	mapCoordinates(img.Width, img.Height, func(gen *rand.Rand, x, y, idx int) {
		ray := model3d.Ray{
			Origin:    r.Camera.Origin,
			Direction: caster(float64(x), float64(y)),
		}
		collision, material, ok := obj.Cast(&ray)
		if !ok {
			return
		}
		point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))
		color := material.Ambient().Add(material.Emission())
		for _, l := range r.Lights {
			brdf := material.BSDF(collision.Normal, point.Sub(l.Origin).Normalize(),
				ray.Origin.Sub(point).Normalize())
			p2l := l.Origin.Sub(point)
			color = color.Add(l.ShadeCollision(collision.Normal, p2l).Mul(brdf))
		}
		img.Data[idx] = color
	})
}
