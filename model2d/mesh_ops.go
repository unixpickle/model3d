package model2d

// Blur moves each vertex closer to the average of its
// neighbors.
//
// The rate argument controls how much the vertices move.
// If it is 1, then the vertices become the average of
// their neighbors. If it is 0, then the vertices remain
// where they are.
func (m *Mesh) Blur(rate float64) *Mesh {
	return m.MapCoords(func(c Coord) Coord {
		count := 0.0
		sum := Coord{}
		for _, s := range m.Find(c) {
			for _, c1 := range s {
				if c1 != c {
					sum = sum.Add(c1)
					count++
				}
			}
		}
		return c.Scale(1 - rate).Add(sum.Scale(rate / count))
	})
}

// Smooth is similar to Blur, but it is less sensitive to
// differences in segment length.
func (m *Mesh) Smooth(iters int) *Mesh {
	im := newIndexMesh(m)
	for i := 0; i < iters; i++ {
		im.Smooth(false)
	}
	return im.Mesh()
}

// SmoothSq is like Smooth, but it minimizes the sum of
// squared segment lengths rather than the sum of lengths
// directly.
// Thus, SmoothSq produces more even segments than Smooth.
func (m *Mesh) SmoothSq(iters int) *Mesh {
	im := newIndexMesh(m)
	for i := 0; i < iters; i++ {
		im.Smooth(true)
	}
	return im.Mesh()
}

// Subdivide uses Chaikin subdivision to add segments
// between every vertex.
//
// This can only be applied to manifold meshes.
// This can be checked with m.Manifold().
func (m *Mesh) Subdivide(iters int) *Mesh {
	if !m.Manifold() {
		panic("mesh is non-manifold")
	}
	if iters < 1 {
		panic("must subdivide for at least one iteration")
	}
	current := m
	for i := 0; i < iters; i++ {
		next := NewMesh()
		current.Iterate(func(s *Segment) {
			mp1 := s[0].Scale(0.75).Add(s[1].Scale(0.25))
			mp2 := s[1].Scale(0.75).Add(s[0].Scale(0.25))
			next.Add(&Segment{mp1, mp2})

			// Do the subdivision for the second vertex.
			// If every segment's second vertex is taken care of,
			// every vertex will be covered.
			others := current.Find(s[1])
			s1 := others[0]
			if s1 == s {
				s1 = others[1]
			}
			mp3 := s1[0].Scale(0.75).Add(s1[1].Scale(0.25))
			next.Add(&Segment{mp2, mp3})
		})
		current = next
	}
	return current
}

// Manifold checks if the mesh is manifold, i.e. if every
// vertex has two segments.
func (m *Mesh) Manifold() bool {
	for _, s := range m.getVertexToSegment() {
		if len(s) != 2 {
			return false
		}
	}
	return true
}
