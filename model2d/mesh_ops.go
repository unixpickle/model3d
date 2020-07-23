package model2d

import "math"

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

// RepairNormals flips normals when they point within the
// shape defined by the mesh, as determined by the
// even-odd rule.
//
// The repaired mesh is returned, along with the number of
// modified segments.
//
// The check is performed by adding the normal, scaled by
// epsilon, to the center of the segment, and then
// counting the number of ray collisions from this point
// in the direction of the normal.
func (m *Mesh) RepairNormals(epsilon float64) (*Mesh, int) {
	collider := MeshToCollider(m)
	solid := NewColliderSolid(collider)
	numFlipped := 0
	newMesh := NewMesh()

	m.Iterate(func(s *Segment) {
		s1 := *s
		normal := s.Normal()
		center := s.Mid()
		movedOut := center.Add(normal.Scale(epsilon))
		if solid.Contains(movedOut) {
			numFlipped++
			s1[0], s1[1] = s1[1], s1[0]
		}
		newMesh.Add(&s1)
	})
	return newMesh, numFlipped
}

// Decimate repeatedly removes vertices from a manifold
// mesh until the mesh contains maxVertices or fewer.
func (m *Mesh) Decimate(maxVertices int) *Mesh {
	res := NewMesh()
	res.AddMesh(m)
	areas := map[Coord]float64{}
	for _, v := range res.VertexSlice() {
		areas[v] = vertexArea(m, v)
	}

	for len(areas) > maxVertices {
		var next Coord
		nextArea := math.Inf(1)
		for v, a := range areas {
			if a < nextArea {
				nextArea = a
				next = v
			}
		}
		delete(areas, next)
		n1, n2 := vertexNeighbors(res, next)
		for _, s := range res.Find(next) {
			res.Remove(s)
		}
		res.Add(&Segment{n1, n2})
		areas[n1] = vertexArea(res, n1)
		areas[n2] = vertexArea(res, n2)
	}
	return res
}

func vertexArea(m *Mesh, c Coord) float64 {
	n1, n2 := vertexNeighbors(m, c)
	mat := NewMatrix2Columns(n2.Sub(c), n1.Sub(c))
	return math.Abs(mat.Det() / 2)
}

func vertexNeighbors(m *Mesh, c Coord) (n1, n2 Coord) {
	neighbors := m.Find(c)
	if len(neighbors) != 2 {
		panic("non-manifold mesh")
	}
	n1 = neighbors[0][0]
	if n1 == c {
		n1 = neighbors[0][1]
	}
	n2 = neighbors[1][1]
	if n2 == c {
		n2 = neighbors[1][0]
	}
	if c != neighbors[0][1] {
		// Attempt to preserve normals.
		n1, n2 = n2, n1
	}
	return
}
