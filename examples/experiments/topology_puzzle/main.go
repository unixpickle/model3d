package main

import (
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	StringRadius = 0.1
	BaseSlack    = 0.02
)

func main() {
	bottomStringBaseShape := &model3d.Cone{
		Radius: 0.5,
		Base:   model3d.X(-2.25),
		Tip:    model3d.XZ(-2.25, 0.5),
	}
	bottomStringHole := &model3d.Cylinder{
		Radius: StringRadius,
		P1:     model3d.X(-2.25),
		P2:     model3d.XZ(-2.25, 0.5+1e-4),
	}
	bottomStringHolder := toolbox3d.ClampAxisMax(&model3d.SubtractedSolid{
		Positive: bottomStringBaseShape,
		Negative: bottomStringHole,
	}, toolbox3d.AxisZ, 0.5-0.02-StringRadius)
	SavePiece("bottom_holder", bottomStringHolder, 0.01)

	otherStringHolder := &model3d.SubtractedSolid{
		Positive: model3d.NewRect(model3d.Origin, model3d.XYZ(1, 1, 1)),
		Negative: &model3d.Cylinder{
			Radius: StringRadius,
			P1:     model3d.XYZ(0.5, 0.5, -1),
			P2:     model3d.XYZ(0.5, 0.5, 2),
		},
	}
	SavePiece("other_holder", otherStringHolder, 0.01)

	bottomStringBaseShape.Radius += BaseSlack
	bottomStringBaseShape.Tip.Z += BaseSlack
	board := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			model3d.NewRect(model3d.XYZ(-3, -2, 0), model3d.XYZ(-1.5, 2, 0.5)),
			model3d.NewRect(model3d.XYZ(-3, -2, 0), model3d.XYZ(3, 2, 0.2)),
			Archway(),
		},
		Negative: model3d.JoinedSolid{
			bottomStringBaseShape,
			bottomStringHole,
		},
	}
	SavePiece("board", board, 0.02)
}

func Archway() model3d.Solid {
	thickness := 0.2
	height := 1.2
	base := model3d.CheckedFuncSolid(
		model3d.XYZ(-thickness, -height, 0),
		model3d.XYZ(thickness, height, height),
		func(c model3d.Coord3D) bool {
			z := height - math.Abs(c.Y)
			return math.Abs(c.Z-z) < thickness*math.Sqrt2
		},
	)
	return model3d.TranslateSolid(base, model3d.X(2.0))
}

func SavePiece(name string, piece model3d.Solid, delta float64) {
	os.Mkdir("renderings", 0755)
	os.Mkdir("models", 0755)

	log.Printf("Processing %s...", name)
	log.Println(" - creating mesh...")
	mesh := model3d.DualContour(piece, delta, true, false)
	log.Printf(" - manifold: %v", !mesh.NeedsRepair() && len(mesh.SingularVertices()) == 0)
	log.Println(" - simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println(" - saving STL...")
	mesh.SaveGroupedSTL(filepath.Join("models", name+".stl"))
	log.Println(" - rendering...")
	mid := piece.Min().Mid(piece.Max())
	size := piece.Min().Dist(piece.Max())
	cameraPos := mid.Add(model3d.XYZ(0.5, -1, 0.8).Scale(size))
	render3d.SaveRendering(filepath.Join("renderings", name+".png"), mesh, cameraPos, 512, 512,
		nil)
	if name == "board" {
		cameraPos.Z = -cameraPos.Z
		render3d.SaveRendering(filepath.Join("renderings", name+"_bottom.png"), mesh, cameraPos,
			512, 512, nil)
	}
	log.Printf(" - done with %d triangles", len(mesh.TriangleSlice()))
}
