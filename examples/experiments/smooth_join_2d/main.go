package main

import (
	"image/png"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
)

func main() {
	RenderExperiment("easy_case", EasyCase)
	RenderExperiment("middle_case", MiddleCase)
	RenderExperiment("parallel_case", ParallelCase)
}

func RenderExperiment(name string, f func() (noSmooth, v1, v2 model2d.Solid)) {
	noSmooth, v1, v2 := f()
	r := &model2d.Rasterizer{
		Scale:     150.0,
		LineWidth: 2.0,
		Bounds:    model2d.NewRect(v1.Min(), v1.Max()),
	}
	meshes := []*model2d.Mesh{
		model2d.MarchingSquaresSearch(noSmooth, 0.01, 8),
		model2d.MarchingSquaresSearch(v1, 0.01, 8),
		model2d.MarchingSquaresSearch(v2, 0.01, 8),
	}
	names := []string{name + "_input.png", name + "_v1.png", name + "_v2.png"}
	for i, mesh := range meshes {
		name := names[i]
		img := r.Rasterize(mesh)
		f, err := os.Create(filepath.Join("images", name))
		essentials.Must(err)
		png.Encode(f, img)
		f.Close()
	}
}

func EasyCase() (noSmooth, v1, v2 model2d.Solid) {
	square := model2d.NewRect(model2d.XY(-1.0, -1.0), model2d.XY(1.0, 1.0))
	circle := &model2d.Circle{Center: model2d.XY(1, 1), Radius: 1.0}
	return model2d.SmoothJoin(0.0, square, circle),
		model2d.SmoothJoin(0.2, square, circle),
		model2d.SmoothJoinV2(0.2, square, circle)
}

func MiddleCase() (noSmooth, v1, v2 model2d.Solid) {
	square := model2d.NewRect(model2d.XY(-0.5, -1.0), model2d.XY(0.0, 1.0))
	circle := &model2d.Circle{Center: model2d.XY(0, 1), Radius: 1.0}
	return model2d.SmoothJoin(0.0, square, circle),
		model2d.SmoothJoin(0.2, square, circle),
		model2d.SmoothJoinV2(0.2, square, circle)
}

func ParallelCase() (noSmooth, v1, v2 model2d.Solid) {
	square := model2d.NewRect(model2d.XY(-0.95, -1.0), model2d.XY(0.0, 1.0))
	circle := &model2d.Circle{Center: model2d.XY(0, 1), Radius: 1.0}
	return model2d.SmoothJoin(0.0, square, circle),
		model2d.SmoothJoin(0.2, square, circle),
		model2d.SmoothJoinV2(0.2, square, circle)
}
