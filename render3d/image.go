package render3d

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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

// FillRange scales the color values so that the largest
// color component is exactly 1.
func (i *Image) FillRange() {
	var max Color
	for _, c := range i.Data {
		max = max.Max(c)
	}
	maximum := math.Max(math.Max(max.X, max.Y), max.Z)
	if maximum <= 0 {
		return
	}
	scale := 1 / maximum
	for j, c := range i.Data {
		i.Data[j] = c.Scale(scale)
	}
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

			res.SetRGBA(x, y, color.RGBA{
				R: uint8(c.X * (256.0 - 0.001)),
				G: uint8(c.Y * (256.0 - 0.001)),
				B: uint8(c.Z * (256.0 - 0.001)),
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

			res.SetGray(x, y, color.Gray{
				Y: uint8((c.X + c.Y + c.Z) * (256.0/3.0 - 0.001)),
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
