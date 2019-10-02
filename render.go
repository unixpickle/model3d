package model3d

import (
	"image"
	"image/color"
	"math"
)

// RenderRayCast renders a Collider as a grayscale image
// using a simple ray casting algorithm.
// Rays are shot out from the origin in the z direction.
// They go up to fov radians along the x and y directions
// to produce the final image.
func RenderRayCast(c Collider, output *image.Gray, origin, x, y, z Coord3D, fov float64) {
	bounds := output.Bounds()

	planeDistance := 1 / math.Tan(fov/2)

	x = x.Scale(1 / x.Norm())
	y = y.Scale(1 / y.Norm())
	if bounds.Dx() > bounds.Dy() {
		y = y.Scale(float64(bounds.Dy()) / float64(bounds.Dx()))
	} else {
		x = x.Scale(float64(bounds.Dx()) / float64(bounds.Dy()))
	}
	z = z.Scale(planeDistance / z.Norm())

	for i := 0; i < bounds.Dy(); i++ {
		scaledY := y.Scale(float64(2*i)/float64((bounds.Dy()-1)) - 1)
		for j := 0; j < bounds.Dx(); j++ {
			scaledX := x.Scale(float64(2*j)/float64((bounds.Dx()-1)) - 1)
			ray := &Ray{Origin: origin, Direction: scaledX.Add(scaledY).Add(z)}
			collides, _, normal := c.FirstRayCollision(ray)
			if collides {
				// Light source comes from camera.
				brightness := math.Max(0, -normal.Dot(ray.Direction)/ray.Direction.Norm())
				grayness := uint8(math.Round(brightness * 0xff))
				output.SetGray(j+bounds.Min.X, i+bounds.Min.Y, color.Gray{Y: grayness})
			} else {
				output.SetGray(j+bounds.Min.X, i+bounds.Min.Y, color.Gray{Y: 0})
			}
		}
	}
}
