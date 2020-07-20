package model2d

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

const (
	RasterizerDefaultSubsamples = 8
	RasterizerDefaultLineWidth  = 1.0
)

// A Rasterizer converts 2D models into raster images.
type Rasterizer struct {
	// Scale determines how many pixels comprise a unit
	// distance in the model being rasterized.
	//
	// This determines how large output images are, given
	// the bounds of the model being rasterized.
	//
	// A value of 0 defaults to a value of 1.
	Scale float64

	// Subsamples indicates how many sub-samples to test
	// for each axis in each pixel.
	// A value of 1 means one sample is taken per pixel,
	// and values higher than one cause anti-aliasing.
	// If 0, RasterizerDefaultSubsamples is used.
	Subsamples int

	// LineWidth is the thickness of lines (in pixels)
	// when rendering a mesh or collider.
	//
	// If 0, RasterizerDefaultLineWidth is used.
	LineWidth float64
}

// Rasterize rasterizes a Solid, Mesh, or Collider.
func (r *Rasterizer) Rasterize(obj interface{}) *image.Gray {
	switch obj := obj.(type) {
	case Solid:
		return r.RasterizeSolid(obj)
	case Collider:
		return r.RasterizeCollider(obj)
	case *Mesh:
		return r.RasterizeCollider(MeshToCollider(obj))
	}
	panic(fmt.Sprintf("cannot rasterize objects of type: %T", obj))
}

// RasterizeSolid rasterizes a Solid into an image.
func (r *Rasterizer) RasterizeSolid(s Solid) *image.Gray {
	scale := r.scale()

	min, max := s.Min(), s.Max()
	outWidth := int(math.Ceil((max.X - min.X) * scale))
	outHeight := int(math.Ceil((max.Y - min.Y) * scale))
	out := image.NewGray(image.Rect(0, 0, outWidth, outHeight))

	pixelWidth := (max.X - min.X) / float64(outWidth)
	pixelHeight := (max.Y - min.Y) / float64(outHeight)

	for y := 0; y < outHeight; y++ {
		for x := 0; x < outWidth; x++ {
			pxMin := XY(float64(x)*pixelWidth+min.X, float64(y)*pixelHeight+min.Y)
			pxMax := XY(float64(x+1)*pixelWidth+min.X, float64(y+1)*pixelHeight+min.Y)
			px := 1 - r.rasterizePixel(s, pxMin, pxMax)
			out.Set(x, y, color.Gray{
				Y: uint8(math.Floor(px * 255.999)),
			})
		}
	}

	return out
}

// RasterizeCollider rasterizes the collider as a line
// drawing.
func (r *Rasterizer) RasterizeCollider(c Collider) *image.Gray {
	solid := NewColliderSolidHollow(c, 0.5*r.lineWidth()/r.scale())
	return r.RasterizeSolid(solid)
}

func (r *Rasterizer) rasterizePixel(s Solid, min, max Coord) float64 {
	subsamples := r.subsamples()
	division := max.Sub(min).Scale(1 / float64(subsamples+1))
	var result float64
	for x := 0; x < subsamples; x++ {
		for y := 0; y < subsamples; y++ {
			c := min
			c.X += division.X * float64(x)
			c.Y += division.Y * float64(y)
			if s.Contains(c) {
				result += 1
			}
		}
	}
	return result / float64(subsamples*subsamples)
}

func (r *Rasterizer) scale() float64 {
	if r.Scale == 0 {
		return 1
	}
	return r.Scale
}

func (r *Rasterizer) subsamples() int {
	if r.Subsamples == 0 {
		return RasterizerDefaultSubsamples
	}
	return r.Subsamples
}

func (r *Rasterizer) lineWidth() float64 {
	if r.LineWidth == 0 {
		return RasterizerDefaultLineWidth
	}
	return r.LineWidth
}
