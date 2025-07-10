package main

import (
	"flag"
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var args struct {
	HeartWidth        float64 `default:"3.0"`
	CrustThickness    float64 `default:"0.3"`
	CrustCutDepth     float64 `default:"0.09"`
	CrustCutSpacing   float64 `default:"1.0"`
	Delta             float64 `default:"0.015"`
	TomatoRadius      float64 `default:"1.0"`
	TomatoThickness   float64 `default:"0.25"`
	ScrewRadius       float64 `default:"0.2"`
	ScrewGrooveSize   float64 `default:"0.03"`
	ScrewHeadRadius   float64 `default:"0.4"`
	ScrewHeadDepth    float64 `default:"0.2"`
	ScrewCutoutRadius float64 `default:"0.6"`
	ScrewGap          float64 `default:"0.005"`
}

func main() {
	toolbox3d.AddFlags(&args, nil)
	flag.Parse()

	// For previewing
	// outline := HeartOutline()
	// outline.Scale(100).MapCoords(model2d.XY(1, -1).Mul).SavePathSVG("heart.svg")

	bread := CreateBreadSlice()
	topBread := model3d.TranslateSolid(bread, model3d.Z(args.TomatoThickness))
	bottomBread := model3d.TranslateSolid(
		model3d.VecScaleSolid(bread, model3d.XYZ(1, 1, -1)),
		model3d.Z(-args.TomatoThickness),
	)
	tomatoes := CreateTomatoes()

	screw := model3d.JoinedSolid{
		&toolbox3d.ScrewSolid{
			P1:         model3d.Z(bottomBread.Min().Z),
			P2:         model3d.Z(topBread.Max().Z),
			Radius:     args.ScrewRadius,
			GrooveSize: args.ScrewGrooveSize,
		},
		&model3d.Cylinder{
			P1:     model3d.Z(bottomBread.Min().Z),
			P2:     model3d.Z(bottomBread.Min().Z + args.ScrewHeadDepth),
			Radius: args.ScrewHeadRadius,
		},
	}
	screwCutout := &model3d.JoinedSolid{
		&toolbox3d.ScrewSolid{
			P1:         model3d.Z(bottomBread.Min().Z),
			P2:         model3d.Z(topBread.Max().Z + 1e-5),
			Radius:     args.ScrewRadius + args.ScrewGap,
			GrooveSize: args.ScrewGrooveSize,
		},
		&model3d.Cylinder{
			P1:     model3d.Z(bottomBread.Min().Z - 1e-5),
			P2:     model3d.Z(bottomBread.Min().Z + args.ScrewHeadDepth),
			Radius: args.ScrewCutoutRadius,
		},
	}

	cutTop := model3d.Subtract(topBread, screwCutout)
	cutBottom := model3d.Subtract(bottomBread, screwCutout)
	cutTomatoes := model3d.Subtract(tomatoes, screwCutout)

	log.Println("Creating top...")
	mesh := Meshify(cutTop)
	mesh.SaveGroupedSTL("panini_top.stl")

	log.Println("Creating bottom...")
	mesh = Meshify(cutBottom)
	mesh.SaveGroupedSTL("panini_bottom.stl")

	log.Println("Creating tomatoes...")
	mesh = Meshify(cutTomatoes)
	mesh.SaveGroupedSTL("panini_tomatoes.stl")

	log.Println("Creating screw...")
	mesh = Meshify(screw)
	mesh.SaveGroupedSTL("panini_screw.stl")
}

func Meshify(s model3d.Solid) *model3d.Mesh {
	return model3d.DualContour(s, args.Delta, true, false).EliminateCoplanar(1e-5)
}

func CreateBreadSlice() model3d.Solid {
	outline := HeartOutline()
	tris2d := model2d.TriangulateMesh(outline)
	tris3d := model3d.NewMesh()
	for _, t := range tris2d {
		tris3d.Add(&model3d.Triangle{
			model3d.XY(t[0].X, t[0].Y),
			model3d.XY(t[1].X, t[1].Y),
			model3d.XY(t[2].X, t[2].Y),
		})
	}
	sol3d := model3d.SDFToSolid(model3d.MeshToSDF(tris3d), args.CrustThickness/2)

	// Cut out small lines
	cuts := model3d.JoinedSolid{}
	for x := -args.HeartWidth * 2; x < args.HeartWidth*2; x += args.CrustCutSpacing {
		cut := &model3d.Cylinder{
			P1:     model3d.XYZ(-50+x+0.45, -80, args.CrustThickness/2),
			P2:     model3d.XYZ(50+x+0.45, 80, args.CrustThickness/2),
			Radius: args.CrustCutDepth,
		}
		cuts = append(cuts, cut)
	}

	return model3d.Subtract(sol3d, cuts)
}

func CreateTomatoes() model3d.Solid {
	outline := HeartOutline()
	outlineSolid := model2d.NewColliderSolid(model2d.MeshToCollider(outline))

	getArea := func(solid model2d.Solid) float64 {
		return model2d.MarchingSquaresSearch(solid, 0.05, 8).Area()
	}

	tomato := &model2d.Circle{Center: model2d.Y(-0.2), Radius: args.TomatoRadius}
	tomatoes := model2d.JoinedSolid{}
	for _, direction := range []model2d.Coord{
		model2d.XY(1.2, -1), model2d.XY(-1.2, -1),
		model2d.XY(3.3, 1), model2d.XY(-3.3, 1),
		model2d.XY(1.0, 1), model2d.XY(-1.0, 1),
		model2d.Y(1.0), model2d.Y(-1.0),
	} {
		delta := 0.02
		for scale := 0.0; true; scale += delta {
			t := model2d.TranslateSolid(tomato, direction.Scale(scale))
			if getArea(model2d.JoinedSolid{outlineSolid, t}) > getArea(outlineSolid)+0.01 {
				tomatoes = append(tomatoes, t)
				break
			}
		}
	}

	return model3d.ProfileSolid(tomatoes, -args.TomatoThickness/2, args.TomatoThickness/2)
}
