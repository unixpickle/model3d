package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unixpickle/model3d/model2d"
)

func main() {
	var smoothIters int
	var maxVertices int
	var subdivisions int
	var upsampleRate float64
	var thicken float64

	flag.IntVar(&smoothIters, "smooth-iters", 100, "number of smoothing iterations")
	flag.IntVar(&maxVertices, "max-vertices", 70, "maximum vertices after decimation")
	flag.IntVar(&subdivisions, "subdivisions", 2, "number of smoothing sub-divisions")
	flag.Float64Var(&upsampleRate, "upsample-rate", 2.0, "extra resolution to add to output")
	flag.Float64Var(&thicken, "thicken", 0.0, "extra thickness (in pixels) to give to the output")
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
	var solid model2d.Solid
	if thicken == 0 {
		solid = model2d.NewColliderSolid(collider)
	} else {
		solid = model2d.NewColliderSolidInset(collider, -thicken)
	}
	solid = &BoundedSolid{
		Solid:  solid,
		MaxVal: size,
	}
	model2d.Rasterize(flag.Args()[1], solid, upsampleRate)
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
