package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Thickness = 0.1
	Slack     = 0.04

	RibbonThickness = 0.05
	RibbonWidth     = 0.25
	BowThickness    = 0.1
)

func main() {
	bottomSides := &model3d.SubtractedSolid{
		Positive: model3d.NewRect(
			model3d.Ones(-1.0),
			model3d.XYZ(1.0, 1.0, 1.0-Thickness),
		),
		Negative: model3d.NewRect(
			model3d.Ones(-1.0+Thickness),
			model3d.XYZ(1.0-Thickness, 1.0-Thickness, 1.0-Thickness+1e-5),
		),
	}
	bottomRibbon := RotateFourTimes(
		model3d.JoinedSolid{
			// On the bottom of the box.
			model3d.NewRect(
				model3d.XYZ(-(1.0+RibbonThickness), -RibbonWidth/2, -(1.0+RibbonThickness)),
				model3d.XYZ(0, RibbonWidth/2, -1.0),
			),
			// Going up the side of the box.
			model3d.NewRect(
				model3d.XYZ(-(1.0+RibbonThickness), -RibbonWidth/2, -(1.0+RibbonThickness)),
				model3d.XYZ(-1.0, RibbonWidth/2, 0.5),
			),
		},
	)
	bottom := model3d.JoinedSolid{bottomSides, bottomRibbon}

	topSides := &model3d.SubtractedSolid{
		Positive: model3d.NewRect(
			model3d.XYZ(-(1.0+Thickness+Slack), -(1.0+Thickness+Slack), 0.5),
			model3d.XYZ(1.0+Thickness+Slack, 1.0+Thickness+Slack, 1.0),
		),
		Negative: model3d.NewRect(
			model3d.XYZ(-(1.0+Slack), -(1.0+Slack), 0.5-1e-5),
			model3d.XYZ(1.0+Slack, 1.0+Slack, 1.0-Thickness),
		),
	}
	topRibbon := RotateFourTimes(model3d.JoinedSolid{
		// On side of lid
		model3d.NewRect(
			model3d.XYZ(
				-(1.0+Thickness+Slack+RibbonThickness),
				-RibbonWidth/2,
				0.5,
			),
			model3d.XYZ(-(1.0+Thickness+Slack), RibbonWidth/2, 1.0+RibbonThickness),
		),
		// On top of lid
		model3d.NewRect(
			model3d.XYZ(-(1.0+Thickness+Slack), -RibbonWidth/2, 1.0),
			model3d.XYZ(0, RibbonWidth/2, 1.0+RibbonThickness),
		),
		BowLoop(model3d.XYZ(0, 0, 1.0+BowThickness/2)),
	})
	top := model3d.JoinedSolid{topSides, topRibbon}

	log.Println("Creating bottom...")
	bottomMesh := model3d.MarchingCubesConj(
		bottom, 0.01, 8,
		&toolbox3d.AxisSqueeze{
			Axis:  toolbox3d.AxisZ,
			Min:   -0.9,
			Max:   0.4,
			Ratio: 0.05,
		},
		&toolbox3d.AxisSqueeze{
			Axis:  toolbox3d.AxisX,
			Min:   RibbonWidth + 0.1,
			Max:   0.9,
			Ratio: 0.05,
		},
		&toolbox3d.AxisSqueeze{
			Axis:  toolbox3d.AxisX,
			Min:   -0.9,
			Max:   -RibbonWidth - 0.1,
			Ratio: 0.05,
		},
	)
	bottomMesh = Decimate(bottomMesh, bottomRibbon)
	bottomColor := ColorFunc(bottomSides, bottomRibbon)
	log.Printf("Saving bottom with %d triangles...", len(bottomMesh.TriangleSlice()))
	bottomMesh.SaveMaterialOBJ("bottom.zip", TriColor(bottomColor))
	log.Println("Rendering bottom...")
	render3d.SaveRandomGrid("rendering_bottom.png", bottomMesh, 3, 3, 300, bottomColor)

	log.Println("Creating top...")
	topMesh := model3d.MarchingCubesSearch(top, 0.01, 8)
	topMesh = Decimate(topMesh, topRibbon)
	topColor := ColorFunc(topSides, topRibbon)
	log.Printf("Saving top with %d triangles...", len(topMesh.TriangleSlice()))
	topMesh.SaveMaterialOBJ("top.zip", TriColor(topColor))
	log.Println("Rendering top...")
	render3d.SaveRandomGrid("rendering_top.png", topMesh, 3, 3, 300, topColor)

	log.Println("Creating joined...")
	joined := model3d.JoinedSolid{top, bottom}
	joinedMesh := model3d.MarchingCubesSearch(joined, 0.02, 8)
	joinedColor := ColorFunc(
		model3d.JoinedSolid{topSides, bottomSides},
		model3d.JoinedSolid{topRibbon, bottomRibbon},
	)
	log.Println("Rendering joined...")
	render3d.SaveRendering("rendering_both.png", joinedMesh, model3d.XYZ(2, -4, 3),
		1000, 1000, joinedColor)

	log.Println("Done.")
}

