package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var (
	WoodColor  = render3d.NewColorRGB(0.5, 0.3, 0.1)
	TableColor = render3d.NewColorRGB(0.7, 0.4, 0.2)
)

func main() {
	roof := Roof()
	wood := Body()
	chair, chairColor := Chair()
	table := Table()

	xf1 := model3d.JoinedTransform{
		model3d.Rotation(model3d.Z(1.0), -0.2),
		&model3d.Translate{Offset: model3d.XYZ(-0.5, 0.0, -0.8)},
	}
	xf2 := model3d.JoinedTransform{
		model3d.Rotation(model3d.Z(1.0), 0.2),
		&model3d.Translate{Offset: model3d.XYZ(0.5, 0.0, -0.8)},
	}
	chair1 := model3d.TransformSolid(xf1, chair)
	chair1Color := chairColor.Transform(xf1)
	chair2 := model3d.TransformSolid(xf2, chair)
	chair2Color := chairColor.Transform(xf2)

	solid := model3d.JoinedSolid{
		roof,
		wood,
		chair1,
		chair2,
		table,
	}

	log.Println("Creating mesh...")
	mesh, interior := model3d.DualContourInterior(solid, 0.005, true, false)
	colorFn := toolbox3d.JoinedSolidCoordColorFunc(
		interior,
		roof, toolbox3d.ConstantCoordColorFunc(render3d.NewColor(1.0)),
		wood, toolbox3d.ConstantCoordColorFunc(WoodColor),
		chair1, chair1Color,
		chair2, chair2Color,
		table, toolbox3d.ConstantCoordColorFunc(TableColor),
	).Cached()

	log.Println("Simplifying mesh...")
	oldCount := mesh.NumTriangles()
	mesh = mesh.EliminateCoplanarFiltered(0.01, colorFn.ChangeFilterFunc(mesh, 0.1))
	newCount := mesh.NumTriangles()
	log.Printf(
		" => went from %d to %d triangles (%.02f%% reduction)",
		oldCount, newCount, 100*float64(oldCount-newCount)/float64(oldCount),
	)

	log.Println("Saving...")
	mesh.SaveMaterialOBJ("cabana.zip", colorFn.TriangleColor)

	log.Println("Rendering...")
	render3d.SaveRendering("rendering.png", mesh, model3d.XYZ(-1.7, 4.0, 0.6), 512, 512, colorFn.RenderColor)
}

func Roof() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1, -1, 0.5),
		model3d.XYZ(1, 1, 1.1),
		func(c model3d.Coord3D) bool {
			xTop := math.Min(c.X, -c.X)
			yTop := math.Min(c.Y, -c.Y)
			top := 1.0 + 0.5*math.Min(xTop, yTop)
			return c.Z < top && c.Z > top-0.05
		},
	)
}

func Body() model3d.Solid {
	return model3d.JoinedSolid{
		model3d.NewRect(
			model3d.XYZ(-1, -1, -1),
			model3d.XYZ(1, 1, -0.8),
		),
		model3d.NewRect(
			model3d.XYZ(-1, -1, -1),
			model3d.XYZ(-1+0.1, -1+0.1, 0.5),
		),
		model3d.NewRect(
			model3d.XYZ(-1, 1-0.1, -1),
			model3d.XYZ(-1+0.1, 1, 0.5),
		),
		model3d.NewRect(
			model3d.XYZ(1-0.1, -1, -1),
			model3d.XYZ(1, -1+0.1, 0.5),
		),
		model3d.NewRect(
			model3d.XYZ(1-0.1, 1-0.1, -1),
			model3d.XYZ(1, 1, 0.5),
		),
	}.Optimize()
}

func Chair() (model3d.Solid, toolbox3d.CoordColorFunc) {
	bodyMesh := model3d.NewMesh()
	bodyMesh.AddQuad(
		model3d.XYZ(-0.2, -0.4, 0.2),
		model3d.XYZ(-0.2, 0.6, 0.2),
		model3d.XYZ(0.2, 0.6, 0.2),
		model3d.XYZ(0.2, -0.4, 0.2),
	)
	bodyMesh.AddQuad(
		model3d.XYZ(-0.2, -0.4, 0.2),
		model3d.XYZ(0.2, -0.4, 0.2),
		model3d.XYZ(0.2, -0.8, 0.6),
		model3d.XYZ(-0.2, -0.8, 0.6),
	)
	thickChair := model3d.NewColliderSolidHollow(model3d.MeshToCollider(bodyMesh), 0.05)

	chairBottom := model3d.JoinedSolid{
		model3d.NewRect(
			model3d.XYZ(-0.2, -0.4, 0.1),
			model3d.XYZ(0.2, 0.6, 0.2),
		),
		model3d.NewRect(
			model3d.XYZ(-0.2, -0.4, 0.0),
			model3d.XYZ(-0.1, 0.6, 0.2),
		),
		model3d.NewRect(
			model3d.XYZ(0.1, -0.4, 0.0),
			model3d.XYZ(0.2, 0.6, 0.2),
		),
	}

	return model3d.JoinedSolid{thickChair, chairBottom}, func(c model3d.Coord3D) render3d.Color {
		if thickChair.Contains(c) {
			if math.Mod(c.X+10.05, 0.2) < 0.1 {
				return render3d.NewColorRGB(16.0/255.0, 48.0/255.0, 149.0/255.0)
			} else {
				return render3d.NewColor(1.0)
			}
		} else {
			return render3d.NewColor(0)
		}
	}
}

func Table() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.YZ(-0.2, -0.6),
			P2:     model3d.YZ(-0.2, -0.55),
			Radius: 0.2,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(-0.15, -0.2, -0.8),
			P2:     model3d.XYZ(-0.15, -0.2, -0.6001),
			Radius: 0.05,
		},
		&model3d.Cylinder{
			P1:     model3d.XYZ(0.15, -0.2, -0.8),
			P2:     model3d.XYZ(0.15, -0.2, -0.6001),
			Radius: 0.05,
		},

		&model3d.Cylinder{
			P1:     model3d.YZ(-0.2+0.15, -0.8),
			P2:     model3d.YZ(-0.2+0.15, -0.6001),
			Radius: 0.05,
		},
		&model3d.Cylinder{
			P1:     model3d.YZ(-0.2-0.15, -0.8),
			P2:     model3d.YZ(-0.2-0.15, -0.6001),
			Radius: 0.05,
		},
	}
}
