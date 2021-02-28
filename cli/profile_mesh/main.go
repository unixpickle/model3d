package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var smoothIters int
	var mcDelta float64
	var thickness float64

	flag.IntVar(&smoothIters, "smooth-iters", 20, "2d mesh smoothing iterations")
	flag.Float64Var(&mcDelta, "delta", 0.5, "marching cubes delta (pixels)")
	flag.Float64Var(&thickness, "thickness", 10.0, "thickness of model (pixels)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <input_image> <output.stl>\n", os.Args[0])
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}
	flag.Parse()
	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	inFile, outFile := flag.Args()[0], flag.Args()[1]
	mesh2d := model2d.MustReadBitmap(inFile, nil).FlipY().Mesh()
	if smoothIters > 0 {
		mesh2d = mesh2d.SmoothSq(smoothIters)
	}

	profile := model3d.ProfileMesh(mesh2d, 0, thickness)
	profile.SaveGroupedSTL(outFile)
}
