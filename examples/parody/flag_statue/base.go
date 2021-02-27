package main

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model3d"
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
		&model3d.Rect{
			MinVal: model3d.XYZ(-BaseLength/2, -BaseWidth/2, 0).Sub(extra),
			MaxVal: model3d.Coord3D{X: BaseLength / 2, Y: BaseWidth / 2,
				Z: BaseHeight + BaseChunkSize}.Add(extra),
		},
		model3d.JoinedSolid{
			BaseSmoothSolid(),
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
	return chunks.Optimize()
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
	return model3d.XYZ(x, y, z)
}

func BaseSmoothSolid() model3d.Solid {
	return model3d.CheckedFuncSolid(
		model3d.XY(-BaseLength/2, -BaseWidth/2),
		model3d.XYZ(BaseLength/2, BaseWidth/2, BaseHeight),
		func(c model3d.Coord3D) bool {
			cScale := model3d.XYZ(2/BaseLength, 2/BaseWidth, 1/BaseHeight)
			return c.Mul(cScale).Norm() < 1
		},
	)
}

type BaseChunk struct {
	Axes   [3]model3d.Coord3D
	Center model3d.Coord3D
}

func (b *BaseChunk) Min() model3d.Coord3D {
	s := BaseChunkSize * math.Sqrt(3)
	return b.Center.Sub(model3d.XYZ(s, s, s))
}

func (b *BaseChunk) Max() model3d.Coord3D {
	s := BaseChunkSize * math.Sqrt(3)
	return b.Center.Add(model3d.XYZ(s, s, s))
}

func (b *BaseChunk) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
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
