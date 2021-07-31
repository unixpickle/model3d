package model2d

import (
	"image"
	"image/color"
	"math"
	"os"

	"github.com/pkg/errors"
)

// Interpolator is a 1-dimensional interpolation kernel.
type Interpolator int

const (
	Bicubic Interpolator = iota
	Bilinear
)

func (i Interpolator) Kernel(t float64) []float64 {
	switch i {
	case Bicubic:
		t2 := t * t
		t3 := t2 * t
		return []float64{
			0.5 * (-t + 2*t2 - t3),
			0.5 * (2 - 5*t2 + 3*t3),
			0.5 * (t + 4*t2 - 3*t3),
			0.5 * (-t2 + t3),
		}
	case Bilinear:
		return []float64{1 - t, t}
	default:
		panic("unknown interpolator")
	}
}

// An InterpBitmap is a dynamic Bitmap backed by an image
// with a color interpolation scheme.
type InterpBitmap struct {
	Data   []color.RGBA
	Width  int
	Height int
	Model  color.Model
	F      ColorBitFunc

	// Interp is the interpolation function.
	// A zero value is Bicubic.
	Interp Interpolator
}

// NewInterpBitmap creates a InterpBitmap from an image.
//
// If c is nil, then the mean RGBA is computed, and pixels
// are considered true if they are closer to the mean in
// L2 distance than they are to the top-left pixel.
// For images with two dominant colors, this is equivalent
// to making the background false, and the foreground
// true, assuming that the first pixel is background.
func NewInterpBitmap(img image.Image, c ColorBitFunc) *InterpBitmap {
	if c == nil {
		c = statisticalColorBitFunc(img)
	}
	rgbaModel := color.RGBAModel
	bounds := img.Bounds()
	data := make([]color.RGBA, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			data = append(data, rgbaModel.Convert(img.At(x, y)).(color.RGBA))
		}
	}
	return &InterpBitmap{
		Data:   data,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Model:  img.ColorModel(),
		F:      c,
	}
}

// ReadInterpBitmap is like NewInterpBitmap, except that
// it reads the image from a file.
func ReadInterpBitmap(path string, c ColorBitFunc) (*InterpBitmap, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "read InterpBitmap")
	}
	defer r.Close()
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, errors.Wrap(err, "read InterpBitmap")
	}
	return NewInterpBitmap(img, c), nil
}

// MustReadInterpBitmap is like ReadInterpBitmap, except
// that it panics if the InterpBitmap cannot be read.
func MustReadInterpBitmap(path string, c ColorBitFunc) *InterpBitmap {
	bmp, err := ReadInterpBitmap(path, c)
	if err != nil {
		panic(err)
	}
	return bmp
}

// Bitmap gets an uninterpolated bitmap from b.
func (b *InterpBitmap) Bitmap() *Bitmap {
	res := NewBitmap(b.Width, b.Height)
	for i, x := range b.Data {
		res.Data[i] = b.F(b.Model.Convert(x))
	}
	return res
}

// Get gets the color at the coordinate.
//
// If the coordinate is out of bounds, a the edge of the
// image is extended.
func (b *InterpBitmap) Get(x, y int) color.RGBA {
	if x < 0 {
		x = 0
	} else if x >= b.Width {
		x = b.Width - 1
	}
	if y < 0 {
		y = 0
	} else if y >= b.Height {
		y = b.Height - 1
	}
	return b.Data[x+y*b.Width]
}

// Min gets the minimum of the pixel bounding box.
func (b *InterpBitmap) Min() Coord {
	return Coord{}
}

// Max gets the maximum of the pixel bounding box.
func (b *InterpBitmap) Max() Coord {
	return XY(float64(b.Width), float64(b.Height))
}

// Contains gets the bit at the interpolated coordinate.
//
// If the coordinate is out of bounds, false is returned.
func (b *InterpBitmap) Contains(c Coord) bool {
	if c.X < 0 || c.Y < 0 || c.X >= float64(b.Width) || c.Y >= float64(b.Height) {
		return false
	}

	// Interpolate around the middle of each pixel
	c = c.Sub(XY(0.5, 0.5))

	tx := c.X - math.Floor(c.X)
	ty := c.Y - math.Floor(c.Y)

	xKernel := b.Interp.Kernel(tx)
	yKernel := b.Interp.Kernel(ty)
	startX := int(c.X) - (len(xKernel)/2 - 1)
	startY := int(c.Y) - (len(yKernel)/2 - 1)

	var result [4]float64
	for i, ky := range yKernel {
		for j, kx := range xKernel {
			color := b.Get(j+startX, i+startY)
			comps := [4]float64{float64(color.R), float64(color.G), float64(color.B),
				float64(color.A)}
			for k, comp := range comps {
				result[k] += kx * ky * comp
			}
		}
	}
	for i, x := range result {
		if x < 0 {
			result[i] = 0
		} else if x > 255 {
			result[i] = 255
		}
	}

	rgba := color.RGBA{
		R: uint8(result[0]),
		G: uint8(result[1]),
		B: uint8(result[2]),
		A: uint8(result[3]),
	}

	return b.F(b.Model.Convert(rgba))
}

// FlipX reverses the x-axis.
func (b *InterpBitmap) FlipX() *InterpBitmap {
	res := *b
	res.Data = make([]color.RGBA, len(b.Data))
	for y := 0; y < b.Height; y++ {
		row := y * b.Width
		for x := 0; x < b.Width; x++ {
			res.Data[row+x] = b.Data[row+b.Width-(x+1)]
		}
	}
	return &res
}

// FlipY reverses the y-axis.
func (b *InterpBitmap) FlipY() *InterpBitmap {
	res := *b
	res.Data = make([]color.RGBA, len(b.Data))
	for y := 0; y < b.Height; y++ {
		row := y * b.Width
		row1 := (b.Height - (y + 1)) * b.Width
		copy(res.Data[row:row+b.Width], b.Data[row1:])
	}
	return &res
}

// Invert creates a new InterpBitmap with the opposite
// color bitmap values.
func (b *InterpBitmap) Invert() *InterpBitmap {
	res := *b
	res.F = func(c color.Color) bool {
		return !b.F(c)
	}
	return &res
}
