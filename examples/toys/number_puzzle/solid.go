package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	SegmentThickness   = 0.15
	SegmentDepth       = 0.1
	SegmentTipInset    = 0.03
	SegmentJointOutset = 0.015

	BoardThickness = 0.3
	BoardBorder    = 0.2
)

func BoardSolid(digits []Digit, size int) model3d.Solid {
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
		solids = append(solids, DigitSolid(Digit{s}))
	}
	solids = append(solids, &model3d.SubtractedSolid{
		Positive: &model3d.Rect{
			MinVal: model3d.Coord3D{X: -BoardBorder, Y: -BoardBorder, Z: -BoardThickness},
			MaxVal: model3d.Coord3D{X: float64(size) + BoardBorder, Y: float64(size) + BoardBorder,
				Z: SegmentDepth},
		},
		Negative: &model3d.Rect{
			MaxVal: model3d.Coord3D{X: float64(size), Y: float64(size),
				Z: SegmentDepth + 1e-5},
		},
	})

	return solids
}

func DigitSolid(d Digit) model3d.Solid {
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
			p1 = p1.Add(p2.Sub(p1).Normalize().Scale(SegmentTipInset))
		} else if points[s[1].Reflect(s[0])] != 0 {
			p1 = p1.Sub(p2.Sub(p1).Normalize().Scale(SegmentJointOutset))
		}
		if points[s[1]] == 1 {
			p2 = p2.Add(p1.Sub(p2).Normalize().Scale(SegmentTipInset))
		} else if points[s[0].Reflect(s[1])] != 0 {
			p2 = p2.Sub(p1.Sub(p2).Normalize().Scale(SegmentJointOutset))
		}

		segments = append(segments, &pointedSegment{
			P1:       p1,
			P2:       p2,
			Vertical: s[0][0] == s[1][0],
		})
	}

	return segments
}

type pointedSegment struct {
	P1       model3d.Coord3D
	P2       model3d.Coord3D
	Vertical bool
}

func (p *pointedSegment) Min() model3d.Coord3D {
	res := p.P1.Min(p.P2)
	if p.Vertical {
		res.X -= SegmentThickness / 2
	} else {
		res.Y -= SegmentThickness / 2
	}
	return res
}

func (p *pointedSegment) Max() model3d.Coord3D {
	res := p.P1.Max(p.P2)
	if p.Vertical {
		res.X += SegmentThickness / 2
	} else {
		res.Y += SegmentThickness / 2
	}
	res.Z += SegmentDepth
	return res
}

func (p *pointedSegment) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}

	tip := SegmentThickness / 2
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
