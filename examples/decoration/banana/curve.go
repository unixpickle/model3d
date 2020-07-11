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
	return res
}

// Project projects a 2D coordinate onto the curve.
//
// It returns the fraction of the arclength along the
// curve of the projected point, as well as the distance
// to the curve.
//
// The collides return value is false if the projection
// falls outside of the curve.
func (c *Curve) Project(coord model2d.Coord) (t, d float64, collides bool) {
	p, _ := c.pointSDF.PointSDF(coord)
	d = coord.Dist(p)
	seg, t := c.lookupX(p.X)
	collides = true
	if d > 1e-5 {
		// If we are beyond the bounds of the curve,
		// then the projection to it won't be normal.
		normalDot := seg.Normal().Dot(coord.Sub(p).Normalize())
		if math.Abs(normalDot) < 0.99 {
			collides = false
		}
	}
	return
}

func (c *Curve) lookupX(x float64) (*model2d.Segment, float64) {
	segIdx := sort.Search(len(c.segments), func(i int) bool {
		return c.segments[i][0].X > x
	})
	if segIdx == 0 {
		return c.segments[0], 0
	}
	segIdx--
	seg := c.segments[segIdx]
	length := c.cumLengths[segIdx]
	length += seg.Length() * (x - seg[0].X) / (seg[1].X - seg[0].X)
	return seg, length / c.cumLength
}
