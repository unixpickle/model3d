package render3d

import (
	"github.com/unixpickle/model3d"
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
	ray := model3d.Ray{Origin: r.Camera.Origin}
	var idx int
	for y := 0.0; y <= maxY; y++ {
		for x := 0.0; x <= maxX; x++ {
			idx++
			ray.Direction = caster(x, y)
			collision, material, ok := obj.Cast(&ray)
			if !ok {
				continue
			}
			point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))
			color := material.Ambience().Add(material.Luminance())
			for _, l := range r.Lights {
				dist := l.Origin.Sub(point).Norm()
				scale := material.Reflect(collision.Normal, point.Sub(l.Origin).Normalize(),
					ray.Origin.Sub(point).Normalize())
				color = color.Add(l.ColorAtDistance(dist).Mul(scale))
			}
			img.Data[idx-1] = color
		}
	}
}
