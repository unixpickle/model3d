package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	sphere := model3d.NewMeshIcosphere(model3d.Origin, 5.0, 8)
	sphere.Iterate(func(t *model3d.Triangle) {
		if t.Min().Z < 0 {
			sphere.Remove(t)
		}
	})

	param := model3d.Floater97(
		sphere,
		model3d.SquareBoundary(sphere),
		model3d.Floater97ShapePreservingWeights(sphere),
		nil,
	)

	mesh2d := model2d.NewMesh()
	sphere.Iterate(func(t *model3d.Triangle) {
		for _, s := range t.Segments() {
			s2 := &model2d.Segment{
				param.Value(s[0]),
				param.Value(s[1]),
			}
			if len(mesh2d.Find(s2[0], s2[1])) == 0 {
				mesh2d.Add(s2)
			}
		}
	})

	model2d.Rasterize("rendering.png", mesh2d, 100.0)
}
