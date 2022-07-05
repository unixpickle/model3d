package main

import (
	"flag"
	"log"

	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"

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

	HookThickness float64
	HookWidth     float64
	HookLength    float64
	HookCenter    bool

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
	flag.Float64Var(&flags.HookThickness, "hook-thickness", 0.02, "thickness of hook")
	flag.Float64Var(&flags.HookWidth, "hook-width", 0.075, "width of hook")
	flag.Float64Var(&flags.HookLength, "hook-length", 0.15, "length of hook")
	flag.BoolVar(&flags.HookCenter, "hook-center", false,
		"center the hook rather than using center of mass")
	flag.Float64Var(&flags.Delta, "delta", 0.0025, "marching cubes delta")
	flag.Parse()

	log.Println("Processing design...")
	pendant := RawPendantSolid(&flags)
	solid := model3d.JoinedSolid{
		pendant,
		CreateHook(&flags, pendant),
	}

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(solid, flags.Delta, 16)

	log.Println("Saving mesh...")
	mesh.SaveGroupedSTL("pendant.stl")

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, nil)
}

func CreateHook(f *Flags, pendant model3d.Solid) model3d.Solid {
	r := f.HookThickness / 2
	x := HookXLocation(f, pendant)
	yMax := HookYLocation(pendant, x)
	yMid := pendant.Min().Y - r
	yMin := yMid - f.HookLength - f.HookThickness
	zMax := f.HookWidth/2 + r
	return toolbox3d.LineJoin(
		r,
		model3d.NewSegment(model3d.XY(x, yMid), model3d.XY(x, yMax)),
		model3d.NewSegment(model3d.XYZ(x, yMid, -zMax), model3d.XYZ(x, yMid, zMax)),
		model3d.NewSegment(model3d.XYZ(x, yMin, -zMax), model3d.XYZ(x, yMin, zMax)),
		model3d.NewSegment(model3d.XYZ(x, yMid, zMax), model3d.XYZ(x, yMin, zMax)),
		model3d.NewSegment(model3d.XYZ(x, yMid, -zMax), model3d.XYZ(x, yMin, -zMax)),
	)
}

func HookXLocation(f *Flags, pendant model3d.Solid) float64 {
	if f.HookCenter {
		return 0
	}
	min, max := pendant.Min(), pendant.Max()
	var sum float64
	var numPoints int
	for numPoints < 2000 {
		p1 := model3d.NewCoord3DRandBounds(min, max)
		p2 := p1.Mul(model3d.XYZ(-1, 1, 1))
		for _, p := range []model3d.Coord3D{p1, p2} {
			if pendant.Contains(p) {
				sum += p.X
				numPoints++
			}
		}
	}
	return sum / float64(numPoints)
}

func HookYLocation(pendant model3d.Solid, x float64) float64 {
	min := pendant.Min()
	max := pendant.Max()
	for y := min.Y; y < max.Y; y += (max.X - min.X) / 2000.0 {
		if pendant.Contains(model3d.XY(x, y)) {
			return y
		}
	}
	panic("no place to connect pendant to hook")
}

func RoundedCylinder(p1, p2 model3d.Coord3D, radius float64) model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     p1,
			P2:     p2,
			Radius: radius,
		},
		&model3d.Sphere{
			Center: p1,
			Radius: radius,
		},
		&model3d.Sphere{
			Center: p2,
			Radius: radius,
		},
	}
}

func RawPendantSolid(f *Flags) model3d.Solid {
	if f.Image == "" {
		essentials.Die("must provide -image flag (see -help)")
	}
	image := model2d.MustReadBitmap(f.Image, nil).FlipX().Mesh().SmoothSq(f.SmoothIters)
	engraving := model2d.NewMesh()
	if f.Engraving != "" {
		engraving = model2d.MustReadBitmap(f.Engraving, nil).FlipX().Mesh().SmoothSq(f.SmoothIters)
	}

	min, max := image.Min(), image.Max()

	image = image.Translate(min.Mid(max).Scale(-1))
	engraving = engraving.Translate(min.Mid(max).Scale(-1))

	size := max.Sub(min)
	scale := f.Size / size.MaxCoord()
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
	delta := size.MaxCoord() / 1000.0
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