func RotateFourTimes(s model3d.Solid) model3d.Solid {
	res := model3d.JoinedSolid{}
	for i := 0; i < 4; i++ {
		angle := math.Pi / 2.0 * float64(i)
		res = append(res, model3d.RotateSolid(s, model3d.Z(1), angle))
	}
	return res
}

func BowLoop(endpoint model3d.Coord3D) model3d.Solid {
	curve := model2d.BezierCurve{
		endpoint.XZ(),
		endpoint.XZ().Add(model2d.X(-1.0)),
		endpoint.XZ().Add(model2d.XY(-1.0, 0.9)),
		endpoint.XZ(),
	}
	mesh3d := model3d.NewMesh()
	n := 100
	for i := 0; i < n; i++ {
		p1 := curve.Eval(float64(i) / float64(n))
		p2 := curve.Eval(float64(i+1) / float64(n))
		p13d := model3d.XZ(p1.X, p1.Y)
		p23d := model3d.XZ(p2.X, p2.Y)
		off := model3d.Y(RibbonWidth / 2)
		mesh3d.AddQuad(
			p13d.Sub(off),
			p13d.Add(off),
			p23d.Add(off),
			p23d.Sub(off),
		)
	}
	return model3d.NewColliderSolidHollow(model3d.MeshToCollider(mesh3d), BowThickness/2)
}

func Decimate(m *model3d.Mesh, ribbon model3d.Solid) *model3d.Mesh {
	sdf := model3d.MeshToSDF(model3d.MarchingCubesSearch(ribbon, 0.02, 8))
	dec := model3d.Decimator{
		// Only eliminate nearly-planar surfaces.
		FeatureAngle:  0.01,
		PlaneDistance: 1e-3,
		// Never eliminate near the ribbon.
		FilterFunc: func(c model3d.Coord3D) bool {
			return sdf.SDF(c) < -0.05
		},
	}
	return dec.Decimate(m)
}

func ColorFunc(sides, ribbon model3d.Solid) render3d.ColorFunc {
	rMesh := model3d.MarchingCubesSearch(ribbon, 0.01, 8)
	rSdf := model3d.MeshToSDF(rMesh)
	sMesh := model3d.MarchingCubesSearch(sides, 0.01, 8)
	sSdf := model3d.MeshToSDF(sMesh)
	return func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
		if rSdf.SDF(c) > sSdf.SDF(c) {
			return render3d.NewColor(1.0)
		} else {
			return render3d.NewColorRGB(129.0/255.0, 216.0/255.0, 208.0/255.0)
		}
	}
}

func TriColor(colorFunc render3d.ColorFunc) func(*model3d.Triangle) [3]float64 {
	return model3d.VertexColorsToTriangle(func(c model3d.Coord3D) [3]float64 {
		r, g, b := render3d.RGB(colorFunc(c, model3d.RayCollision{}))
		return [3]float64{r, g, b}
	})
}
