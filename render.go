package model3d

import (
	"image"
	"image/color"
	"math"
	"math/rand"
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
				// Light source goes in the direction the camera is
				// looking at, but it is not a point-light.
				brightness := math.Max(0, -normal.Dot(z)/z.Norm())
				grayness := uint8(math.Round(brightness * 0xff))
				output.SetGray(j+bounds.Min.X, i+bounds.Min.Y, color.Gray{Y: grayness})
			} else {
				output.SetGray(j+bounds.Min.X, i+bounds.Min.Y, color.Gray{Y: 0})
			}
		}
	}
}

// RenderRandomGrid renders a collider from various random
// angles and saves the images into a grid.
func RenderRandomGrid(c Collider, rows, cols, thumbWidth, thumbHeight int) *image.Gray {
	min := c.Min()
	max := c.Max()
	center := max.Add(min).Scale(0.5)
	diff := max.Add(min.Scale(-1))
	radius := math.Sqrt(3) * math.Max(diff.X, math.Max(diff.Y, diff.Z)) / 2

	output := image.NewGray(image.Rect(0, 0, cols*thumbWidth, rows*thumbHeight))

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			z := (Coord3D{
				X: rand.NormFloat64(),
				Y: rand.NormFloat64(),
				Z: rand.NormFloat64(),
			}).Normalize()
			origin := center.Add(z.Scale(-radius * math.Sqrt2))

			// Use the derivatives with respect to lat/lon
			// as the y and x axis.
			geo := z.Scale(-1).Geo()
			x := (GeoCoord{Lat: geo.Lat, Lon: geo.Lon + 1e-4}).Coord3D().Add(z)
			y := (GeoCoord{Lat: geo.Lat + 1e-4, Lon: geo.Lon}).Coord3D().Add(z)

			// Ensure that we get non-nan axes which are
			// truly orthogonal, even for the poles.
			x.X += 1e-8
			y.X += 1e-8
			x = x.Normalize()
			y = y.Normalize()
			x = x.Add(z.Scale(-x.Dot(z))).Normalize()
			x = x.Normalize()
			y = y.Add(z.Scale(-y.Dot(z)))
			y = y.Add(x.Scale(-y.Dot(x)))

			rect := image.Rect(j*thumbWidth, i*thumbHeight, (j+1)*thumbWidth, (i+1)*thumbHeight)
			RenderRayCast(c, output.SubImage(rect).(*image.Gray), origin, x, y, z, math.Pi/2)
		}
	}

	return output
}
