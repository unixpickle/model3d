package render3d

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
)

type Image struct {
	Data   []Color
	Width  int
	Height int
}

// NewImage creates an image at a certain size.
func NewImage(width, height int) *Image {
	return &Image{
		Data:   make([]Color, width*height),
		Width:  width,
		Height: height,
	}
}

// At gets the color at the coordinate.
func (i *Image) At(x, y int) Color {
	if x < 0 || y < 0 || x >= i.Width || y >= i.Height {
		panic("position out of bounds")
	}
	return i.Data[x+y*i.Width]
}

// Set stores a color at the coordinate.
func (i *Image) Set(x, y int, c Color) {
	if x < 0 || y < 0 || x >= i.Width || y >= i.Height {
		panic("position out of bounds")
	}
	i.Data[x+y*i.Width] = c
}

// CopyFrom copies the image img into this image at the
// given coordinates x, y in i.
//
// This can be used to copy a smaller image i1 into a
// larger image i.
func (i *Image) CopyFrom(i1 *Image, x, y int) {
	copyWidth := essentials.MinInt(i1.Width, i.Width-x)
	copyHeight := essentials.MinInt(i1.Height, i.Height-y)
	for row := 0; row < copyHeight; row++ {
		destIdx := (row+y)*i.Width + x
		sourceIdx := row * i1.Width
		copy(i.Data[destIdx:destIdx+copyWidth], i1.Data[sourceIdx:])
	}
}

// FillRange scales the color values so that the largest
// color component is exactly 1.
func (i *Image) FillRange() {
	var max Color
	for _, c := range i.Data {
		max = max.Max(c)
	}
	maximum := max.MaxCoord()
	if maximum <= 0 {
		return
	}
	scale := 1 / maximum
	for j, c := range i.Data {
		i.Data[j] = c.Scale(scale)
	}
}

// Scale scales all colors by a constant.
func (i *Image) Scale(s float64) {
	for j, c := range i.Data {
		i.Data[j] = c.Scale(s)
	}
}

// Downsample returns a new image by downsampling i by an
// integer scale factor. The factor must evenly divide the
// width and height.
//
// New colors are computed by averaging squares of colors
// in linear RGB space.
func (i *Image) Downsample(factor int) *Image {
	if i.Width%factor != 0 || i.Height%factor != 0 {
		panic(fmt.Sprintf("image size %d x %d cannot be divided evenly by factor %d",
			i.Width, i.Height, factor))
	}
	weight := 1.0 / float64(factor*factor)
	out := NewImage(i.Width/factor, i.Height/factor)
	for i1 := 0; i1 < out.Height; i1++ {
		for j := 0; j < out.Width; j++ {
			var sum Color
			for k := 0; k < factor; k++ {
				for l := 0; l < factor; l++ {
					offset := i.Width*(i1*factor+k) + (j*factor + l)
					sum = sum.Add(i.Data[offset])
				}
			}
			out.Data[i1*out.Width+j] = sum.Scale(weight)
		}
	}
	return out
}

// RGBA creates a standard library RGBA image from i.
//
// Values outside the range of [0, 1] are clamped.
func (i *Image) RGBA() *image.RGBA {
	res := image.NewRGBA(image.Rect(0, 0, i.Width, i.Height))

	var idx int
	for y := 0; y < i.Height; y++ {
		for x := 0; x < i.Width; x++ {
			c := ClampColor(i.Data[idx])
			idx++

			r, g, b := RGB(c)
			res.SetRGBA(x, y, color.RGBA{
				R: uint8(r * (256.0 - 0.001)),
				G: uint8(g * (256.0 - 0.001)),
				B: uint8(b * (256.0 - 0.001)),
				A: 0xff,
			})
		}
	}

	return res
}

// Gray creates a standard library Gray image from i.
//
// Values outside the range of [0, 1] are clamped.
func (i *Image) Gray() *image.Gray {
	res := image.NewGray(image.Rect(0, 0, i.Width, i.Height))

	var idx int
	for y := 0; y < i.Height; y++ {
		for x := 0; x < i.Width; x++ {
			c := ClampColor(i.Data[idx])
			idx++

			// Use RGB because not all colors are
			// perceived as equally bright, and the image
			// library knows how to weight them.
			r, g, b := RGB(c)
			res.Set(x, y, color.RGBA{
				R: uint8(r * (256.0 - 0.001)),
				G: uint8(g * (256.0 - 0.001)),
				B: uint8(b * (256.0 - 0.001)),
				A: 0xff,
			})
		}
	}

	return res
}

// Save saves the image to a file.
//
// It uses the extension to determine the type.
// Use either .png, .jpg, or .jpeg.
func (i *Image) Save(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return fmt.Errorf("save image: unknown extension '%s'", ext)
	}
	w, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "save image")
	}
	defer w.Close()
	rgba := i.RGBA()
	if ext == ".png" {
		err = png.Encode(w, rgba)
	} else {
		err = jpeg.Encode(w, rgba, nil)
	}
	if err != nil {
		return errors.Wrap(err, "save image")
	}
	return nil
}
