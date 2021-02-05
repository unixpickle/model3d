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
	flag.Float64Var(&mcDelta, "delta", 0.02, "delta for marching cubes")
	flag.IntVar(&smoothIters, "smooth-iters", 50, "number of smoothing iterations")
	flag.Float64Var(&zScale, "z-scale", 0.4, "scale for z-axis of final mesh")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: heart_sphere [flags] <input.png> <output.stl>")
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

	log.Println("Converting 2D mesh into into...")
	mesh2d = mesh2d.Scale(2 / math.Max(float64(bmp.Width), float64(bmp.Height)))
	mesh2d = mesh2d.MapCoords(mesh2d.Min().Mid(mesh2d.Max()).Scale(-1).Add)
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(mesh2d))

	log.Println("Creating 3D mesh...")
	solid := &SpheredSolid{Solid2D: solid2d}
	xformed := model3d.TransformSolid(&model3d.Matrix3Transform{
		Matrix: &model3d.Matrix3{1, 0, 0, 0, 1, 0, 0, 0, zScale},
	}, solid)
	mesh := model3d.MarchingCubesSearch(xformed, mcDelta, 8)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL(outPath)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

type SpheredSolid struct {
	Solid2D model2d.Solid
}

func (s *SpheredSolid) Min() model3d.Coord3D {
	return model3d.XYZ(-1, -1, -1)
}

func (s *SpheredSolid) Max() model3d.Coord3D {
	return model3d.XYZ(1, 1, 1)
}

func (s *SpheredSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}
	radius := math.Sqrt(1 - c.Z*c.Z)
	c2d := c.XY().Scale(1 / (radius + 1e-5))
	return s.Solid2D.Contains(c2d)
}
