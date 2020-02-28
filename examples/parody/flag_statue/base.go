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
	extra := model3d.Coord3D{X: 1, Y: 1}.Scale(BaseChunkSize)
	return model3d.IntersectedSolid{
		&model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -BaseLength / 2, Y: -BaseWidth / 2, Z: 0}.Sub(extra),
			MaxVal: model3d.Coord3D{X: BaseLength / 2, Y: BaseWidth / 2,
				Z: BaseHeight + BaseChunkSize}.Add(extra),
		},
		model3d.JoinedSolid{
			BaseSmoothSolid{},
			GenerateChunkyFinish(),
		},
	}
}

func GenerateChunkyFinish() model3d.Solid {
	var chunks model3d.JoinedSolid
	for i := 0; i < 500; i++ {
		center := SampleBasePoint()

		rotAxis := model3d.NewCoord3DRandUnit()
		ax1, ax2 := rotAxis.OrthoBasis()
		theta := rand.Float64() - 0.5
		sin, cos := math.Sin(theta), math.Cos(theta)
		ax1, ax2 = ax1.Scale(cos).Add(ax2.Scale(sin)), ax1.Scale(-sin).Add(ax2.Scale(cos))
		chunks = append(chunks, &BaseChunk{
			Axes:   [3]model3d.Coord3D{rotAxis, ax1, ax2},
			Center: center,
		})
	}

	// Chunk parts of the base together to avoid
	// redundant computations.
	essentials.VoodooSort(chunks, func(i, j int) bool {
		return chunks[i].Min().X < chunks[j].Min().X
	})
	var aggChunks model3d.JoinedSolid
	aggSize := len(chunks) / 20
	for i := 0; i < len(chunks); i += aggSize {
		subset := chunks[i:]
		if len(subset) > aggSize {
			subset = subset[:aggSize]
		}
		aggChunks = append(aggChunks, model3d.CacheSolidBounds(subset))
	}
	return aggChunks
}

func SampleBasePoint() model3d.Coord3D {
	var x, y, z float64
	for {
		// This is not quite uniform, but it is spread
		// out fairly nicely over the ellipsoid base.
		x = math.Tanh(rand.Float64()*4-2) * BaseLength / 2
		y = math.Tanh(rand.Float64()*4-2) * BaseWidth / 2
		z = math.Sqrt(1-(math.Pow(2*x/BaseLength, 2)+math.Pow(2*y/BaseWidth, 2))) * BaseHeight
		if !math.IsNaN(z) {
			break
		}
	}
	return model3d.Coord3D{X: x, Y: y, Z: z}
}

type BaseSmoothSolid struct{}

func (b BaseSmoothSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -BaseLength / 2, Y: -BaseWidth / 2}
}
func (b BaseSmoothSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BaseLength / 2, Y: BaseWidth / 2, Z: BaseHeight}
}

func (b BaseSmoothSolid) Contains(c model3d.Coord3D) bool {
	cScale := model3d.Coord3D{X: 2 / BaseLength, Y: 2 / BaseWidth, Z: 1 / BaseHeight}
	return model3d.InSolidBounds(b, c) && c.Mul(cScale).Norm() < 1
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
