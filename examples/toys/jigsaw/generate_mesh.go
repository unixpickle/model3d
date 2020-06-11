package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	BoardSize    = 5.0
	CutThickness = 0.02
	PieceDepth   = 0.3

	HolderBorder = 0.2
	HolderDepth  = 0.3
	HolderInset  = 0.15
)

func GenerateMesh() {
	log.Println("Creating solid...")
	pieces := &PiecesSolid{
		Cuts: CutSolid(),
	}
	log.Println("Creating mesh...")
	piecesMesh := model3d.MarchingCubesSearch(pieces, 0.02, 8)
	log.Println("Simplifying mesh...")
	piecesMesh = piecesMesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	piecesMesh.SaveGroupedSTL("pieces.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_pieces.png", piecesMesh, 3, 3, 300, nil)

	log.Println("Creating holder mesh...")
	holder := &model3d.SubtractedSolid{
		Positive: &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -HolderBorder, Y: -HolderBorder,
				Z: -(HolderDepth - HolderInset)},
			MaxVal: model3d.Coord3D{X: BoardSize + HolderBorder, Y: BoardSize + HolderBorder,
				Z: HolderInset},
		},
		Negative: &model3d.RectSolid{
			MinVal: model3d.Coord3D{},
			MaxVal: model3d.Coord3D{X: BoardSize, Y: BoardSize, Z: HolderInset + 1e-5},
		},
	}
	holderMesh := model3d.MarchingCubesSearch(holder, 0.02, 8)
	log.Println("Simplifying mesh...")
	holderMesh = holderMesh.EliminateCoplanar(1e-5)
	log.Println("Saving mesh...")
	holderMesh.SaveGroupedSTL("holder.stl")
	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering_holder.png", holderMesh, 3, 3, 300, nil)
}

type PiecesSolid struct {
	Cuts model2d.Solid
}

func (p *PiecesSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (p *PiecesSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BoardSize, Y: BoardSize, Z: PieceDepth}
}

func (p *PiecesSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(p, c) && !p.Cuts.Contains(c.XY())
}

func CutSolid() model2d.Solid {
	beziers := []model2d.BezierCurve{
		// Vertical lines.
		{
			{X: 0.2 * BoardSize, Y: -0.05 * BoardSize},
			{X: 0.03 * BoardSize, Y: 0.3 * BoardSize},
			{X: 0.3 * BoardSize, Y: 0.6 * BoardSize},
			{X: 0.2 * BoardSize, Y: 1.05 * BoardSize},
		},
		{
			{X: 0.4 * BoardSize, Y: -0.05 * BoardSize},
			{X: 0.3 * BoardSize, Y: 0.2 * BoardSize},
			{X: 0.6 * BoardSize, Y: 0.4 * BoardSize},
			{X: 0.38 * BoardSize, Y: 0.9 * BoardSize},
			{X: 0.4 * BoardSize, Y: 1.05 * BoardSize},
		},
		{
			{X: 0.5 * BoardSize, Y: -0.05 * BoardSize},
			{X: 0.7 * BoardSize, Y: 0.4 * BoardSize},
			{X: 0.6 * BoardSize, Y: 0.5 * BoardSize},
			{X: 0.6 * BoardSize, Y: 0.8 * BoardSize},
			{X: 0.69 * BoardSize, Y: 1.05 * BoardSize},
		},
		{
			{X: 0.9 * BoardSize, Y: -0.05 * BoardSize},
			{X: 0.8 * BoardSize, Y: 0.4 * BoardSize},
			{X: 0.7 * BoardSize, Y: 0.5 * BoardSize},
			{X: 0.75 * BoardSize, Y: 0.8 * BoardSize},
			{X: 0.8 * BoardSize, Y: 1.05 * BoardSize},
		},

		// Horizontal lines.
		{
			{X: -0.05 * BoardSize, Y: 0.2 * BoardSize},
			{X: 0.3 * BoardSize, Y: 0.03 * BoardSize},
			{X: 0.6 * BoardSize, Y: 0.3 * BoardSize},
			{X: 1.05 * BoardSize, Y: 0.2 * BoardSize},
		},
		{
			{X: -0.05 * BoardSize, Y: 0.4 * BoardSize},
			{X: 0.2 * BoardSize, Y: 0.3 * BoardSize},
			{X: 0.4 * BoardSize, Y: 0.6 * BoardSize},
			{X: 0.9 * BoardSize, Y: 0.38 * BoardSize},
			{X: 1.05 * BoardSize, Y: 0.4 * BoardSize},
		},
		{
			{X: -0.05 * BoardSize, Y: 0.5 * BoardSize},
			{X: 0.4 * BoardSize, Y: 0.7 * BoardSize},
			{X: 0.5 * BoardSize, Y: 0.6 * BoardSize},
			{X: 0.8 * BoardSize, Y: 0.6 * BoardSize},
			{X: 1.05 * BoardSize, Y: 0.69 * BoardSize},
		},
		{
			{X: -0.05 * BoardSize, Y: 0.9 * BoardSize},
			{X: 0.4 * BoardSize, Y: 0.8 * BoardSize},
			{X: 0.5 * BoardSize, Y: 0.7 * BoardSize},
			{X: 0.8 * BoardSize, Y: 0.75 * BoardSize},
			{X: 1.05 * BoardSize, Y: 0.8 * BoardSize},
		},
	}
	mesh2d := model2d.NewMesh()
	for t := 0.01; t < 1.0; t += 0.01 {
		for _, b := range beziers {
			p1 := b.Eval(t - 0.01)
			p2 := b.Eval(t)
			mesh2d.Add(&model2d.Segment{p1, p2})
		}
	}
	return model2d.NewColliderSolidHollow(model2d.MeshToCollider(mesh2d), CutThickness)
}
