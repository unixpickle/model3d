package main

import (
	"math"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func Attachment(name string) model3d.Solid {
	if name == "sphere" {
		return sphereAttachment()
	} else if name == "heart" {
		return heartAttachment()
	}
	panic("no attachment named: " + name)
}

func sphereAttachment() model3d.Solid {
	return &model3d.Sphere{Radius: 0.15}
}

func heartAttachment() model3d.Solid {
	shape := model2d.BezierCurve{
		model2d.XY(0, 0),
		model2d.XY(-3, 0),
		model2d.XY(-1, 7),
		model2d.XY(4, 2),
		model2d.XY(4, 0),
	}
	mesh := model2d.NewMesh()
	for t := 0.0; t < 1.0; t += 0.01 {
		nextT := math.Min(1.0, t+0.01)
		mesh.Add(&model2d.Segment{
			shape.Eval(t),
			shape.Eval(nextT),
		})
		mesh.Add(&model2d.Segment{
			shape.Eval(t).Mul(model2d.XY(1, -1)),
			shape.Eval(nextT).Mul(model2d.XY(1, -1)),
		})
	}
	mesh = mesh.Scale(0.4 / 7.0)
	model2d.Rasterize("/home/alex/Desktop/heart.png", mesh, 400)
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(mesh))
	sdf2d := model2d.MeshToSDF(mesh)
	spheres := toolbox3d.NewHeightMap(solid2d.Min(), solid2d.Max(), 1024)
	for i := 0; i < 2000; i++ {
		center := model2d.NewCoordRandBounds(solid2d.Min(), solid2d.Max())
		if !solid2d.Contains(center) {
			continue
		}
		center = model2d.ProjectMedialAxis(sdf2d, center, 32, 1e-5)
		fullRadius := sdf2d.SDF(center)
		spheres.AddSphereFill(center, fullRadius, 0.08)
	}
	return toolbox3d.HeightMapToSolidBidir(spheres)
}
