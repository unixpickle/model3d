package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

var (
	DirtColor    = render3d.NewColorRGB(0.33, 0.25, 0.17)
	VaseColor    = render3d.NewColorRGB(0.66, 0.44, 0.20)
	CactusColor1 = render3d.NewColorRGB(0.0, 0.7, 0.0)
	CactusColor2 = render3d.NewColorRGB(0.0, 0.5, 0.0)
)

func main() {
	vase := VaseSolid()
	body := CactusSolid()
	joined := model3d.JoinedSolid{
		vase,
		body,
	}
	mesh := model3d.MarchingCubesSearch(joined, 0.02, 8)
	mesh.SaveGroupedSTL("cactus.stl")
	colorFunc := ColorFunc()
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc)
	render3d.SaveRotatingGIF("rendering.gif", mesh, model3d.Z(1), model3d.Y(1), 300, 20, 5.0, colorFunc)

	triColor := model3d.VertexColorsToTriangle(func(c model3d.Coord3D) [3]float64 {
		r, g, b := render3d.RGB(colorFunc(c, model3d.RayCollision{}))
		return [3]float64{r, g, b}
	})
	mesh.SaveMaterialOBJ("cactus.zip", triColor)
}

func VaseSolid() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-1.0, -1.0, -1.0),
		model3d.XYZ(1.0, 1.0, 0.0),
		func(c model3d.Coord3D) bool {
			minRadius := 0.0
			radius := 0.5 + (c.Z+1.0)/4.0
			if c.Z > -0.15 {
				minRadius = radius - 0.15
			}
			if c.Z < -0.1 {
				radius -= 0.05
			}
			r := c.XY().Norm()
			return r < radius && r >= minRadius
		},
	)
}

func CactusSolid() model3d.Solid {
	var rightSide []model3d.Segment
	var leftSide []model3d.Segment
	curve := model2d.JoinedCurve{
		model2d.BezierCurve{model2d.XY(0, 0.5), model2d.XY(0.2, 0.5)},
		model2d.BezierCurve{model2d.XY(0.2, 0.5), model2d.XY(0.5, 0.5), model2d.XY(0.5, 0.8)},
	}
	for t := 0.0; t+0.01 < 1.0; t += 0.01 {
		p1 := curve.Eval(t)
		p2 := curve.Eval(t + 0.01)
		rightSide = append(
			rightSide,
			model3d.NewSegment(model3d.XZ(p1.X, p1.Y+0.1), model3d.XZ(p2.X, p2.Y+0.1)),
		)
		leftSide = append(
			leftSide,
			model3d.NewSegment(model3d.XZ(-p1.X, p1.Y-0.1), model3d.XZ(-p2.X, p2.Y-0.1)),
		)
	}
	return model3d.JoinedSolid{
		toolbox3d.LineJoin(0.2, model3d.NewSegment(model3d.Z(-0.5), model3d.Z(1.0))),
		toolbox3d.LineJoin(0.15, rightSide...),
		toolbox3d.LineJoin(0.15, leftSide...),
	}
}

func ColorFunc() render3d.ColorFunc {
	vase := model3d.MeshToSDF(model3d.MarchingCubesSearch(VaseSolid(), 0.03, 8))
	body := model3d.MeshToSDF(model3d.MarchingCubesSearch(CactusSolid(), 0.01, 8))

	return func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
		vFace, _, vDist := vase.FaceSDF(c)
		bFace, _, bDist := body.FaceSDF(c)
		if math.Abs(vDist) < math.Abs(bDist) {
			if vFace.Normal().Z > 0.99 && c.Z < -0.04 {
				return DirtColor
			}
			return VaseColor
		}
		normal := bFace.Normal()
		xz := normal.XZ().Norm()
		if normal.X < 0 {
			xz *= -1
		}
		theta := math.Atan2(normal.Y, xz)
		if int((theta+math.Pi*2)/(math.Pi/13))%2 == 0 {
			return CactusColor1
		} else {
			return CactusColor2
		}
	}
}
