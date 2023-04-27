package main

import (
	"fmt"
	"math"
	"path/filepath"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func LetterCircle(bottomLayer model3d.Solid) (model3d.Solid, toolbox3d.CoordColorFunc) {
	text := "Happy 1st"
	theta := 0.0
	var solids model3d.JoinedSolid
	for _, ch := range text {
		if ch == ' ' {
			theta += 2 * math.Pi / 20
			continue
		}
		img := model2d.MustReadBitmap(filepath.Join("letters", fmt.Sprintf("%c.png", ch)), nil)
		mesh2d := img.FlipY().Mesh().SmoothSq(20)
		mesh2d = mesh2d.Scale(0.01)
		mesh2d = mesh2d.Translate(mesh2d.Min().Mid(mesh2d.Max()).Scale(-1))
		coll2d := model2d.MeshToCollider(mesh2d)
		coll3d := model3d.ProfileCollider(coll2d, 0, 0.1)
		solid3d := model3d.RotateSolid(
			model3d.TranslateSolid(
				model3d.RotateSolid(
					model3d.NewColliderSolidInset(coll3d, -0.02),
					model3d.X(1),
					-0.8*math.Pi/2,
				),
				model3d.Y(bottomLayer.Max().Y),
			),
			model3d.Z(1),
			theta,
		)
		solids = append(solids, solid3d)
		theta += 2 * math.Pi / 20
	}
	return solids, toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(1.0, 0.0, 0.0))
}
