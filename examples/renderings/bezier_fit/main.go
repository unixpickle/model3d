package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func main() {
	points := []model2d.Coord{}
	for i := 0; i < 10; i++ {
		t := float64(i) / 9 * math.Pi / 2
		points = append(points, model2d.XY(math.Cos(t), math.Sin(t)))
	}

	var palette []color.Color
	palette = append(palette, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
	for i := 0; i < 84; i++ {
		frac := uint8(float64(i) * 0xff / 84)
		palette = append(palette, color.RGBA{R: 0xff, G: frac, B: frac, A: 0xff})
		palette = append(palette, color.RGBA{R: frac, G: 0xff, B: frac, A: 0xff})
		palette = append(palette, color.RGBA{R: frac, G: frac, B: 0xff, A: 0xff})
	}

	fitter := &model2d.BezierFitter{NumIters: 1}
	var fit model2d.BezierCurve

	g := &gif.GIF{}
	for i := 0; i < 25; i++ {
		fit = fitter.FitCubic(points, fit)
		g.Image = append(g.Image, Rasterize(points, fit, palette))
		g.Delay = append(g.Delay, 10)
	}

	w, err := os.Create("rendering.gif")
	essentials.Must(err)
	defer w.Close()
	gif.EncodeAll(w, g)
}

func Rasterize(points []model2d.Coord, curve model2d.BezierCurve,
	palette []color.Color) *image.Paletted {
	rect := model2d.NewRect(model2d.XY(-0.3, -0.3), model2d.XY(1.3, 1.3))
	rast := &model2d.Rasterizer{
		Scale:     100.0,
		LineWidth: 2.0,
		Bounds:    rect,
	}

	curveMesh := model2d.NewMesh()
	const delta = 0.05
	for t := 0.0; t < 1.0; t += delta {
		curveMesh.Add(&model2d.Segment{
			curve.Eval(t),
			curve.Eval(math.Min(1, t+delta)),
		})
	}

	curveImg := rast.Rasterize(curveMesh)

	rast.LineWidth = 7.0
	dots := model2d.NewMesh()
	for _, p := range points {
		dots.Add(&model2d.Segment{p, p})
	}
	pointsImg := rast.Rasterize(dots)

	dots = model2d.NewMesh()
	for _, p := range curve[1:3] {
		dots.Add(&model2d.Segment{p, p})
	}
	controlImg := rast.Rasterize(dots)

	bg := rast.Rasterize(rect)

	colorized := model2d.ColorizeOverlay(
		[]*image.Gray{bg, pointsImg, curveImg, controlImg},
		[]color.Color{
			color.Gray{Y: 0xff},
			color.RGBA{R: 0xff, A: 0xff},
			color.RGBA{G: 0xff, A: 0xff},
			color.RGBA{B: 0xff, A: 0xff},
		},
	)
	res := image.NewPaletted(colorized.Bounds(), palette)
	draw.Draw(res, res.Bounds(), colorized, colorized.Bounds().Min, draw.Src)
	return res
}
