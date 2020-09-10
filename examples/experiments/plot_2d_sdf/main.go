package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

var (
	NeutralColor  = [3]float64{1, 1, 1}
	PositiveColor = [3]float64{101.0 / 255.0, 188.0 / 255.0, 212.0 / 255.0}
	NegativeColor = [3]float64{252.0 / 255.0, 121.0 / 255.0, 121.0 / 255.0}
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: plot_2d_sdf <input.png> <output.png>")
		os.Exit(1)
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	bmp := model2d.MustReadBitmap(inPath, nil)
	mesh := bmp.Mesh().SmoothSq(10)
	sdf := model2d.MeshToSDF(mesh)

	sdfValues := make([]float64, 0, bmp.Width*bmp.Height)
	maxValue := 0.0
	minValue := 0.0
	for y := 0; y < bmp.Height; y++ {
		for x := 0; x < bmp.Width; x++ {
			sdfValue := sdf.SDF(model2d.XY(float64(x), float64(y)))
			sdfValues = append(sdfValues, sdfValue)
			minValue = math.Min(minValue, sdfValue)
			maxValue = math.Max(maxValue, sdfValue)
		}
	}

	out := image.NewRGBA(image.Rect(0, 0, bmp.Width, bmp.Height))
	for y := 0; y < bmp.Height; y++ {
		for x := 0; x < bmp.Width; x++ {
			sdfValue := sdfValues[0]
			sdfValues = sdfValues[1:]

			var targetColor [3]float64
			var frac float64
			if sdfValue > 0 {
				targetColor = PositiveColor
				frac = sdfValue / maxValue
			} else {
				targetColor = NegativeColor
				frac = sdfValue / minValue
			}
			var rgb [3]float64
			for i, c := range targetColor {
				rgb[i] = (1-frac)*NeutralColor[i] + frac*c
			}
			out.SetRGBA(x, y, color.RGBA{
				R: uint8(rgb[0] * 255.999),
				G: uint8(rgb[1] * 255.999),
				B: uint8(rgb[2] * 255.999),
				A: 0xff,
			})
		}
	}
	f, err := os.Create(outPath)
	essentials.Must(err)
	defer f.Close()
	png.Encode(f, out)
}
