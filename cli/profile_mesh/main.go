package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	SmoothIters int     `default:"20" usage:"2d mesh smoothing iterations"`
	Thickness   float64 `default:"10.0" usage:"thickness of model (pixels)"`
}

func main() {
	var args Args
	toolbox3d.AddFlags(&args, nil)
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
	if args.SmoothIters > 0 {
		mesh2d = mesh2d.SmoothSq(args.SmoothIters)
	}

	profile := model3d.ProfileMesh(mesh2d, 0, args.Thickness)
	profile.SaveGroupedSTL(outFile)
}
