package main

import (
	"math"
	"math/rand"

	"github.com/unixpickle/essentials"

	"github.com/unixpickle/model3d"
)

const (
	BaseWidth     = 2.5
	BaseLength    = 6.0
	BaseHeight    = 1.0
	BaseChunkSize = 0.2
)

func GenerateBase() model3d.Solid {
	return model3d.IntersectedSolid{
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -BaseLength / 2, Y: -BaseWidth / 2, Z: 0},
			MaxVal: model3d.Coord3D{X: BaseLength / 2, Y: BaseWidth / 2, Z: BaseHeight},
		},
		model3d.JoinedSolid{
			GenerateBasePolytope(),
			GenerateChunkyFinish(),
		},
	}
}

func GenerateBasePolytope() model3d.Solid {
	var result model3d.ConvexPolytope
	for i := 0; i < 1000; i++ {
		v, v1 := SampleBasePoint()
		result = append(result, &model3d.LinearConstraint{
			Normal: v,
			Max:    v1.Norm(),
		})
	}
	return &BasePolytopeSolid{P: result}
}

func GenerateChunkyFinish() model3d.Solid {
	var chunks model3d.JoinedSolid
	for i := 0; i < 300; i++ {
		_, v1 := SampleBasePoint()

		rotAxis := model3d.NewCoord3DRandUnit()
		ax1, ax2 := rotAxis.OrthoBasis()
		theta := rand.Float64() - 0.5
		sin, cos := math.Sin(theta), math.Cos(theta)
		ax1, ax2 = ax1.Scale(cos).Add(ax2.Scale(sin)), ax1.Scale(-sin).Add(ax2.Scale(cos))
		chunks = append(chunks, &BaseChunk{
			Axes:   [3]model3d.Coord3D{rotAxis, ax1, ax2},
			Center: v1,
		})
	}

	// Chunk parts of the base together to avoid
	// redundant computations.
	essentials.VoodooSort(chunks, func(i, j int) bool {
		return chunks[i].Min().X < chunks[j].Min().X
	})
	var aggChunks model3d.JoinedSolid
	aggSize := len(chunks) / 10
	for i := 0; i < len(chunks); i += aggSize {
		subset := chunks[i:]
		if len(subset) > aggSize {
			subset = subset[:aggSize]
		}
		aggChunks = append(aggChunks, model3d.CacheSolidBounds(subset))
	}
	return aggChunks
}

func SampleBasePoint() (spherical, base model3d.Coord3D) {
	spherical = model3d.NewCoord3DRandUnit()
	if spherical.Z < 0 {
		spherical = spherical.Scale(-1)
	}
	base = spherical.Mul(model3d.Coord3D{X: BaseLength / 2, Y: BaseWidth / 2, Z: BaseHeight})
	return
}

type BasePolytopeSolid struct {
	P model3d.ConvexPolytope
}

func (b BasePolytopeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -BaseLength / 2, Y: -BaseWidth / 2}
}
func (b BasePolytopeSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BaseLength / 2, Y: BaseWidth / 2, Z: BaseHeight}
}

func (b BasePolytopeSolid) Contains(c model3d.Coord3D) bool {
	return model3d.InSolidBounds(b, c) && b.P.Contains(c)
}

type BaseChunk struct {
	Axes   [3]model3d.Coord3D
	Center model3d.Coord3D
}

func (b *BaseChunk) Min() model3d.Coord3D {
	s := BaseChunkSize * math.Sqrt(3)
	return b.Center.Sub(model3d.Coord3D{X: s, Y: s, Z: s})
}

func (b *BaseChunk) Max() model3d.Coord3D {
	s := BaseChunkSize * math.Sqrt(3)
	return b.Center.Add(model3d.Coord3D{X: s, Y: s, Z: s})
}

func (b *BaseChunk) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(b, c) {
		return false
	}
	c = c.Sub(b.Center)
	for _, axis := range b.Axes {
		if math.Abs(c.Dot(axis)) >= BaseChunkSize/2 {
			return false
		}
	}
	return true
}
