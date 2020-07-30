package main

import (
	"flag"
	"log"

	"github.com/unixpickle/model3d/model2d"
)

func main() {
	var inFile string
	var smoothIters int
	var maxVertices int
	var subdivisions int
	var upsampleRate float64

	flag.StringVar(&inFile, "input", "input.png", "input image file")
	flag.IntVar(&smoothIters, "smooth-iters", 100, "number of smoothing iterations")
	flag.IntVar(&maxVertices, "max-vertices", 70, "maximum vertices after decimation")
	flag.IntVar(&subdivisions, "subdivisions", 2, "number of smoothing sub-divisions")
	flag.Float64Var(&upsampleRate, "upsample-rate", 2.0, "extra resolution to add to output")
	flag.Parse()

	log.Println("Reading file as mesh...")
	mesh, size := ReadImage(inFile)

	log.Println("Smoothing mesh...")
	mesh = mesh.SmoothSq(smoothIters)

	log.Println("Decimating mesh...")
	mesh = mesh.Decimate(maxVertices)

	log.Println("Subdividing mesh...")
	mesh = mesh.Subdivide(subdivisions)

	log.Println("Rendering...")
	solid := &BoundedSolid{
		Solid:  model2d.NewColliderSolid(model2d.MeshToCollider(mesh)),
		MaxVal: size,
	}
	model2d.Rasterize("output.png", solid, upsampleRate)
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
