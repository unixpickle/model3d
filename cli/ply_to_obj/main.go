// Command ply_to_obj loads a colored mesh from a PLY file
// and exports it as a textured OBJ file.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	var textureSize int
	var noImage bool
	var smoothStepSize float64
	var smoothIters int
	flag.IntVar(&textureSize, "texture-size", 32, "resolution of texture image")
	flag.BoolVar(&noImage, "no-image", false,
		"use a quantized per-face material instead of a texture image")
	flag.Float64Var(&smoothStepSize, "smooth-step-size", 0.05, "step size for Laplacian smoothing")
	flag.IntVar(&smoothIters, "smooth-iters", 0, "steps of Laplacian smoothing")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ply_to_obj [flags] input.ply output.zip")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		os.Exit(1)
	}
	inputPath, outputPath := args[0], args[1]

	f, err := os.Open(inputPath)
	essentials.Must(err)
	tris, colors, err := model3d.ReadColorPLY(f)
	f.Close()
	essentials.Must(err)

	mesh := model3d.NewMeshTriangles(tris)

	cf := toolbox3d.CoordColorFunc(func(c model3d.Coord3D) render3d.Color {
		arr := colors.Value(c)
		return render3d.NewColorRGB(float64(arr[0])/255, float64(arr[1])/255, float64(arr[2])/255)
	})

	if smoothIters > 0 {
		smoother := &model3d.MeshSmoother{
			StepSize:   smoothStepSize,
			Iterations: smoothIters,
		}
		mapping := smoother.SmoothMapping(mesh)
		mesh = mesh.MapCoords(mapping.Value)

		inv := model3d.NewCoordMap[model3d.Coord3D]()
		mapping.Range(func(k, v model3d.Coord3D) bool {
			inv.Store(v, k)
			return true
		})
		cf = cf.Map(inv.Value)
	}

	if noImage {
		essentials.Must(
			mesh.SaveMaterialOBJ(
				outputPath,
				cf.Cached().QuantizedTriangleColor(mesh, textureSize*textureSize),
			),
		)
	} else {
		essentials.Must(
			mesh.SaveQuantizedMaterialOBJ(
				outputPath,
				textureSize,
				cf.Cached().TriangleColor,
			),
		)
	}
}
