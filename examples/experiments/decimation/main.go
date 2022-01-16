package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/unixpickle/model3d/model3d"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	model := &model3d.Sphere{Radius: 1.0}
	mesh := model3d.MarchingCubesSearch(model, 0.01, 8)
	numTris := len(mesh.TriangleSlice())
	log.Printf("original: triangles=%d error=%e", numTris, MeanError(model, mesh))

	for eps := 0.0005; eps < 0.1; eps *= 1.5 {
		dec := &model3d.Decimator{
			PlaneDistance:    eps,
			BoundaryDistance: eps,
		}
		mesh1 := dec.Decimate(mesh)
		numTris1 := len(mesh1.TriangleSlice())
		log.Printf("eps %f: triangles=%d error=%e", eps, numTris1, MeanError(model, mesh1))
	}
}

func MeanError(model model3d.SDF, mesh *model3d.Mesh) float64 {
	meshSDF := model3d.MeshToSDF(mesh)
	totalMSE := 0.0
	count := 0.0
	for i := 0; i < 10000; i++ {
		p := model3d.NewCoord3DRandBounds(model.Min(), model.Max())
		real := model.SDF(p)
		approx := meshSDF.SDF(p)
		totalMSE += (real - approx) * (real - approx)
		count += 1
	}
	return totalMSE / count
}
