// Command render_stl renders a 3D model in an STL file to
// a PNG file from several randomized angles.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var gridSize int
	var imageSize int
	var fps float64
	var frames int
	var verbose bool
	flag.IntVar(&gridSize, "grid-size", 3, "grid size (used for rows and columns)")
	flag.IntVar(&imageSize, "image-size", 300, "size of each image in the grid")
	flag.Float64Var(&fps, "fps", 10.0, "FPS for GIF outputs")
	flag.IntVar(&frames, "frames", 20, "total number of frames for GIF outputs")
	flag.BoolVar(&verbose, "verbose", false, "run in verbose mode")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: "+os.Args[0]+" [flags] <model.stl> [output.png | output.gif]")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	stlPath := flag.Args()[0]
	if verbose {
		log.Println("Loading model from", stlPath, "...")
	}
	r, err := os.Open(stlPath)
	essentials.Must(err)
	triangles, err := model3d.ReadSTL(r)
	r.Close()
	essentials.Must(err)
	mesh := model3d.NewMeshTriangles(triangles)

	if verbose {
		log.Println("Converting mesh to collider ...")
	}
	collider := model3d.MeshToCollider(mesh)

	outPath := flag.Args()[1]
	if verbose {
		log.Println("Rendering mesh to", outPath, "...")
	}
	if strings.HasSuffix(outPath, ".gif") {
		essentials.Must(render3d.SaveRotatingGIF(
			outPath,
			collider,
			model3d.Z(1),
			model3d.YZ(-1, 0.1).Normalize(),
			imageSize,
			frames,
			fps,
			nil,
		))
	} else {
		essentials.Must(render3d.SaveRandomGrid(outPath, collider, gridSize, gridSize, imageSize, nil))
	}
}
