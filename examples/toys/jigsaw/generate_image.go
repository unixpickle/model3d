package main

import (
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func GenerateImage(inPath, outPath string) {
	r, err := os.Open(inPath)
	essentials.Must(err)
	img, _, err := image.Decode(r)
	r.Close()
	essentials.Must(err)

	bounds := img.Bounds()

	cut := CutSolid()
	scale := BoardSize / float64(bounds.Dx())

	outImage := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			localPoint := model2d.Coord{
				X: float64(x-bounds.Min.X) * scale,
				Y: float64(bounds.Max.Y-(y+1)) * scale,
			}
			if !cut.Contains(localPoint) {
				outImage.Set(x, y, img.At(x, y))
			}
		}
	}

	w, err := os.Create(outPath)
	essentials.Must(err)
	defer w.Close()
	essentials.Must(png.Encode(w, outImage))
}
