package main

import (
	"math"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	HeartSamples     = 5000
	HeartSmoothInset = 0.1
	HeartMaxRadius   = 0.29
	HeartDepthScale  = 0.75
)

func CreateHeart() model3d.Solid {
	roughHeart := createRoughHeart()
	mesh := model3d.MarchingCubesSearch(roughHeart, MarchingDelta, 8)
	smoother := model3d.MeshSmoother{
		StepSize:   0.05,
		Iterations: 50,
	}
	mesh = smoother.Smooth(mesh)
	return model3d.NewColliderSolid(model3d.MeshToCollider(mesh))
}

func createRoughHeart() model3d.Solid {
	mesh2d := model2d.MustReadBitmap("heart.png", nil).FlipY().Mesh().SmoothSq(50)
	min := mesh2d.Min()
	max := mesh2d.Max()
	mesh2d = mesh2d.MapCoords(func(c model2d.Coord) model2d.Coord {
		return c.Sub(min).Div(max.Sub(min))
	})

	for _, inset := range []float64{HeartSmoothInset, -HeartSmoothInset} {
		solid := model2d.NewColliderSolidInset(model2d.MeshToCollider(mesh2d), inset)
		mesh2d = model2d.MarchingSquaresSearch(solid, HeartSmoothInset/5, 8)
	}

	sdf2d := model2d.MeshToSDF(mesh2d)

	hm := toolbox3d.NewHeightMap(sdf2d.Min(), sdf2d.Max(), 1000)
	var sphereCount int
	for sphereCount < HeartSamples {
		c := model2d.NewCoordRandUniform()
		if sdf2d.SDF(c) < 0 {
			continue
		}
		sphereCount++
		proj := model2d.ProjectMedialAxis(sdf2d, c, 0, 0)
		radius := sdf2d.SDF(proj)
		hm.AddSphere(proj, math.Min(radius, HeartMaxRadius))
	}

	return model3d.TransformSolid(&model3d.Matrix3Transform{
		Matrix: &model3d.Matrix3{1, 0, 0, 0, 0, HeartDepthScale, 0, 1, 0},
	}, toolbox3d.HeightMapToSolidBidir(hm))
}
