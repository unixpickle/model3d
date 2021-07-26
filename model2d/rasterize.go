package model2d

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
)

const (
	RasterizerDefaultSubsamples = 8
	RasterizerDefaultLineWidth  = 1.0
)

// Rasterize renders a Solid, Collider, or Mesh to an
// image file.
//
// The bounds of the object being rendered are scaled by
// the provided scale factor to convert to pixel
// coordinates.
//
// This uses the default rasterization settings, such as
// the default line width and anti-aliasing settings.
// To change this, use a Rasterizer object directly.
func Rasterize(path string, obj interface{}, scale float64) error {
	rast := Rasterizer{Scale: scale}
	img := rast.Rasterize(obj)
	if err := SaveImage(path, img); err != nil {
		return errors.Wrap(err, "rasterize image")
	}
	return nil
}

// RasterizeColor is like Rasterize, but it renders
// multiple objects in different colors.
func RasterizeColor(path string, objs []interface{}, colors []color.Color, scale float64) error {
	b0 := objs[0].(Bounder)
	min, max := b0.Min(), b0.Max()
	for _, obj := range objs {
		b := obj.(Bounder)
		min = min.Min(b.Min())
		max = max.Max(b.Max())
	}

	rast := Rasterizer{Scale: scale, Bounds: NewRect(min, max)}
	imgs := make([]*image.Gray, len(objs))
	for i, obj := range objs {
		imgs[i] = rast.Rasterize(obj)
	}
	img := ColorizeOverlay(imgs, colors)
	if err := SaveImage(path, img); err != nil {
		return errors.Wrap(err, "rasterize image")
	}
	return nil
}

// SaveImage saves a rasterized image to a file, inferring
// the file type from the extension.
func SaveImage(path string, img image.Image) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return fmt.Errorf("save image: unknown extension: %s", filepath.Ext(path))
	}

	w, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "save image")
	}

	if ext == ".png" {
		err = png.Encode(w, img)
	} else {
		err = jpeg.Encode(w, img, nil)
	}

	if err == nil {
		err = w.Close()
	} else {
		w.Close()
	}

	if err != nil {
		return errors.Wrap(err, "save image")
	}
	return nil
}

// Colorize turns a grayscale image into a color image
// with an alpha channel.
//
// It is assumed that black (0) is "positive" while white
// (0xff) is negative.
func Colorize(g *image.Gray, co color.Color) *image.RGBA {
	intr, intg, intb, inta := color.RGBAModel.Convert(co).RGBA()
	red, green, blue, alpha := float64(intr)/0x100, float64(intg)/0x100, float64(intb)/0x100,
		float64(inta)/0x100
	img := image.NewRGBA(g.Bounds())
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := g.GrayAt(x, y)
			frac := float64(0xff-gray.Y) / 0xff
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(red * frac),
				G: uint8(green * frac),
				B: uint8(blue * frac),
				A: uint8(alpha * frac),
			})
		}
	}
	return img
}

// ColorizeOverlay turns a series of grayscale images into corresponding
// colors and then overlays them, each on top of the last.
//
// All images must have the same bounds.
func ColorizeOverlay(gs []*image.Gray, cs []color.Color) *image.RGBA {
	if len(gs) != len(cs) {
		panic("images and colors must have same length")
	}
	for i := 1; i < len(gs); i++ {
		if gs[i].Bounds() != gs[0].Bounds() {
			panic("all input images must have the same bounds")
		}
	}
	res := image.NewRGBA(gs[0].Bounds())
	for i, g := range gs {
		draw.Draw(res, g.Bounds(), Colorize(g, cs[i]), image.Point{}, draw.Over)
	}
	return res
}

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

	// Bounds, if non-nil, is used to override the bounds
	// of any rasterized object.
	// This can be used to add padding, or have a
	// consistent canvas when drawing a moving scene.
	Bounds Bounder
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

	min, max := r.bounds(s)
	outWidth := int(math.Ceil((max.X - min.X) * scale))
	outHeight := int(math.Ceil((max.Y - min.Y) * scale))
	out := image.NewGray(image.Rect(0, 0, outWidth, outHeight))

	pixelWidth := (max.X - min.X) / float64(outWidth)
	pixelHeight := (max.Y - min.Y) / float64(outHeight)

	indices := make([][2]int, 0, outWidth*outHeight)

	for y := 0; y < outHeight; y++ {
		for x := 0; x < outWidth; x++ {
			indices = append(indices, [2]int{x, y})
		}
	}

	essentials.ConcurrentMap(0, len(indices), func(i int) {
		x, y := indices[i][0], indices[i][1]
		pxMin := XY(float64(x)*pixelWidth+min.X, float64(y)*pixelHeight+min.Y)
		pxMax := XY(float64(x+1)*pixelWidth+min.X, float64(y+1)*pixelHeight+min.Y)
		px := 1 - r.rasterizePixel(s, pxMin, pxMax)
		out.Set(x, y, color.Gray{
			Y: uint8(math.Floor(px * 255.999)),
		})
	})

	return out
}

