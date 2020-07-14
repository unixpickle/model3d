package main

import (
	"math"
	"sort"

	"github.com/unixpickle/model3d/model2d"
)

// A Curve stores a simple curve describing the banana's
// arch.
// The banana itself is simply cross-sections along this
// curve, parallel to the curve's direction.
type Curve struct {
	// Sorted by minimum X, ascending.
	segments   []*model2d.Segment
	cumLengths []float64
	cumLength  float64

	pointSDF model2d.PointSDF

	min model2d.Coord
	max model2d.Coord
}

func NewCurve() *Curve {
	bezier := model2d.BezierCurve{
		model2d.XY(0, 1.5),
		model2d.XY(0.5, 0),
		model2d.XY(0.9, 1.9),
		model2d.XY(1, 2),
	}
	res := &Curve{}
	mesh := model2d.NewMesh()
	for x := 0; x < 10000; x++ {
		x1 := float64(x) / 10000
		y1 := bezier.EvalX(x1)
		x2 := float64(x+1) / 10000
		y2 := bezier.EvalX(x2)
		seg := &model2d.Segment{{X: x1 * BananaLength, Y: y1}, {X: x2 * BananaLength, Y: y2}}
		res.segments = append(res.segments, seg)
		res.cumLengths = append(res.cumLengths, res.cumLength)
		res.cumLength += seg.Length()
		mesh.Add(seg)
	}
	res.pointSDF = model2d.MeshToSDF(mesh)
	res.min = mesh.Min()
	res.max = mesh.Max()
	return res
}

// Min gets the minimum point on the curve.
func (c *Curve) Min() model2d.Coord {
	return c.min
}

// Min gets the maximum point on the curve.
func (c *Curve) Max() model2d.Coord {
	return c.max
}

// Project projects a 2D coordinate onto the curve by
// finding the nearest point on the curve.
//
// The t return value is the fraction of the arc-length of
// the curve (from left to right) at the projection.
//
// The d return value is the distance from coord to the
// projected point.
func (c *Curve) Project(coord model2d.Coord) (t, d float64) {
	p, signedDist := c.pointSDF.PointSDF(coord)
	return c.arcFractionForX(p.X), math.Abs(signedDist)
}

func (c *Curve) arcFractionForX(x float64) float64 {
	segIdx := sort.Search(len(c.segments), func(i int) bool {
		return c.segments[i][0].X > x
	})
	if segIdx == 0 {
		return 0
	}
	segIdx--
	seg := c.segments[segIdx]
	length := c.cumLengths[segIdx]
	length += seg.Length() * (x - seg[0].X) / (seg[1].X - seg[0].X)
	return length / c.cumLength
}
