package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func main() {
	var smoothIters int
	var maxVertices int
	var subdivisions int
	var upsampleRate float64
	var thicken float64
	var outline bool

	flag.IntVar(&smoothIters, "smooth-iters", 100, "number of smoothing iterations")
	flag.IntVar(&maxVertices, "max-vertices", 70, "maximum vertices after decimation")
	flag.IntVar(&subdivisions, "subdivisions", 2, "number of smoothing sub-divisions")
	flag.Float64Var(&upsampleRate, "upsample-rate", 2.0, "extra resolution to add to output")
	flag.Float64Var(&thicken, "thicken", 0.0, "extra thickness (in pixels) to give to the output")
	flag.BoolVar(&outline, "outline", false, "produce an outline instead of a solid")
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

	log.Println("Decimating mesh...")
	mesh = mesh.Decimate(maxVertices)

	log.Println("Subdividing mesh...")
	if subdivisions > 0 {
		mesh = mesh.Subdivide(subdivisions)
	}

	log.Println("Rendering...")
	outPath := flag.Args()[1]
	if IsSVGPath(outPath) {
		SaveSVG(outPath, mesh, size, outline, upsampleRate)
	} else {
		SaveRasterImage(outPath, mesh, size, outline, thicken, upsampleRate)
	}
}

func SaveSVG(outPath string, mesh *model2d.Mesh, size model2d.Coord, outline bool,
	upsampleRate float64) {
	mesh = mesh.Scale(upsampleRate)
	size = size.Scale(upsampleRate)
	bounds := model2d.NewRect(model2d.Origin, size)

	var data []byte
	if outline {
		data = model2d.EncodeCustomSVG(
			[]*model2d.Mesh{mesh},
			[]string{"black"},
			[]float64{1.0},
			bounds,
		)
	} else {
		data = model2d.EncodeCustomPathSVG(
			[]*model2d.Mesh{mesh},
			[]string{"black"},
			[]string{"none"},
			[]float64{1.0},
			bounds,
		)
	}
	essentials.Must(ioutil.WriteFile(outPath, data, 0644))
}

func SaveRasterImage(outPath string, mesh *model2d.Mesh, size model2d.Coord, outline bool,
	thicken, upsampleRate float64) {
	collider := model2d.MeshToCollider(mesh)
	rast := &model2d.Rasterizer{Scale: upsampleRate}
	var img image.Image
	if outline {
		img = rast.RasterizeCollider(&BoundedCollider{
			Collider: collider,
			MaxVal:   size,
		})
	} else if thicken == 0 {
		img = rast.RasterizeColliderSolid(&BoundedCollider{
			Collider: collider,
			MaxVal:   size,
		})
	} else {
		img = rast.RasterizeSolid(&BoundedSolid{
			Solid:  model2d.NewColliderSolidInset(collider, -thicken),
			MaxVal: size,
		})
	}
	essentials.Must(model2d.SaveImage(outPath, img))
}

func ReadImage(path string) (mesh *model2d.Mesh, size model2d.Coord) {
	bmp := model2d.MustReadBitmap(path, nil)
	size = model2d.XY(float64(bmp.Width), float64(bmp.Height))
	return bmp.Mesh(), size
}

func IsSVGPath(path string) bool {
	return filepath.Ext(strings.ToLower(path)) == ".svg"
}

type BoundedSolid struct {
	model2d.Solid
	MaxVal model2d.Coord
}

func (b BoundedSolid) Min() model2d.Coord {
	return model2d.Coord{}
}

func (b BoundedSolid) Max() model2d.Coord {
	return b.MaxVal
}

type BoundedCollider struct {
	model2d.Collider
	MaxVal model2d.Coord
}

func (b BoundedCollider) Min() model2d.Coord {
	return model2d.Coord{}
}

func (b BoundedCollider) Max() model2d.Coord {
	return b.MaxVal
}