// RasterizeSolidFilter rasterizes a Solid using a
// heuristic filter than can eliminate the need to render
// blank regions of the image.
//
// If f returns false for a given rectangular region, it
// means that the solid is definitely uniform within the
// region (i.e. there is no boundary in the region).
// The exact pattern with which f is called will depend on
// the image and rasterization parameters.
func (r *Rasterizer) RasterizeSolidFilter(s Solid, f func(r *Rect) bool) *image.Gray {
	scale := r.scale()

	min, max := r.bounds(s)
	outWidth := int(math.Ceil((max.X - min.X) * scale))
	outHeight := int(math.Ceil((max.Y - min.Y) * scale))
	out := image.NewGray(image.Rect(0, 0, outWidth, outHeight))

	pixelWidth := (max.X - min.X) / float64(outWidth)
	pixelHeight := (max.Y - min.Y) / float64(outHeight)

	indices := make([][2]int, 0, outWidth*outHeight)
	filterSize := essentials.MaxInt(1, 16/r.subsamples())
	for y := 0; y < outHeight; y += filterSize {
		for x := 0; x < outWidth; x += filterSize {
			nextX := essentials.MinInt(outWidth, x+filterSize)
			nextY := essentials.MinInt(outHeight, y+filterSize)
			bounds := &Rect{
				MinVal: XY(float64(x)*pixelWidth+min.X, float64(y)*pixelHeight+min.Y),
				MaxVal: XY(float64(nextX)*pixelWidth+min.X, float64(nextY)*pixelHeight+min.Y),
			}
			shouldRender := f(bounds)
			for subY := y; subY < nextY; subY++ {
				for subX := x; subX < nextX; subX++ {
					if shouldRender {
						indices = append(indices, [2]int{subX, subY})
					} else {
						if s.Contains(bounds.MinVal.Mid(bounds.MaxVal)) {
							out.Set(subX, subY, color.Gray{Y: 0})
						} else {
							out.Set(subX, subY, color.Gray{Y: 255})
						}
					}
				}
			}
		}
	}

	essentials.ConcurrentMap(0, len(indices), func(i int) {
		x, y := indices[i][0], indices[i][1]
		pxMin := XY(float64(x)*pixelWidth+min.X, float64(y)*pixelHeight+min.Y)
		pxMax := XY(float64(x+1)*pixelWidth+min.X, float64(y+1)*pixelHeight+min.Y)
		px := 1 - r.rasterizePixel(s, pxMin, pxMax)
		out.Set(x, y, color.Gray{
			Y: uint8(math.Floor(px * 255.999)),
		})
	})

	return out
}

// RasterizeCollider rasterizes the collider as a line
// drawing.
func (r *Rasterizer) RasterizeCollider(c Collider) *image.Gray {
	extraRadius := 0.5 * r.lineWidth() / r.scale()
	solid := NewColliderSolidHollow(c, extraRadius)
	return r.RasterizeSolidFilter(solid, func(r *Rect) bool {
		center := r.MinVal.Mid(r.MaxVal)
		radius := r.MinVal.Dist(center) + extraRadius
		return c.CircleCollision(center, radius)
	})
}

// RasterizeColliderSolid rasterizes the collider as a
// filled in Solid using the even-odd test.
func (r *Rasterizer) RasterizeColliderSolid(c Collider) *image.Gray {
	solid := NewColliderSolid(c)
	return r.RasterizeSolidFilter(solid, func(r *Rect) bool {
		center := r.MinVal.Mid(r.MaxVal)
		radius := r.MinVal.Dist(center)
		return c.CircleCollision(center, radius)
	})
}

func (r *Rasterizer) bounds(b Bounder) (min, max Coord) {
	if r.Bounds == nil {
		return b.Min(), b.Max()
	} else {
		return r.Bounds.Min(), r.Bounds.Max()
	}
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
