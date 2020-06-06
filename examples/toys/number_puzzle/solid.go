package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
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
	segmentSet := map[Segment]bool{}
	for _, s := range d {
		segmentSet[s] = true
		for _, l := range s {
			points[l] += 1
		}
	}

	var segments2d model2d.JoinedSolid
	for _, s := range d {
		p1 := model3d.Coord2D{X: float64(s[0][0]), Y: float64(s[0][1])}
		p2 := model3d.Coord2D{X: float64(s[1][0]), Y: float64(s[1][1])}

		// Move tips inward and connected points outward.
		if points[s[0]] == 1 {
			p1 = p1.Add(p2.Sub(p1).Normalize().Scale(a.SegmentTipInset))
		} else if segmentSet[NewSegment(s[0], s[1].Reflect(s[0]))] {
			p1 = p1.Sub(p2.Sub(p1).Normalize().Scale(a.SegmentJointOutset))
		}
		if points[s[1]] == 1 {
			p2 = p2.Add(p1.Sub(p2).Normalize().Scale(a.SegmentTipInset))
		} else if segmentSet[NewSegment(s[1], s[0].Reflect(s[1]))] {
			p2 = p2.Sub(p1.Sub(p2).Normalize().Scale(a.SegmentJointOutset))
		}

		segments2d = append(segments2d, &pointedSegment{
			Args:     a,
			P1:       p1,
			P2:       p2,
			Vertical: s[0][0] == s[1][0],
		})
	}

	mesh2d := model2d.MarchingSquaresSearch(segments2d, 0.005, 8)
	collider2d := model2d.MeshToCollider(mesh2d)
	solid2d := model2d.NewColliderSolidInset(collider2d, a.SegmentInset)
	return &segmentProfile3D{
		Args:    a,
		Profile: solid2d,
	}
}

type segmentProfile3D struct {
	Args    *Args
	Profile model2d.Solid
}

func (s *segmentProfile3D) Min() model3d.Coord3D {
	m := s.Profile.Min()
	return model3d.Coord3D{X: m.X, Y: m.Y}
}

func (s *segmentProfile3D) Max() model3d.Coord3D {
	m := s.Profile.Max()
	return model3d.Coord3D{X: m.X, Y: m.Y, Z: s.Args.SegmentDepth}
}

func (s *segmentProfile3D) Contains(c model3d.Coord3D) bool {
	return model3d.InBounds(s, c) && s.Profile.Contains(c.Coord2D())
}

type pointedSegment struct {
	Args     *Args
	P1       model2d.Coord
	P2       model2d.Coord
	Vertical bool
}

func (p *pointedSegment) Min() model2d.Coord {
	res := p.P1.Min(p.P2)
	if p.Vertical {
		res.X -= p.Args.SegmentThickness / 2
	} else {
		res.Y -= p.Args.SegmentThickness / 2
	}
	return res
}

func (p *pointedSegment) Max() model2d.Coord {
	res := p.P1.Max(p.P2)
	if p.Vertical {
		res.X += p.Args.SegmentThickness / 2
	} else {
		res.Y += p.Args.SegmentThickness / 2
	}
	return res
}

func (p *pointedSegment) Contains(c model2d.Coord) bool {
	if !model2d.InBounds(p, c) {
		return false
	}

	tip := p.Args.SegmentThickness / 2
	axis := p.P1.Sub(p.P2).Normalize()
	tipDist := math.Min(
		math.Abs(axis.Dot(c)-axis.Dot(p.P1)),
		math.Abs(axis.Dot(c)-axis.Dot(p.P2)),
	)
	if tipDist < tip {
		tipInset := tip - tipDist
		sideDist := math.Abs(c.Y - p.P1.Y)
		if p.Vertical {
			sideDist = math.Abs(c.X - p.P1.X)
		}
		// Add a small epsilon so that segments touching at a
		// 90 degree angle definitely intersect.
		if sideDist+tipInset > tip+1e-5 {
			return false
		}
	}
	return true
}
