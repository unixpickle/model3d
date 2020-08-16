package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"

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
	mesh = mesh.Subdivide(subdivisions)

	log.Println("Rendering...")
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
	essentials.Must(model2d.SaveImage(flag.Args()[1], img))
}

func ReadImage(path string) (mesh *model2d.Mesh, size model2d.Coord) {
	bmp := model2d.MustReadBitmap(path, nil)
	size = model2d.XY(float64(bmp.Width), float64(bmp.Height))
	return bmp.Mesh(), size
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
