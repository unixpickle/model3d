package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	CornerSmoothing = 0.04
	Thickness       = 0.05

	HeartCutSize    = 0.08
	HeartCutXOffset = 0.03

	EngravingDepth = 0.025

	HookRadius     = 0.3
	HookThickness  = 0.05
	HookZThickness = 0.04
)

func main() {
	heartMesh := model2d.MustReadBitmap("outline.png", nil).Mesh().SmoothSq(50)
	heartMesh = heartMesh.MapCoords(heartMesh.Min().Scale(-1).Add)
	heartMesh = heartMesh.MapCoords(heartMesh.Max().Recip().Mul)
	solid := model2d.JoinedSolid{
		model2d.NewColliderSolid(model2d.MeshToCollider(heartMesh)),
		&model2d.Rect{
			MinVal: model2d.XY(0.45, 0.8),
			MaxVal: model2d.XY(0.55, 2.5),
		},
		&model2d.Rect{
			MinVal: model2d.XY(0.5, 1.65+0.5),
			MaxVal: model2d.XY(0.8, 1.75+0.5),
		},
		&model2d.Rect{
			MinVal: model2d.XY(0.5, 1.85+0.5),
			MaxVal: model2d.XY(0.8, 1.95+0.5),
		},
	}

	model2d.Rasterize("rendering_2d.png", solid, 200.0)

	log.Println("Creating 3D solid...")
	mesh2d := model2d.MarchingSquaresSearch(solid, 0.01, 8)
	collider2d := model2d.MeshToCollider(mesh2d)
	collider3d := model3d.ProfileCollider(collider2d, -Thickness/2, Thickness/2)
	smoothEdges := model3d.NewColliderSolidHollow(collider3d, CornerSmoothing)
	solid3d := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			smoothEdges,
			model3d.ProfileSolid(solid, -Thickness/2, Thickness/2),
			CreateHook(collider2d),
		},
		Negative: model3d.JoinedSolid{
			CreateEngraving(),
			CreateHeartSlice(),
		},
	}

	log.Println("Creating 3D mesh...")
	mesh3d := model3d.MarchingCubesSearch(solid3d, 0.0025, 8)
	log.Println("Simplifying mesh...")
	mesh3d = mesh3d.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	mesh3d.SaveGroupedSTL("key.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_3d.png", mesh3d, 3, 3, 300, nil)
}

func CreateHook(existingKey model2d.Collider) model3d.Solid {
	rc, ok := existingKey.FirstRayCollision(&model2d.Ray{
		Origin:    model2d.XY(0.5, -10),
		Direction: model2d.Y(1),
	})
	if !ok {
		panic("no ray collision for middle of heart")
	}
	y := rc.Scale - 10
	center := model2d.XY(0.5, y)
	hook2d := &model2d.SubtractedSolid{
		Positive: &model2d.Circle{
			Center: center,
			Radius: HookRadius,
		},
		Negative: &model2d.Circle{
			Center: center,
			Radius: HookRadius - HookThickness,
		},
	}
	return model3d.ProfileSolid(hook2d, -HookZThickness/2, HookZThickness/2)
}

func CreateEngraving() model3d.Solid {
	bmp := model2d.MustReadBitmap("engraving.png", nil).FlipX()
	mesh2d := bmp.Mesh().SmoothSq(50)
	mesh2d = mesh2d.Scale(1 / float64(bmp.Width))
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(mesh2d))
	return model3d.ProfileSolid(solid2d, Thickness/2+CornerSmoothing-EngravingDepth, Thickness/2+CornerSmoothing+1e-5)
}

func CreateHeartSlice() model3d.Solid {
	heartMesh := model2d.MustReadBitmap("outline.png", nil).Mesh().SmoothSq(50)
	heartMesh = heartMesh.MapCoords(heartMesh.Min().Scale(-1).Add)
	heartMesh = heartMesh.MapCoords(heartMesh.Max().Recip().Mul)
	heartMesh = heartMesh.Scale(HeartCutSize)
	heartMesh = heartMesh.MapCoords(model2d.NewMatrix2Rotation(-math.Pi / 2).MulColumn)
	heartSolid := model2d.NewColliderSolid(model2d.MeshToCollider(heartMesh))

	cuts := model2d.JoinedSolid{}

	for y := 1.2; y < 2.0; y += HeartCutSize * 2 {
		cut := model2d.TransformSolid(&model2d.Translate{
			Offset: model2d.XY(0.5+HeartCutXOffset, y),
		}, heartSolid)
		cuts = append(cuts, cut)
	}

	return model3d.ProfileSolid(
		cuts,
		-(Thickness/2 + CornerSmoothing + 1e-5),
		Thickness/2+CornerSmoothing+1e-5,
	)
}
