package main

import (
	_ "image/gif"
	"io/ioutil"
	"log"
	"math"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const (
	SquareSize       = 0.5
	SquareDepth      = 0.1
	SquarePoleLength = 0.18
	SquarePoleRadius = 0.05
	SquareHolderSize = 0.25
	SquareResolution = 10

	BoardBorder         = 0.3
	BoardInnerThickness = 0.1
	BoardSpacing        = 0.04

	PartSpacing = 0.04
)

const MinSpacing = 0.03

var DefaultColor = [3]float64{0.039, 0.729, 0.71}

func main() {
	log.Println("Creating board mesh...")
	puzzle := CreateBoard()
	log.Println("Creating piece mesh...")
	square := CreateSquare()

	log.Println("Combining meshes...")
	borderCollider := model3d.MeshToCollider(puzzle)

	for i := 0; i < 4; i++ {
		xDelta := float64(i+1)*BoardSpacing + float64(i)*SquareSize
		for j := 0; j < 4; j++ {
			if i == 0 && j == 0 {
				continue
			}
			yDelta := float64(j+1)*BoardSpacing + float64(j)*SquareSize
			delta := model3d.Coord3D{X: xDelta, Y: yDelta, Z: SquarePoleLength/2 + SquareDepth}
			piece := square.MapCoords(delta.Add)

			// Make sure the piece isn't too close to the border
			// or they might fuse together during printing.
			piece.Iterate(func(t *model3d.Triangle) {
				for _, p := range t {
					ray := &model3d.Ray{
						Origin:    p,
						Direction: model3d.Coord3D{X: 1.1, Y: 2.3, Z: 3.2},
					}
					if borderCollider.RayCollisions(ray)%2 == 1 ||
						borderCollider.SphereCollision(p, MinSpacing) {
						essentials.Die("undesired collision")
					}
				}
			})

			puzzle.AddMesh(piece)
		}
	}

	ioutil.WriteFile("puzzle.stl", puzzle.EncodeSTL(), 0755)
}

func CreateSquare() *model3d.Mesh {
	m := model3d.NewMesh()

	// Create the faces.
	for i := 0; i < SquareResolution; i++ {
		x1 := float64(i) / SquareResolution * SquareSize
		x2 := float64(i+1) / SquareResolution * SquareSize
		for j := 0; j < SquareResolution; j++ {
			y1 := float64(j) / SquareResolution * SquareSize
			y2 := float64(j+1) / SquareResolution * SquareSize
			p1 := model3d.Coord3D{X: x1, Y: y1, Z: 0}
			p2 := model3d.Coord3D{X: x2, Y: y1, Z: 0}
			p3 := model3d.Coord3D{X: x2, Y: y2, Z: 0}
			p4 := model3d.Coord3D{X: x1, Y: y2, Z: 0}

			// Create top face.
			m.Add(&model3d.Triangle{p1, p2, p3})
			m.Add(&model3d.Triangle{p1, p3, p4})

			// Create bottom face as well.
			for _, p := range []*model3d.Coord3D{&p1, &p2, &p3, &p4} {
				p.Z -= SquareDepth*2 + SquarePoleLength
				frac := SquareHolderSize / SquareSize
				p.X = SquareSize/2 + (p.X-SquareSize/2)*frac
				p.Y = SquareSize/2 + (p.Y-SquareSize/2)*frac
			}
			m.Add(&model3d.Triangle{p1, p3, p2})
			m.Add(&model3d.Triangle{p1, p4, p3})
		}
	}

	// Create the remainder of the square.
	m.Iterate(func(t *model3d.Triangle) {
		for i := 0; i < 3; i++ {
			p1 := t[i]
			p2 := t[(i+1)%3]
			if len(m.Find(p1, p2)) > 1 {
				continue
			}

			// Create the side of the square.
			p1, p2 = p2, p1
			p3, p4 := p2, p1
			if p1.Z == 0 {
				p3.Z -= SquareDepth
				p4.Z -= SquareDepth
			} else {
				p3.Z += SquareDepth
				p4.Z += SquareDepth
			}
			m.Add(&model3d.Triangle{p1, p2, p3})
			m.Add(&model3d.Triangle{p1, p3, p4})

			// Create the part of the square that connects
			// to the pole.
			polePoint := func(p model3d.Coord3D) model3d.Coord3D {
				angle := math.Atan2(p.Y-SquareSize/2, p.X-SquareSize/2)
				// Hack to fix rounding errors and make all the
				// vertices from the top and the bottom line up
				// perfectly.
				angle = math.Round(angle*123456.0) / 123456.0
				return model3d.Coord3D{
					X: math.Cos(angle)*SquarePoleRadius + SquareSize/2,
					Y: math.Sin(angle)*SquarePoleRadius + SquareSize/2,
					Z: p.Z,
				}
			}
			if p1.Z == 0 {
				p1, p2 = p4, p3
			} else {
				p1, p2 = p3, p4
			}
			p3 = polePoint(p2)
			p4 = polePoint(p1)
			m.Add(&model3d.Triangle{p1, p2, p3})
			m.Add(&model3d.Triangle{p1, p3, p4})

			// Create the pole.
			if p1.Z > -(SquareDepth + SquarePoleLength - 1e-4) {
				p1, p2 = p4, p3
				p3, p4 = p2, p1
				p3.Z = -(SquareDepth*2 + SquarePoleLength)
				p4.Z = -(SquareDepth*2 + SquarePoleLength)
				p3.Z += SquareDepth
				p4.Z += SquareDepth
				m.Add(&model3d.Triangle{p1, p2, p3})
				m.Add(&model3d.Triangle{p1, p3, p4})
			}
		}
	})

	return m
}

func CreateBoard() *model3d.Mesh {
	return model3d.SolidToMesh(BoardSolid{}, 0.05, 3, 0, 0).EliminateCoplanar(1e-8)
}

type BoardSolid struct{}

func (b BoardSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{
		X: -BoardBorder,
		Y: -BoardBorder,
		Z: -(SquarePoleLength/2 + SquareDepth + PartSpacing + BoardInnerThickness),
	}
}

func (b BoardSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{
		X: BoardBorder + BoardSpacing*5 + SquareSize*4,
		Y: BoardBorder + BoardSpacing*5 + SquareSize*4,
		Z: SquareDepth + SquarePoleLength/2,
	}
}

func (b BoardSolid) Contains(p model3d.Coord3D) bool {
	min := b.Min()
	max := b.Max()
	if min.Min(p) != min || max.Max(p) != max {
		return false
	}

	if p.X < 0 || p.Y < 0 || p.X > max.X-BoardBorder || p.Y > max.Y-BoardBorder {
		return true
	}

	if p.Z < min.Z+BoardInnerThickness {
		return true
	}

	for i := 0; i < 4; i++ {
		x := float64(i)*SquareSize + float64(i+1)*BoardSpacing + SquareSize/2
		radius := SquarePoleRadius + PartSpacing
		if p.Z < -BoardInnerThickness/2 {
			radius += (SquareSize - SquareHolderSize) / 2
		}
		if p.X > x-radius && p.X < x+radius {
			return false
		}
		if p.Y > x-radius && p.Y < x+radius {
			return false
		}
	}

	return p.Z < BoardInnerThickness/2
}
