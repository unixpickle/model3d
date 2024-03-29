package model2d

import "github.com/unixpickle/model3d/numerical"

type indexMesh struct {
	Coords   []Coord
	Segments [][2]int
}

func newIndexMesh(m *Mesh) *indexMesh {
	var coords []Coord
	var segments [][2]int
	coordToIndex := map[Coord]int{}
	m.Iterate(func(s *Segment) {
		var indices [2]int
		for i, c := range s {
			if idx, ok := coordToIndex[c]; ok {
				indices[i] = idx
			} else {
				idx = len(coords)
				coordToIndex[c] = idx
				coords = append(coords, c)
				indices[i] = idx
			}
		}
		segments = append(segments, indices)
	})
	return &indexMesh{
		Coords:   coords,
		Segments: segments,
	}
}

func (i *indexMesh) Smooth(squares bool) {
	grad := smoothingGradient(i, squares)
	stepSize := optimalSmoothingStepSize(i, grad, squares)
	for j, c := range i.Coords {
		i.Coords[j] = c.Add(grad[j].Scale(stepSize))
	}
}

func (i *indexMesh) Mesh() *Mesh {
	res := NewMesh()
	for _, seg := range i.Segments {
		res.Add(&Segment{i.Coords[seg[0]], i.Coords[seg[1]]})
	}
	return res
}

func smoothingGradient(m *indexMesh, squares bool) []Coord {
	res := make([]Coord, len(m.Coords))
	for _, seg := range m.Segments {
		p1, p2 := m.Coords[seg[0]], m.Coords[seg[1]]
		p1ToP2 := p2.Sub(p1)
		if !squares {
			p1ToP2 = p1ToP2.Normalize()
		}
		res[seg[0]] = res[seg[0]].Add(p1ToP2)
		res[seg[1]] = res[seg[1]].Sub(p1ToP2)
	}
	return res
}

func optimalSmoothingStepSize(m *indexMesh, grad []Coord, squares bool) float64 {
	if squares {
		return optimalSmoothingStepSizeSquares(m, grad)
	}
	tmpCoords := make([]Coord, len(m.Coords))
	evalLength := func(stepSize float64) float64 {
		for j, c := range m.Coords {
			tmpCoords[j] = c.Add(grad[j].Scale(stepSize))
		}
		var result float64
		for _, s := range m.Segments {
			d := tmpCoords[s[0]].Dist(tmpCoords[s[1]])
			result += d
		}
		return result
	}
	minStep := 0.0
	maxStep := 1.0
	minLength := evalLength(minStep)
	maxLength := evalLength(maxStep)
	for maxLength < minLength {
		maxStep *= 2
		maxLength = evalLength(maxStep)
	}
	return numerical.GSS(minStep, maxStep, 32, evalLength)
}

func optimalSmoothingStepSizeSquares(m *indexMesh, grad []Coord) float64 {
	// Squared length of a segment is:
	//
	//     ||(p1+a*g1)-(p2+a*g2)||^2
	//     = ||(p1-p2) + a*(g1-g2)||^2
	//     = ||p1-p2||^2 + 2*a*(p1-p2)*(g1-g2) + a^2*||g1-g2||^2
	//     = poly(a=||g1-g2||^2, b=2*(p1-p2)*(g1-g2), c=||p1-p2||^2)
	//     Minimum = -b/2a
	//

	var polyB float64
	var polyA float64

	for _, s := range m.Segments {
		p1 := m.Coords[s[0]]
		p2 := m.Coords[s[1]]
		g1 := grad[s[0]]
		g2 := grad[s[1]]
		gDiff := g1.Sub(g2)
		pDiff := p1.Sub(p2)
		polyA += gDiff.Dot(gDiff)
		polyB += 2 * pDiff.Dot(gDiff)
	}

	return -polyB / (2 * polyA)
}
