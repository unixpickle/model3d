package model3d

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/pkg/errors"
)

// RenderRayCast renders a Collider as a grayscale image
// using a simple ray casting algorithm.
// Rays are shot out from the origin in the z direction.
// They go up to fov radians along the x and y directions
// to produce the final image.
func RenderRayCast(c Collider, output *image.Gray, origin, x, y, z Coord3D, fov float64) {
	rayCastBounds(c, output.Bounds(), origin, x, y, z, fov,
		func(x, y int, brightness float64, c Coord3D) {
			if brightness == 0 {
				output.SetGray(x, y, color.Gray{Y: 0})
			} else {
				grayness := uint8(math.Round(brightness * 0xff))
				output.SetGray(x, y, color.Gray{Y: grayness})
			}
		})
}

// RenderRayCastColor is like RenderRayCast, but it uses a
// color function to decide the color of each pixel.
func RenderRayCastColor(c Collider, output *image.RGBA, origin, x, y, z Coord3D, fov float64,
	f func(c Coord3D) [3]float64) {
	rayCastBounds(c, output.Bounds(), origin, x, y, z, fov,
		func(x, y int, brightness float64, c Coord3D) {
			if brightness == 0 {
				output.SetRGBA(x, y, color.RGBA{A: 0xff})
			} else {
				rgb := f(c)
				output.SetRGBA(x, y, color.RGBA{
					R: uint8(math.Round(rgb[0] * brightness * 0xff)),
					G: uint8(math.Round(rgb[1] * brightness * 0xff)),
					B: uint8(math.Round(rgb[2] * brightness * 0xff)),
					A: 0xff,
				})
			}
		})
}

func rayCastBounds(c Collider, bounds image.Rectangle, origin, x, y, z Coord3D, fov float64,
	f func(x, y int, brightness float64, c Coord3D)) {
	planeDistance := 1 / math.Tan(fov/2)

	x = x.Normalize()
	y = y.Normalize()
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
			coll, collides := c.FirstRayCollision(ray)
			if collides {
				brightness := math.Max(0, -coll.Normal.Dot(z)/z.Norm())
				p := ray.Origin.Add(ray.Direction.Scale(coll.Scale))
				f(j+bounds.Min.X, i+bounds.Min.Y, brightness, p)
			} else {
				f(j+bounds.Min.X, i+bounds.Min.Y, 0, Coord3D{})
			}
		}
	}
}

// RenderRandomGrid renders a collider from various random
// angles and saves the images into a grid.
func RenderRandomGrid(c Collider, rows, cols, thumbWidth, thumbHeight int) *image.Gray {
	output := image.NewGray(image.Rect(0, 0, cols*thumbWidth, rows*thumbHeight))
	iterateRandomGrid(c, rows, cols, thumbWidth, thumbHeight,
		func(rect image.Rectangle, origin, x, y, z Coord3D) {
			RenderRayCast(c, output.SubImage(rect).(*image.Gray), origin, x, y, z, math.Pi/2)
		})
	return output
}

// RenderRandomGridColor is like RenderRandomGrid, but it
// uses a color function to decide the color of each
// pixel.
func RenderRandomGridColor(c Collider, rows, cols, thumbWidth, thumbHeight int,
	f func(c Coord3D) [3]float64) *image.RGBA {
	output := image.NewRGBA(image.Rect(0, 0, cols*thumbWidth, rows*thumbHeight))
	iterateRandomGrid(c, rows, cols, thumbWidth, thumbHeight,
		func(rect image.Rectangle, origin, x, y, z Coord3D) {
			subImage := output.SubImage(rect).(*image.RGBA)
			RenderRayCastColor(c, subImage, origin, x, y, z, math.Pi/2, f)
		})
	return output
}

func iterateRandomGrid(c Collider, rows, cols, thumbWidth, thumbHeight int,
	f func(rect image.Rectangle, origin, x, y, z Coord3D)) {
	min := c.Min()
	max := c.Max()
	center := max.Add(min).Scale(0.5)
	diff := max.Sub(min)
	radius := math.Sqrt(3) * math.Max(diff.X, math.Max(diff.Y, diff.Z)) / 2

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
			y := (GeoCoord{Lat: geo.Lat - 1e-4, Lon: geo.Lon}).Coord3D().Add(z)

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
			f(rect, origin, x, y, z)
		}
	}
}

// SaveRandomGrid is like RenderRandomGrid, except that it
// saves the result to a PNG file.
func SaveRandomGrid(outFile string, c Collider, rows, cols, thumbWidth, thumbHeight int) error {
	return saveRandomGrid(outFile, RenderRandomGrid(c, rows, cols, thumbWidth, thumbHeight))
}

// SaveRandomGridColor is like RenderRandomGridColor,
// except that it saves the result to a PNG file.
func SaveRandomGridColor(outFile string, c Collider, rows, cols, thumbWidth, thumbHeight int,
	f func(c Coord3D) [3]float64) error {
	return saveRandomGrid(outFile, RenderRandomGridColor(c, rows, cols, thumbWidth, thumbHeight, f))
}

func saveRandomGrid(outFile string, img image.Image) error {
	f, err := os.Create(outFile)
	if err != nil {
		return errors.Wrap(err, "save random grid")
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return errors.Wrap(err, "save random grid")
	}
	return nil
}
