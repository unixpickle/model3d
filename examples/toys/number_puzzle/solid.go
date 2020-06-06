package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

func BoardSolid(a *Args, digits []Digit, size int) model3d.Solid {
	segments := map[Segment]bool{}
	for x := 0; x <= size; x++ {
		for y := 0; y <= size; y++ {
			l := Location{y, x}
			if x < size {
				segments[NewSegment(l, Location{y, x + 1})] = true
			}
			if y < size {
				segments[NewSegment(l, Location{y + 1, x})] = true
			}
		}
	}
	for _, d := range digits {
		for _, s := range d {
			delete(segments, s)
		}
	}

	var solids model3d.JoinedSolid
	for s := range segments {
		solids = append(solids, DigitSolid(a, Digit{s}))
	}
	border := a.BoardBorder + a.SegmentThickness/2
	solids = append(solids, &model3d.SubtractedSolid{
		Positive: &model3d.Rect{
			MinVal: model3d.Coord3D{X: -border, Y: -border, Z: -a.BoardThickness},
			MaxVal: model3d.Coord3D{X: float64(size) + border, Y: float64(size) + border,
				Z: a.SegmentDepth},
		},
		Negative: &model3d.Rect{
			MinVal: model3d.Coord3D{
				X: -a.SegmentThickness / 2,
				Y: -a.SegmentThickness / 2,
			},
			MaxVal: model3d.Coord3D{
				X: float64(size) + a.SegmentThickness/2,
				Y: float64(size) + a.SegmentThickness/2,
				Z: a.SegmentDepth + 1e-5,
			},
		},
	})

	return solids
}

func DigitSolid(a *Args, d Digit) model3d.Solid {
	points := map[Location]int{}
	for _, s := range d {
		for _, l := range s {
			points[l] += 1
		}
	}

	var segments model3d.JoinedSolid
	for _, s := range d {
		p1 := model3d.Coord3D{X: float64(s[0][0]), Y: float64(s[0][1])}
		p2 := model3d.Coord3D{X: float64(s[1][0]), Y: float64(s[1][1])}

		// Move tips inward and connected points outward.
		if points[s[0]] == 1 {
			p1 = p1.Add(p2.Sub(p1).Normalize().Scale(a.SegmentTipInset))
		} else if points[s[1].Reflect(s[0])] != 0 {
			p1 = p1.Sub(p2.Sub(p1).Normalize().Scale(a.SegmentJointOutset))
		}
		if points[s[1]] == 1 {
			p2 = p2.Add(p1.Sub(p2).Normalize().Scale(a.SegmentTipInset))
		} else if points[s[0].Reflect(s[1])] != 0 {
			p2 = p2.Sub(p1.Sub(p2).Normalize().Scale(a.SegmentJointOutset))
		}

		segments = append(segments, &pointedSegment{
			Args:     a,
			P1:       p1,
			P2:       p2,
			Vertical: s[0][0] == s[1][0],
		})
	}

	return segments
}

type pointedSegment struct {
	Args     *Args
	P1       model3d.Coord3D
	P2       model3d.Coord3D
	Vertical bool
}

func (p *pointedSegment) Min() model3d.Coord3D {
	res := p.P1.Min(p.P2)
	if p.Vertical {
		res.X -= p.Args.SegmentThickness / 2
	} else {
		res.Y -= p.Args.SegmentThickness / 2
	}
	return res
}

func (p *pointedSegment) Max() model3d.Coord3D {
	res := p.P1.Max(p.P2)
	if p.Vertical {
		res.X += p.Args.SegmentThickness / 2
	} else {
		res.Y += p.Args.SegmentThickness / 2
	}
	res.Z += p.Args.SegmentDepth
	return res
}

func (p *pointedSegment) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}

	tip := p.Args.SegmentThickness / 2
	c2 := c.Coord2D()
	axis := p.P1.Sub(p.P2).Normalize()
	tipDist := math.Min(
		math.Abs(axis.Dot(c)-axis.Dot(p.P1)),
		math.Abs(axis.Dot(c)-axis.Dot(p.P2)),
	)
	if tipDist < tip {
		tipInset := tip - tipDist
		sideDist := math.Abs(c2.Y - p.P1.Y)
		if p.Vertical {
			sideDist = math.Abs(c2.X - p.P1.X)
		}
		// Add a small epsilon so that segments touching at a
		// 90 degree angle definitely intersect.
		if sideDist+tipInset > tip+1e-5 {
			return false
		}
	}
	return true
}
