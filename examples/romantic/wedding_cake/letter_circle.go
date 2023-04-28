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

var LetterColor = render3d.NewColorRGB(0x0A/255.0, 0xBA/255.0, 0xB5/255.0)

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
		mesh2d := img.FlipX().Mesh().SmoothSq(20)
		mesh2d = mesh2d.Scale(0.007)
		mesh2d = mesh2d.Translate(mesh2d.Min().Mid(mesh2d.Max()).Scale(-1))
		coll2d := model2d.MeshToCollider(mesh2d)
		coll3d := model3d.ProfileCollider(coll2d, 0, 0.1)

		instantiate := func(radius, translateZ float64) model3d.Solid {
			return model3d.RotateSolid(
				model3d.TranslateSolid(
					model3d.RotateSolid(
						model3d.NewColliderSolidInset(coll3d, -0.02),
						model3d.X(1),
						-0.8*math.Pi/2,
					),
					model3d.YZ(bottomLayer.Max().Y+radius, translateZ),
				),
				model3d.Z(1),
				theta,
			)
		}
		instantiateMesh := func(radius, translateZ float64) *model3d.Mesh {
			return model3d.MarchingCubesSearch(instantiate(radius, translateZ), 0.03, 8)
		}

		// Find z translation to hit ground.
		z := bottomLayer.Min().Z - instantiateMesh(0, 0).Min().Z

		// Binary search to move letter into solid.
		outerColl := model3d.MeshToCollider(model3d.MarchingCubesSearch(bottomLayer, 0.05, 8))
		minRad := -0.1
		maxRad := 0.1
		for i := 0; i < 16; i++ {
			r := (minRad + maxRad) / 2
			mesh := instantiateMesh(r, z)
			var collides bool
			mesh.Iterate(func(t *model3d.Triangle) {
				if !collides {
					if len(outerColl.TriangleCollisions(t)) > 0 {
						collides = true
					}
				}
			})
			if collides {
				minRad = r
			} else {
				maxRad = r
			}
		}

		solids = append(solids, instantiate(minRad-0.05, z))
		theta += 2 * math.Pi / 19
	}
	return solids, toolbox3d.ConstantCoordColorFunc(LetterColor)
}
