package toolbox3d

import (
	"image"
	"image/color"
	"math"

	"github.com/unixpickle/model3d"
)

// An Equirect is an equirectangular bitmap representing
// colors on a sphere.
//
// It can be used, for example, to aid in implementing a
// 3D polar function.
type Equirect struct {
	img    image.Image
	width  float64
	height float64
}

// NewEquirect creates an Equirect from an image. It is
// assumed that the top of the image is north (positive
// latitude), the bottom of the image is south, the left is
// west (negative longitude), the right is east (positive
// longitude).
func NewEquirect(img image.Image) *Equirect {
	return &Equirect{
		img:    img,
		width:  float64(img.Bounds().Dx()),
		height: float64(img.Bounds().Dy()),
	}
}

// At gets the color at the given GeoCoord.
func (e *Equirect) At(g model3d.GeoCoord) color.Color {
	g = g.Normalize()
	x := math.Round((e.width - 1) * (g.Lon + math.Pi) / (2 * math.Pi))
	y := math.Round((e.height - 1) * (-g.Lat + math.Pi/2) / math.Pi)
	return e.img.At(int(x), int(y))
}
