package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func main() {
	tower := CastleTower()
	entrance := EntranceBuilding()
	sandPart := model3d.JoinedSolid{
		model3d.TranslateSolid(tower, model3d.XY(0.55, 0.525)),
		model3d.TranslateSolid(tower, model3d.XY(0.55, -0.525)),
		model3d.TranslateSolid(tower, model3d.XY(-0.55, 0.525)),
		model3d.TranslateSolid(tower, model3d.XY(-0.55, -0.525)),
		model3d.TranslateSolid(entrance, model3d.Y(1.2)),
		model3d.TranslateSolid(
			model3d.RotateSolid(entrance, model3d.Z(1), math.Pi),
			model3d.Y(-1.2),
		),
		SandLump(),
	}
	cutout := &model3d.SubtractedSolid{
		Positive: sandPart.Optimize(),
		Negative: model3d.JoinedSolid{
			StairCutout(),
			TextCutout(),
		},
	}
	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(cutout, 0.008, 8)
	log.Println("Decimating mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)
	log.Println("Rendering...")
	colorFn := toolbox3d.ConstantCoordColorFunc(render3d.NewColorRGB(0.92, 0.87, 0.62))
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFn.RenderColor)
	log.Println("Saving...")
	mesh.SaveMaterialOBJ("sand_castle.zip", colorFn.TriangleColor)
}

func CastleTower() model3d.Solid {
	basicShape := model3d.CheckedFuncSolid(
		model3d.XYZ(-0.5, -0.5, 0),
		model3d.XYZ(0.5, 0.5, 1.1),
		func(c model3d.Coord3D) bool {
			z := c.Z
			rad := 0.5 - z*0.2
			cRad := c.XY().Norm()
			theta := math.Atan2(c.X, c.Y)

			if cRad > rad {
				return false
			}

			for i := 0; i < 5; i++ {
				startZ := 0.05 + float64(i)/5.3
				endZ := startZ + 0.2
				if (c.Z < startZ && i > 0) || (c.Z > endZ && i < 5) {
					continue
				}
				indent := 0.015
				indentSqueeze := 0.5
				zLineDist := math.Min(math.Abs(c.Z-startZ), math.Abs(c.Z-endZ))
				zIndent := indentSqueeze * math.Sqrt(math.Max(0, indent*indent-zLineDist*zLineDist))
				if zIndent > 0 && cRad > rad-zIndent {
					return false
				}

				numBricks := 14 - i
				brickTheta := math.Pi * 2 / float64(numBricks)
				offset := 0.0
				if i%2 == 1 {
					offset += brickTheta / 2
				}
				for j := 0; j < numBricks; j++ {
					startTheta := float64(j) * brickTheta
					dist := toolbox3d.AngleDist(theta, startTheta) * rad
					indent := indentSqueeze * math.Sqrt(math.Max(0, indent*indent-dist*dist))
					if indent > 0 && cRad > rad-indent {
						return false
					}
				}
			}

			if c.Z < 1.0 {
				return true
			}

			if cRad < rad-0.1 {
				return false
			}
			sizeAngle := 0.2 / rad
			for i := 0; i < 4; i++ {
				if toolbox3d.AngleDist(theta, float64(i)*math.Pi*2.0/4.0) < sizeAngle/2 {
					return true
				}
			}
			return false
		},
	)
	return basicShape
}

func EntranceBuilding() model3d.Solid {
	shape := model3d.JoinedSolid{
		model3d.NewRect(model3d.XY(-0.5, -0.15), model3d.XYZ(0.5, 0.15, 0.3)),
		&model3d.Capsule{P1: model3d.Y(0.1), P2: model3d.YZ(0.1, 0.3), Radius: 0.15},
	}
	sdf := model3d.DualContourSDF(shape, 0.015)
	solid := model3d.SDFToSolid(sdf, 0.05)

	// Make the building slightly wider at the base.
	return model3d.CheckedFuncSolid(
		solid.Min(),
		solid.Max(),
		func(c model3d.Coord3D) bool {
			xy := c.XY().Scale(1.0 + c.Z*0.1)
			return solid.Contains(model3d.XYZ(xy.X, xy.Y, c.Z))
		},
	)
}

func SandLump() model3d.Solid {
	dropRadius := 0.6
	edgeBounds := model2d.MeshToSDF(
		model2d.MarchingSquaresSearch(
			model2d.NewColliderSolidInset(
				model2d.NewRect(
					model2d.XY(-2.0+dropRadius, -2.0+dropRadius),
					model2d.XY(2.0-dropRadius, 2.0-dropRadius),
				),
				-dropRadius,
			), 0.1, 8,
		),
	)
	edgeDropoff := model2d.BezierCurve{
		model2d.XY(0.0, 0),
		model2d.XY(dropRadius/2, 0),
		model2d.XY(dropRadius, -0.5),
	}
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-2.0, -2.0, -0.5),
		model3d.XYZ(2.0, 2.0, 0.3),
		func(c model3d.Coord3D) bool {
			dist := dropRadius - edgeBounds.SDF(c.XY())
			dropoff := 0.0
			if dist > 0 {
				if dist > dropRadius {
					return false
				}
				dropoff = edgeDropoff.EvalX(dist)
			}
			x := c.X * 3.5
			y := c.Y * 3.5
			z := 0.15*math.Sqrt(2+2*(0.1*math.Cos(3*(x+0.5*math.Sin(y*2+6)))+0.2*math.Cos(1*(2*y-math.Sqrt(x+10)))+0.2*math.Cos(0.1+x*2)+0.2*math.Cos(0.3+y*1.7)+0.1*math.Cos(2*math.Sqrt(x*x+y*y)))) + 0.15*math.Exp(-math.Pow(x*x+y*y, 2)/50)
			z += dropoff - 0.15
			return c.Z < z
		},
	)
}

func StairCutout() model3d.Solid {
	stepSize := 0.1
	steps := 4.0
	oneSide := model3d.CheckedFuncSolid(
		model3d.XYZ(-0.3, -2.0, -0.4),
		model3d.XYZ(0.3, -2.0+stepSize*steps, 0.2),
		func(c model3d.Coord3D) bool {
			stepY := math.Floor((c.Y+2.0)/stepSize) * stepSize
			return c.Z > -0.4+stepY
		},
	)
	return model3d.JoinedSolid{
		oneSide,
		model3d.RotateSolid(oneSide, model3d.Z(1), math.Pi),
	}
}

func TextCutout() model3d.Solid {
	textMesh := model2d.MustReadBitmap("text_imprint.png", nil).FlipX().Mesh().SmoothSq(20)
	textMesh = textMesh.Translate(textMesh.Max().Mid(textMesh.Min()).Scale(-1))
	textMesh = textMesh.Scale(1 / textMesh.Max().X)
	textCollider2d := model2d.MeshToCollider(textMesh)
	textCollider3d := model3d.ProfileCollider(textCollider2d, -0.075, 0.075)
	textSolid3d := model3d.NewColliderSolidInset(textCollider3d, -0.01)
	return model3d.RotateSolid(
		model3d.TranslateSolid(
			model3d.RotateSolid(textSolid3d, model3d.X(1), -math.Pi/4),
			model3d.YZ(1.75, -0.1),
		),
		model3d.Z(1), -math.Pi/2,
	)
}
