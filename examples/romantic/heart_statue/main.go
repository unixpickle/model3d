package main

import (
	"log"
	"os"

	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	MarchingDelta = 0.0075

	BaseMinX = -0.4
	BaseMaxX = 1.0 - BaseMinX
	BaseMinY = -0.3
	BaseMaxY = -BaseMinY
	BaseMinZ = -0.05
	BaseMaxZ = 0.15

	LetterHeight       = 0.5
	LetterThickness    = 0.1
	LeftLetterXOffset  = -0.5
	RightLetterXOffset = 0.55
	LetterZ            = BaseMaxZ - 0.05
)

func main() {
	log.Println("Creating solid...")
	heart := CreateHeart()
	base := &model3d.Rect{
		MinVal: model3d.XYZ(BaseMinX, BaseMinY, BaseMinZ),
		MaxVal: model3d.XYZ(BaseMaxX, BaseMaxY, BaseMaxZ),
	}
	letter1 := LoadLetter("letter_1.png", 0.5+LeftLetterXOffset)
	letter2 := LoadLetter("letter_2.png", 0.5+RightLetterXOffset)

	fullSolid := model3d.JoinedSolid{
		heart,
		base,
		letter1,
		letter2,
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(fullSolid, MarchingDelta, 8)
	log.Println("Simplifying mesh...")
	mesh = mesh.EliminateCoplanar(1e-5)

	log.Println("Creating color func...")
	colorFunc := NewColorFunc(map[model3d.Solid][3]float64{
		heart:   {1.0, 0.0, 0.0},      // red
		base:    {1.0, 0.84, 0.0},     // gold
		letter1: {0.5, 0, 0.5},        // purple
		letter2: {0.504, 0.843, 0.81}, // tiffany blue
	})

	log.Println("Exporting model...")
	w, err := os.Create("color_heart_statue.zip")
	essentials.Must(err)
	defer w.Close()
	essentials.Must(mesh.SaveMaterialOBJ(
		"color_heart_statue.zip",
		colorFunc.TriangleColor,
	))

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, colorFunc.RenderColor)
}

func LoadLetter(filename string, x float64) model3d.Solid {
	mesh2d := model2d.MustReadBitmap(filename, nil).FlipY().Mesh().SmoothSq(50)
	mesh2d = mesh2d.Translate(mesh2d.Min().Scale(-1))
	mesh2d = mesh2d.Scale(LetterHeight / mesh2d.Max().Y)
	solid2d := model2d.NewColliderSolid(model2d.MeshToCollider(mesh2d))

	profile := model3d.ProfileSolid(solid2d, 0, LetterThickness)
	profile = model3d.TransformSolid(&model3d.Matrix3Transform{
		Matrix: &model3d.Matrix3{1.0, 0, 0, 0, 0.0, 1.0, 0, 1.0, 0.0},
	}, profile)
	profile = model3d.TranslateSolid(
		profile,
		model3d.XYZ(x-profile.Max().X/2, -LetterThickness/2, LetterZ),
	)

	return profile
}

func NewColorFunc(colors map[model3d.Solid][3]float64) toolbox3d.CoordColorFunc {
	var args []interface{}
	for solid, color := range colors {
		mesh := model3d.MarchingCubesSearch(solid, MarchingDelta, 8)
		args = append(args, mesh, render3d.NewColorRGB(color[0], color[1], color[2]))
	}
	return toolbox3d.JoinedCoordColorFunc(args...)
}
