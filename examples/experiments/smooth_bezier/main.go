package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func main() {
	var smoothIters int
	var tolerance float64
	var l2Penalty float64
	var upsampleRate float64
	var thicken float64
	var outline bool
	var animationPath string
	flag.IntVar(&smoothIters, "smooth-iters", 100, "number of smoothing iterations")
	flag.Float64Var(&tolerance, "tolerance", model2d.DefaultBezierFitTolerance, "MSE tolerance")
	flag.Float64Var(&l2Penalty, "l2-penalty", 0.0, "L2 loss penalty")
	flag.Float64Var(&upsampleRate, "upsample-rate", 2.0, "extra resolution to add to output")
	flag.Float64Var(&thicken, "thicken", 0.0, "extra thickness (in pixels) to give to the output")
	flag.BoolVar(&outline, "outline", false, "produce an outline instead of a solid")
	flag.StringVar(&animationPath, "animation", "", "path for output GIF")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[flags] <input> <output>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
	}

	log.Println("Reading file as mesh...")
	mesh, size := ReadImage(flag.Args()[0])

	log.Println("Smoothing mesh...")
	mesh = mesh.SmoothSq(smoothIters)

	log.Println("Fitting beziers...")
	fitter := &model2d.BezierFitter{
		Tolerance: tolerance,
		L2Penalty: l2Penalty,
		Momentum:  0.5,
	}
	beziers := fitter.Fit(mesh)
	log.Printf("Fit %d beziers.", len(beziers))

	log.Println("Rendering...")
	outMesh := model2d.NewMesh()
	for _, b := range beziers {
		n := 100
		for i := 0; i < n; i++ {
			t1 := float64(i) / float64(n)
			t2 := float64(i+1) / float64(n)
			outMesh.Add(&model2d.Segment{
				b.Eval(t1),
				b.Eval(t2),
			})
		}
	}
	collider := model2d.MeshToCollider(outMesh)
	rast := &model2d.Rasterizer{Scale: upsampleRate, Bounds: model2d.NewRect(model2d.Coord{}, size)}
	var img image.Image
	if outline {
		img = rast.RasterizeCollider(collider)
	} else if thicken == 0 {
		img = rast.RasterizeColliderSolid(collider)
	} else {
		img = rast.RasterizeSolid(model2d.NewColliderSolidInset(collider, -thicken))
	}
	essentials.Must(model2d.SaveImage(flag.Args()[1], img))

	if animationPath != "" {
		log.Println("Rendering animation sequence...")
		RenderSequence(animationPath, upsampleRate, size, mesh, beziers)
	}
}

func RenderSequence(animationPath string, upsampleRate float64, size model2d.Coord,
	mesh *model2d.Mesh, beziers []model2d.BezierCurve) {
	rasterizeObj := func(obj interface{}, thickness float64) *image.Gray {
		rast := &model2d.Rasterizer{
			Scale:     upsampleRate,
			Bounds:    model2d.NewRect(model2d.Coord{}, size),
			LineWidth: thickness,
		}
		return rast.Rasterize(obj)
	}

	baseImages := []*image.Gray{
		rasterizeObj(model2d.NewRect(model2d.Coord{}, size), 1.0),
		rasterizeObj(mesh, 2.0),
	}
	colors := []color.Color{
		color.Gray{Y: 0xff},
		color.Gray{Y: 0xd0},
	}

	palette := make(color.Palette, 0, 256)
	for i := 0; i < 256/2; i++ {
		frac := float64(i) / float64(256/2-1)
		palette = append(palette, color.Gray{Y: uint8(frac * 0xff)})
		palette = append(palette, color.RGBA{
			R: uint8(frac * 0x65),
			G: uint8(frac * 0xbc),
			B: uint8(frac * 0xd4),
			A: uint8(frac * 0xff),
		})
	}
	g := &gif.GIF{}

	addFrame := func(imgs []*image.Gray, colors []color.Color) {
		combined := model2d.ColorizeOverlay(imgs, colors)
		frame := image.NewPaletted(combined.Bounds(), palette)
		for y := 0; y < combined.Bounds().Dy(); y++ {
			for x := 0; x < combined.Bounds().Dx(); x++ {
				frame.Set(x, y, combined.At(x, y))
			}
		}
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, 30)
	}

	addFrame(baseImages, colors)
	for _, b := range beziers {
		oneBezier := model2d.NewMesh()
		for j := 0; j < 100; j++ {
			t1 := float64(j) / 100
			t2 := float64(j+1) / 100
			oneBezier.Add(&model2d.Segment{b.Eval(t1), b.Eval(t2)})
		}
		i1 := append(baseImages, rasterizeObj(oneBezier, 15.0))
		c1 := append(colors, color.RGBA{R: 0x65, G: 0xbc, B: 0xd4, A: 0xff})
		addFrame(i1, c1)
		baseImages = append(baseImages, rasterizeObj(oneBezier, 7.0))
		colors = append(colors, color.Gray{Y: 0})
		addFrame(baseImages, colors)
	}

	w, err := os.Create(animationPath)
	essentials.Must(err)
	defer w.Close()
	gif.EncodeAll(w, g)
}

func ReadImage(path string) (mesh *model2d.Mesh, size model2d.Coord) {
	bmp := model2d.MustReadBitmap(path, nil)
	size = model2d.XY(float64(bmp.Width), float64(bmp.Height))
	return bmp.Mesh(), size
}
