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
	var numSpheres int
	var rasterResolution int
	var mcDelta float64
	var smoothIters int
	flag.IntVar(&numSpheres, "spheres", 5000, "number of spheres to generate")
	flag.IntVar(&rasterResolution, "raster-resolution", 1000,
		"larger side-length for the rasterized height map")
	flag.Float64Var(&mcDelta, "delta", 0.01, "delta for marching cubes")
	flag.IntVar(&smoothIters, "smooth-iters", 10, "number of smoothing iterations")
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

	log.Println("Loading 2D bitmap into an SDF...")
	bmp := model2d.MustReadBitmap(inPath, nil).FlipY()
	mesh2d := bmp.Mesh().SmoothSq(10)
	mesh2d = mesh2d.Scale(1 / math.Max(float64(bmp.Width), float64(bmp.Height)))
	sdf2d := model2d.MeshToSDF(mesh2d)

	log.Println("Creating height map of spheres...")

	var hmLock sync.Mutex
	hm := NewHeightMap(sdf2d.Min(), sdf2d.Max(), rasterResolution)

	numGos := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	for i := 0; i < numGos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localHM := NewHeightMap(sdf2d.Min(), sdf2d.Max(), rasterResolution)
			for i := 0; i <= numSpheres/numGos; i++ {
				c := model2d.NewCoordRandUniform().Mul(sdf2d.Max().Sub(sdf2d.Min())).Add(sdf2d.Min())
				dist := sdf2d.SDF(c)
				if dist < 0 {
					i--
					continue
				}
				localHM.AddSphere(c, dist)
			}
			hmLock.Lock()
			defer hmLock.Unlock()
			hm.AddHeightMap(localHM)
		}()
	}
	wg.Wait()

	log.Println("Creating mesh from height map...")
	solid := NewHeightMapSolid(hm)
	mesh := model3d.MarchingCubesSearch(solid, mcDelta, 8)

	log.Println("Smoothing mesh...")
	var smoothRates []float64
	for i := 0; i < smoothIters; i++ {
		smoothRates = append(smoothRates, -1)
	}
	mesh = mesh.Blur(smoothRates...)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL(outPath)
}

type HeightMapSolid struct {
	HeightMap *HeightMap
	MaxHeight float64
}

func NewHeightMapSolid(hm *HeightMap) *HeightMapSolid {
	return &HeightMapSolid{
		HeightMap: hm,
		MaxHeight: hm.MaxHeight(),
	}
}

func (h *HeightMapSolid) Min() model3d.Coord3D {
	return model3d.XY(h.HeightMap.Min.X, h.HeightMap.Min.Y)
}

func (h *HeightMapSolid) Max() model3d.Coord3D {
	return model3d.XYZ(h.HeightMap.Max.X, h.HeightMap.Max.Y, h.MaxHeight)
}

func (h *HeightMapSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(h, c) && c.Z*c.Z < h.HeightMap.GetHeightSquared(c.XY())
}
