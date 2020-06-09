package main

import (
	"log"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const (
	BoardWidth   = 5.0
	CutThickness = 0.03
	PieceDepth   = 0.2
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
}

type PiecesSolid struct {
	Cuts model2d.Solid
}

func (p *PiecesSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (p *PiecesSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BoardWidth, Y: BoardWidth, Z: PieceDepth}
}

func (p *PiecesSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(p, c) && !p.Cuts.Contains(c.XY())
}

func CutSolid() model2d.Solid {
	beziers := []model2d.BezierCurve{
		// Vertical lines.
		{
			{X: 0.2 * BoardWidth, Y: 0},
			{X: 0.03 * BoardWidth, Y: 0.3 * BoardWidth},
			{X: 0.3 * BoardWidth, Y: 0.6 * BoardWidth},
			{X: 0.2 * BoardWidth, Y: BoardWidth},
		},
		{
			{X: 0.4 * BoardWidth, Y: 0},
			{X: 0.3 * BoardWidth, Y: 0.2 * BoardWidth},
			{X: 0.6 * BoardWidth, Y: 0.4 * BoardWidth},
			{X: 0.38 * BoardWidth, Y: 0.9 * BoardWidth},
			{X: 0.4 * BoardWidth, Y: BoardWidth},
		},
		{
			{X: 0.5 * BoardWidth, Y: 0},
			{X: 0.7 * BoardWidth, Y: 0.4 * BoardWidth},
			{X: 0.6 * BoardWidth, Y: 0.5 * BoardWidth},
			{X: 0.6 * BoardWidth, Y: 0.8 * BoardWidth},
			{X: 0.69 * BoardWidth, Y: BoardWidth},
		},
		{
			{X: 0.9 * BoardWidth, Y: 0},
			{X: 0.8 * BoardWidth, Y: 0.4 * BoardWidth},
			{X: 0.7 * BoardWidth, Y: 0.5 * BoardWidth},
			{X: 0.75 * BoardWidth, Y: 0.8 * BoardWidth},
			{X: 0.8 * BoardWidth, Y: BoardWidth},
		},

		// Horizontal lines.
		{
			{X: 0, Y: 0.2 * BoardWidth},
			{X: 0.3 * BoardWidth, Y: 0.03 * BoardWidth},
			{X: 0.6 * BoardWidth, Y: 0.3 * BoardWidth},
			{X: BoardWidth, Y: 0.2 * BoardWidth},
		},
		{
			{X: 0, Y: 0.4 * BoardWidth},
			{X: 0.2 * BoardWidth, Y: 0.3 * BoardWidth},
			{X: 0.4 * BoardWidth, Y: 0.6 * BoardWidth},
			{X: 0.9 * BoardWidth, Y: 0.38 * BoardWidth},
			{X: BoardWidth, Y: 0.4 * BoardWidth},
		},
		{
			{X: 0, Y: 0.5 * BoardWidth},
			{X: 0.4 * BoardWidth, Y: 0.7 * BoardWidth},
			{X: 0.5 * BoardWidth, Y: 0.6 * BoardWidth},
			{X: 0.8 * BoardWidth, Y: 0.6 * BoardWidth},
			{X: BoardWidth, Y: 0.69 * BoardWidth},
		},
		{
			{X: 0, Y: 0.9 * BoardWidth},
			{X: 0.4 * BoardWidth, Y: 0.8 * BoardWidth},
			{X: 0.5 * BoardWidth, Y: 0.7 * BoardWidth},
			{X: 0.8 * BoardWidth, Y: 0.75 * BoardWidth},
			{X: BoardWidth, Y: 0.8 * BoardWidth},
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
