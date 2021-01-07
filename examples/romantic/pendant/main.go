package main

import (
	"flag"
	"log"
	"math"

	"github.com/unixpickle/model3d/render3d"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

type Flags struct {
	Image       string
	Engraving   string
	SmoothIters int

	Size           float64
	Rounding       float64
	Thickness      float64
	EngravingDepth float64

	Delta float64
}

func main() {
	var flags Flags
	flag.StringVar(&flags.Image, "image", "", "image file for pendant")
	flag.StringVar(&flags.Engraving, "engraving", "", "optional image file for engraving")
	flag.IntVar(&flags.SmoothIters, "smooth-iters", 50, "number of 2D mesh smoothing steps")
	flag.Float64Var(&flags.Size, "size", 0.7, "size of largest dimension")
	flag.Float64Var(&flags.Rounding, "rounding", 0.01, "amount of rounding to apply")
	flag.Float64Var(&flags.Thickness, "thickness", 0.05, "thickness of pendant")
	flag.Float64Var(&flags.EngravingDepth, "engraving-depth", 0.01, "depth of engraving")
	flag.Float64Var(&flags.Delta, "delta", 0.0025, "marching cubes delta")
	flag.Parse()

	log.Println("Processing design...")
	solid := RawPendantSolid(&flags)

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, flags.Delta, 16)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("pendant.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func RawPendantSolid(f *Flags) model3d.Solid {
	if f.Image == "" {
		essentials.Die("must provide -image flag (see -help)")
	}
	image := model2d.MustReadBitmap(f.Image, nil).FlipY().Mesh().SmoothSq(f.SmoothIters)
	engraving := model2d.NewMesh()
	if f.Engraving != "" {
		engraving = model2d.MustReadBitmap(f.Engraving, nil).FlipY().Mesh().SmoothSq(f.SmoothIters)
	}

	min, max := image.Min(), image.Max()

	image = image.MapCoords(min.Mid(max).Scale(-1).Add)
	engraving = engraving.MapCoords(min.Mid(max).Scale(-1).Add)

	size := max.Sub(min)
	scale := f.Size / math.Max(size.X, size.Y)
	image = image.Scale(scale)
	engraving = engraving.Scale(scale)

	image = InsetShape(image, f.Rounding)
	engravingSolid := model3d.ProfileSolid(
		model2d.NewColliderSolid(model2d.MeshToCollider(engraving)),
		f.Thickness/2-f.EngravingDepth,
		f.Thickness/2+1e-5,
	)

	return &model3d.SubtractedSolid{
		Positive: ShapeToRoundedCollider(image, f.Thickness, f.Rounding),
		Negative: engravingSolid,
	}
}

func InsetShape(shape *model2d.Mesh, amount float64) *model2d.Mesh {
	size := shape.Max().Sub(shape.Min())
	delta := math.Max(size.X, size.Y) / 1000.0
	solid := model2d.NewColliderSolidInset(model2d.MeshToCollider(shape), amount)
	mesh := model2d.MarchingSquaresSearch(solid, delta, 8)
	return mesh
}

func ShapeToRoundedCollider(shape *model2d.Mesh, thickness, outset float64) model3d.Solid {
	collider2d := model2d.MeshToCollider(shape)
	solid2d := model2d.NewColliderSolid(collider2d)
	th := thickness/2 - outset
	collider3d := model3d.ProfileCollider(collider2d, -th, th)
	solid3d := model3d.ProfileSolid(solid2d, -th, th)
	return model3d.JoinedSolid{
		solid3d,
		model3d.NewColliderSolidHollow(collider3d, outset),
	}
}
