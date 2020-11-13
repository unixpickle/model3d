package main

import (
	"flag"
	"math"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func main() {
	var inputImage string
	var smoothIters int
	var useContentCenter bool
	flag.StringVar(&inputImage, "image", "", "input image file to process")
	flag.IntVar(&smoothIters, "smooth-iters", 50, "mesh smoothing iterations")
	flag.BoolVar(&useContentCenter, "content-center", false,
		"use the center of the content instead of the whole image")
	flag.Parse()

	if inputImage == "" {
		essentials.Die("Missing -image flag. See -help.")
	}

	bmp := model2d.MustReadBitmap(inputImage, nil)
	mesh := bmp.Mesh().SmoothSq(smoothIters)
	collider := model2d.MeshToCollider(mesh)

	var center model2d.Coord
	if useContentCenter {
		center = collider.Min().Mid(collider.Max())
	} else {
		center = model2d.XY(float64(bmp.Width), float64(bmp.Height)).Scale(0.5)
	}

	radii, maxRadius := CollectRadii(collider, center)
	object := CreatePlot(radii, maxRadius)
	model2d.Rasterize("radii.png", object, 1.0)
}

func CollectRadii(collider model2d.Collider, center model2d.Coord) (data []float64, max float64) {
	for theta := 0.0; theta < math.Pi*2; theta += 0.01 {
		ray := &model2d.Ray{
			Origin:    center,
			Direction: model2d.XY(math.Cos(theta), math.Sin(theta)),
		}
		collision, ok := collider.FirstRayCollision(ray)
		if !ok {
			data = append(data, 0)
		} else {
			data = append(data, collision.Scale)
			max = math.Max(max, collision.Scale)
		}
	}
	return
}

func CreatePlot(radii []float64, maxRadius float64) *model2d.Mesh {
	mesh := model2d.NewMesh()
	mesh.Add(&model2d.Segment{model2d.X(0), model2d.X(maxRadius)})

	for i := 1; i < len(radii); i++ {
		x0 := float64(i-1) * maxRadius / float64(len(radii))
		x1 := float64(i) * maxRadius / float64(len(radii))
		mesh.Add(&model2d.Segment{model2d.XY(x0, -radii[i-1]), model2d.XY(x1, -radii[i])})
	}
	return mesh
}
