package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var rounding2D float64
	var numSpheres int
	var rasterResolution int
	var mcDelta float64
	var smoothIters int
	var smoothStep float64
	var maxRadius float64
	var useMedialAxis bool
	flag.Float64Var(&rounding2D, "rounding", 0.0,
		"pixels by which to round the input image's corners")
	flag.IntVar(&numSpheres, "spheres", 5000, "number of spheres to generate")
	flag.IntVar(&rasterResolution, "raster-resolution", 1000,
		"larger side-length for the rasterized height map")
	flag.Float64Var(&mcDelta, "delta", 0.01, "delta for marching cubes")
	flag.IntVar(&smoothIters, "smooth-iters", 50, "number of smoothing iterations")
	flag.Float64Var(&smoothStep, "smooth-step", 0.05, "smoothing gradient step size")
	flag.Float64Var(&maxRadius, "max-radius", -1, "if specified, the maximum sphere radius")
	flag.BoolVar(&useMedialAxis, "medial-axis", false, "use centers along the medial axis")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: pillow_2d_shape [flags] <input.png> <output.stl>")
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
	mesh2d := bmp.Mesh().Subdivide(1).SmoothSq(50)

	if rounding2D != 0 {
		log.Println("Rounding 2D mesh...")
		solid2d := model2d.NewColliderSolidInset(model2d.MeshToCollider(mesh2d), rounding2D)
		mesh2d = model2d.MarchingSquaresSearch(solid2d, 0.5, 8)
		solid2d = model2d.NewColliderSolidInset(model2d.MeshToCollider(mesh2d), -rounding2D)
		mesh2d = model2d.MarchingSquaresSearch(solid2d, 0.5, 8)
	}

	log.Println("Converting 2D mesh into SDF...")
	mesh2d = mesh2d.Scale(1 / math.Max(float64(bmp.Width), float64(bmp.Height)))
	sdf2d := model2d.MeshToSDF(mesh2d)

	log.Println("Creating height map of spheres...")

	var hmLock sync.Mutex
	hm := NewHeightMap(sdf2d.Min(), sdf2d.Max(), rasterResolution)
	totalCovered := 0

	numGos := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	for i := 0; i < numGos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localHM := NewHeightMap(sdf2d.Min(), sdf2d.Max(), rasterResolution)
			localCovered := 0
			for i := 0; i <= numSpheres/numGos; i++ {
				c := model2d.NewCoordRandUniform().Mul(sdf2d.Max().Sub(sdf2d.Min())).Add(sdf2d.Min())
				if useMedialAxis {
					c = model2d.ProjectMedialAxis(sdf2d, c, 0, 0)
				}
				dist := sdf2d.SDF(c)
				if dist < 0 {
					i--
					continue
				}
				if maxRadius != -1 && dist > maxRadius {
					dist = maxRadius
				}
				if localHM.AddSphere(c, dist) {
					localCovered++
				}
			}
			hmLock.Lock()
			defer hmLock.Unlock()
			hm.AddHeightMap(localHM)
			totalCovered += localCovered
		}()
	}
	wg.Wait()
	log.Printf(" => spheres used: %d/%d", totalCovered, (numSpheres/numGos+1)*numGos)
	log.Printf(" =>   max height: %f", hm.MaxHeight())

	log.Println("Creating mesh from height map...")
	solid := NewHeightMapSolid(hm)
	mesh := model3d.MarchingCubesSearch(solid, mcDelta, 8)

	log.Println("Smoothing mesh...")
	mesh = mesh.FlattenBase(math.Pi/2 - 0.001)
	minZ := mesh.Min().Z
	smoother := &model3d.MeshSmoother{
		StepSize:   smoothStep,
		Iterations: smoothIters,
		HardConstraintFunc: func(c model3d.Coord3D) bool {
			return c.Z == minZ
		},
	}
	mesh = smoother.Smooth(mesh)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL(outPath)
}
