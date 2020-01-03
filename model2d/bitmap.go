package model2d

import (
	"image"
	"image/color"
	"os"

	"github.com/pkg/errors"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// ColorBitFunc turns colors into single bits.
type ColorBitFunc func(c color.Color) bool

// A Bitmap is a two-dimensional image with boolean
// values.
// The data is stored in row-major order.
type Bitmap struct {
	Data   []bool
	Width  int
	Height int
}

// NewBitmap creates an empty bitmap.
func NewBitmap(width, height int) *Bitmap {
	return &Bitmap{
		Data:   make([]bool, width*height),
		Width:  width,
		Height: height,
	}
}

// NewBitmapImage creates a Bitmap from an image, by
// calling c for each pixel and using the result as the
// bit.
//
// If c is nil, then the mean RGBA is computed, and pixels
// are considered true if they are closer to the mean in
// L2 distance than they are to the top-left pixel.
// For images with two dominant colors, this is equivalent
// to making the background false, and the foreground
// true, assuming that the first pixel is background.
func NewBitmapImage(img image.Image, c ColorBitFunc) *Bitmap {
	if c == nil {
		c = statisticalColorBitFunc(img)
	}

	b := img.Bounds()
	res := NewBitmap(b.Dx(), b.Dy())

	var idx int
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			res.Data[idx] = c(img.At(x, y))
			idx++
		}
	}

	return res
}

// ReadBitmap is like NewBitmapImage, except that it reads
// the image from a file.
func ReadBitmap(path string, c ColorBitFunc) (*Bitmap, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "read bitmap")
	}
	defer r.Close()
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, errors.Wrap(err, "read bitmap")
	}
	return NewBitmapImage(img, c), nil
}

func (b *Bitmap) Get(x, y int) bool {
	if x < 0 || y < 0 || x >= b.Width || y >= b.Height {
		return false
	}
	return b.Data[x+y*b.Width]
}

func (b *Bitmap) Set(x, y int, v bool) {
	if x < 0 || y < 0 || x >= b.Width || y >= b.Height {
		panic("coordinate out of bounds")
	}
	b.Data[x+y*b.Width] = v
}

func statisticalColorBitFunc(img image.Image) ColorBitFunc {
	var mean [4]float64
	var first [4]float64
	var count int
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			for i, comp := range []uint32{r, g, b, a} {
				mean[i] += float64(comp)
			}
			if count == 0 {
				first = mean
			}
			count++
		}
	}

	for i := range mean {
		mean[i] /= float64(count)
	}

	return func(c color.Color) bool {
		var meanDist float64
		var firstDist float64
		r, g, b, a := c.RGBA()
		for i, comp := range []uint32{r, g, b, a} {
			meanDist += squareDist(comp, mean[i])
			firstDist += squareDist(comp, first[i])
		}
		return firstDist < meanDist
	}
}

func squareDist(x uint32, y float64) float64 {
	d := float64(x) - y
	return d * d
}
