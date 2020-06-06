package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	template, ok := FixedTemplates()[args.FixedTemplate]
	if !ok {
		fmt.Fprintln(os.Stderr, "unknown fixed template: "+args.FixedTemplate)
		flag.Usage()
		os.Exit(1)
	}

	log.Println("Searching for digit placements...")
	placements := SearchPlacement(template, AllDigits(), 5)
	if placements == nil {
		panic("no way to place digits")
	}

	log.Println("Creating board...")
	boardSolid := BoardSolid(&args, placements, 5)
	board := model3d.MarchingCubesSearch(boardSolid, 0.01, 8)
	board = board.EliminateCoplanar(1e-5)
	board.SaveGroupedSTL("board.stl")

	saveMesh := model3d.NewMesh()
	renderModel := render3d.JoinedObject{render3d.Objectify(board, nil)}
	for i, d := range placements {
		log.Println("Creating digit", i+1, "...")
		solid := DigitSolid(&args, d)
		ax := &toolbox3d.AxisSqueeze{
			Axis:  toolbox3d.AxisY,
			Min:   0.02,
			Max:   args.SegmentDepth - 0.02,
			Ratio: 0.25,
		}
		mesh := model3d.MarchingCubesSearch(model3d.TransformSolid(ax, solid), 0.01, 8)
		mesh = mesh.MapCoords(ax.Inverse().Apply)
		mesh = mesh.EliminateCoplanar(1e-5)

		saveMesh.AddMesh(mesh)

		color := render3d.NewColorRGB(rand.Float64(), rand.Float64(), rand.Float64())
		object := render3d.Objectify(mesh,
			func(c model3d.Coord3D, rc model3d.RayCollision) render3d.Color {
				return color
			})
		renderModel = append(renderModel, object)
	}

	render3d.SaveRendering("rendering.png", renderModel, model3d.Coord3D{X: 2.5, Y: -3, Z: 6},
		500, 500, nil)
	saveMesh.SaveGroupedSTL("digits.stl")
}

func FixedDigits() []Digit {
	// If you want to generate an arbitrary board, return nil.

	// Fill in three squares along the diagonal.
	// This makes the puzzle fairly difficult to solve.
	return []Digit{
		NewDigitContinuous([]Location{
			{2, 2},
			{2, 3},
			{3, 3},
			{3, 2},
			{2, 2},
		}),
		NewDigitContinuous([]Location{
			{0, 0},
			{0, 1},
			{1, 1},
			{1, 0},
			{0, 0},
		}),
		NewDigitContinuous([]Location{
			{4, 4},
			{4, 5},
			{5, 5},
			{5, 4},
			{4, 4},
		}),
	}
}
