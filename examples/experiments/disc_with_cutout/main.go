package main

import (
	"flag"
	"log"
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Args struct {
	DiscThickness             float64 `default:"0.1"`
	BaseThickness             float64 `default:"0.2"`
	ObstacleThickness         float64 `default:"0.1"`
	ObstacleGapToDisc         float64 `default:"0.1"`
	ObstacleHeight            float64 `default:"0.9"`
	GrooveDepth               float64 `default:"0.05"`
	PoleRadius                float64 `default:"0.07"`
	PoleOutsetX               float64 `default:"0.6"`
	PoleOutsetY               float64 `default:"0.4"`
	PoleHeight                float64 `default:"0.7"`
	ObstacleCutoutRadius      float64 `default:"0.09"`
	ObstacleCutoutExtraHeight float64 `default:"0.2"`
	DiscCutoutThickness       float64 `default:"0.12"`
	DiscHoverGap              float64 `default:"0.02"`
	DiscRadius                float64 `default:"1.0"`
	BaseSize                  float64 `default:"1.3"`

	Delta float64 `default:"0.01"`
}

func main() {
	var a Args
	toolbox3d.AddFlags(&a, nil)
	flag.Parse()

	log.Println("Creating disc component...")
	disc := CreateDisc(&a, true)
	pole := CreatePole(&a)
	part1 := model3d.JoinedSolid{disc, pole}
	mesh1 := model3d.DualContour(part1, a.Delta, true, false)

	log.Println("Creating board and obstacle component...")
	obstacle := CreateObstacle(&a)
	board := CreateBoard(&a)
	part2 := model3d.JoinedSolid{obstacle, board}
	mesh2 := model3d.DualContour(part2, a.Delta, true, false)

	log.Println("Decimating meshes...")
	mesh1 = mesh1.EliminateCoplanar(1e-5)
	mesh2 = mesh2.EliminateCoplanar(1e-5)

	log.Println("Saving meshes...")
	mesh1.SaveGroupedSTL("disc.stl")
	mesh2.SaveGroupedSTL("board.stl")

	joined := mesh1.Copy().Rotate(model3d.Z(1), 0.5)
	joined.AddMesh(mesh2)
	render3d.SaveRendering(
		"rendering1.png",
		joined,
		model3d.XYZ(1.0, -3.0, 2.0).Scale(a.BaseSize),
		512, 512, nil,
	)

	joined = mesh1.Copy().Rotate(model3d.Z(1), -0.3)
	joined.AddMesh(mesh2)
	render3d.SaveRendering(
		"rendering2.png",
		joined,
		model3d.XYZ(1.0, -3.0, 2.0).Scale(a.BaseSize),
		512, 512, nil,
	)
}

func CreateDisc(a *Args, cutoutForObstacle bool) model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-a.DiscRadius, -a.DiscRadius, -a.GrooveDepth),
		model3d.XYZ(a.DiscRadius, a.DiscRadius, a.DiscThickness),
		func(c model3d.Coord3D) bool {
			r := c.XY().Norm()
			if r > a.DiscRadius {
				return false
			} else if cutoutForObstacle && math.Abs(c.Y) < a.DiscCutoutThickness/2 && c.X > 0 {
				return false
			} else if c.Z < 0 {
				targetGroveRad := a.DiscRadius - a.GrooveDepth
				groveSize := c.Z + a.GrooveDepth
				return math.Abs(r-targetGroveRad) < groveSize
			} else {
				return true
			}
		},
	)
}

func CreatePole(a *Args) model3d.Solid {
	p1, p2 := polePoints(a)
	extendedP1 := p1.Add(p1.Sub(p2).Normalize())
	cyl := &model3d.Cylinder{
		P1:     extendedP1,
		P2:     p2,
		Radius: a.PoleRadius,
	}
	bounds := model3d.BoundsRect(cyl)
	bounds.MinVal.Z = p1.Z - 1e-5
	return model3d.IntersectedSolid{cyl, bounds}
}

func CreatePoleCutout(a *Args) model3d.Solid {
	p1, p2 := polePoints(a)
	p2 = p2.Add(p2.Sub(p1).Normalize().Scale(a.ObstacleCutoutExtraHeight))
	maxRadius := p1.XY().Norm() + a.ObstacleCutoutRadius
	return model3d.CheckedFuncSolid(
		model3d.XYZ(-maxRadius, -maxRadius, p1.Z),
		model3d.XYZ(maxRadius, maxRadius, p2.Z),
		func(c model3d.Coord3D) bool {
			zFrac := (c.Z - p1.Z) / (p2.Z - p1.Z)
			point := p1.Scale(1 - zFrac).Add(p2.Scale(zFrac))
			return math.Abs(c.XY().Norm()-point.XY().Norm()) < a.ObstacleCutoutRadius
		},
	)
}

func polePoints(a *Args) (model3d.Coord3D, model3d.Coord3D) {
	p1 := model3d.XYZ(-a.PoleOutsetX, -a.PoleOutsetY, a.DiscThickness)
	p2 := model3d.XYZ(-a.PoleOutsetX, a.PoleOutsetY, a.DiscThickness+a.PoleHeight)
	return p1, p2
}

func CreateObstacle(a *Args) model3d.Solid {
	mainRect := model3d.NewRect(
		model3d.XYZ(-a.BaseSize, -a.ObstacleThickness/2, a.DiscThickness+a.ObstacleGapToDisc),
		model3d.XYZ(0, a.ObstacleThickness/2, a.DiscThickness+a.ObstacleHeight),
	)
	connectRect := model3d.NewRect(
		model3d.XYZ(-a.BaseSize, -a.ObstacleThickness/2, -a.DiscHoverGap-1e-5),
		model3d.XYZ(-a.DiscRadius-a.ObstacleGapToDisc, a.ObstacleThickness/2, a.DiscThickness+a.ObstacleGapToDisc),
	)
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			mainRect,
			connectRect,
		},
		Negative: CreatePoleCutout(a),
	}
}

func CreateBoard(a *Args) model3d.Solid {
	return &model3d.SubtractedSolid{
		Positive: &model3d.Cylinder{
			P1:     model3d.Z(-a.BaseThickness),
			P2:     model3d.Z(-a.DiscHoverGap),
			Radius: a.BaseSize,
		},
		Negative: CreateDisc(a, false),
	}
}
