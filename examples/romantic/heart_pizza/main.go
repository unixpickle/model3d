package main

import (
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	CrustRadius = 0.2
)

func main() {
	solid := model3d.JoinedSolid{
		GetHeartRim(),
		GetPizzaBase(),
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func GetHeartRim() model3d.Solid {
	outline := GetHeartOutline().Decimate(200)
	segments := []model3d.Segment{}
	outline.Iterate(func(s *model2d.Segment) {
		segments = append(segments, model3d.NewSegment(
			model3d.XY(s[0].X, s[0].Y),
			model3d.XY(s[1].X, s[1].Y),
		))
	})
	return toolbox3d.LineJoin(CrustRadius, segments...)
}

func GetPizzaBase() model3d.Solid {
	outline := GetHeartOutline()
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(outline))
	solid3d := model3d.ProfileSolid(solid2d, -CrustRadius, 0)
	return solid3d
}

func GetHeartOutline() *model2d.Mesh {
	mesh := model2d.MustReadBitmap("heart.png", nil).FlipY().Mesh().SmoothSq(30).Scale(0.0015)
	return mesh
}
