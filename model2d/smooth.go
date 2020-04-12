package model2d

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
	tmpCoords := make([]Coord, len(m.Coords))
	evalLength := func(stepSize float64) float64 {
		for j, c := range m.Coords {
			tmpCoords[j] = c.Add(grad[j].Scale(stepSize))
		}
		var result float64
		for _, s := range m.Segments {
			d := tmpCoords[s[0]].Dist(tmpCoords[s[1]])
			if squares {
				result += d * d
			} else {
				result += d
			}
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

	for i := 0; i < 32; i++ {
		mid1 := minStep*2.0/3.0 + maxStep*1.0/3.0
		mid2 := minStep*1.0/3.0 + maxStep*2.0/3.0
		l1 := evalLength(mid1)
		l2 := evalLength(mid2)
		if l2 > l1 {
			maxStep = mid2
		} else {
			minStep = mid1
		}
	}

	return (minStep + maxStep) / 2
}
