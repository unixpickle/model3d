package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var mcDelta float64
	var smoothIters int
	var zScale float64
	var searchGridSize int
	var searchIters int
	flag.Float64Var(&mcDelta, "delta", 0.02, "delta for marching cubes")
	flag.IntVar(&smoothIters, "smooth-iters", 50, "number of smoothing iterations")
	flag.Float64Var(&zScale, "z-scale", 1.0, "scale for z-axis of final mesh")
	flag.IntVar(&searchGridSize, "search-grid", 30, "grid size for max search")
	flag.IntVar(&searchIters, "search-iters", 3, "recursive iterations for max search")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pillow_sdf [flags] <input.png> <output.stl>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	flag.Parse()
	if len(flag.Args()) != 2 {
		flag.Usage()
	}

	inPath := flag.Args()[0]
	outPath := flag.Args()[1]

	log.Println("Loading 2D bitmap into a 2D mesh...")
	bmp := model2d.MustReadBitmap(inPath, nil).FlipY()
	mesh2d := bmp.Mesh().SmoothSq(smoothIters)

	log.Println("Converting 2D mesh into SDF...")
	mesh2d = mesh2d.Scale(2 / math.Max(float64(bmp.Width), float64(bmp.Height)))
	mesh2d = mesh2d.MapCoords(mesh2d.Min().Mid(mesh2d.Max()).Scale(-1).Add)
	sdf2d := model2d.MeshToSDF(mesh2d)

	log.Println("Finding central point...")
	center := FindSDFMax(sdf2d, searchGridSize, searchIters)

	log.Println("Creating 3D mesh...")
	solid := NewPillowedSolid(sdf2d, center)
	xformed := model3d.TransformSolid(&model3d.Matrix3Transform{
		Matrix: &model3d.Matrix3{1, 0, 0, 0, 1, 0, 0, 0, zScale},
	}, solid)
	mesh := model3d.MarchingCubesSearch(xformed, mcDelta, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL(outPath)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func FindSDFMax(sdf model2d.SDF, gridSize, rounds int) model2d.Coord {
	boundsMin, boundsMax := sdf.Min(), sdf.Max()
	bestPoint := model2d.Coord{}
	bestSDF := math.Inf(-1)
	for i := 0; i < rounds; i++ {
		for x := 0; x < gridSize; x++ {
			for y := 0; y < gridSize; y++ {
				xy := model2d.XY(float64(x)/float64(gridSize-1), float64(y)/float64(gridSize-1))
				point := boundsMin.Add(boundsMax.Sub(boundsMin).Mul(xy))
				value := sdf.SDF(point)
				if value > bestSDF {
					bestSDF = value
					bestPoint = point
				}
			}
		}
		size := boundsMax.Sub(boundsMin).Scale(1 / float64(gridSize-1))
		boundsMin = bestPoint.Sub(size)
		boundsMax = bestPoint.Add(size)
	}
	return bestPoint
}

type PillowedSolid struct {
	sdf      model2d.SDF
	center   model2d.Coord
	radius   float64
	fullSize float64
}

func NewPillowedSolid(sdf model2d.SDF, center model2d.Coord) *PillowedSolid {
	return &PillowedSolid{
		sdf:    sdf,
		center: center,
		radius: sdf.SDF(center),
	}
}

func (p *PillowedSolid) Min() model3d.Coord3D {
	min := p.sdf.Min()
	return model3d.XYZ(min.X, min.Y, -p.radius)
}

func (p *PillowedSolid) Max() model3d.Coord3D {
	max := p.sdf.Max()
	return model3d.XYZ(max.X, max.Y, p.radius)
}

func (p *PillowedSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	radius := math.Sqrt(p.radius*p.radius - c.Z*c.Z)
	c2d := c.XY()
	dist := p.sdf.SDF(c2d)
	return dist-(p.radius-radius) > 0
}
