package model2d

import "math"

// Invert flips every segment in the mesh, effectively
// inverting all the normals.
func (m *Mesh) Invert() *Mesh {
	res := NewMesh()
	m.Iterate(func(s *Segment) {
		res.Add(&Segment{s[1], s[0]})
	})
	return res
}

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
	result := true
	m.getVertexToFace().Range(func(_ Coord, s []*Segment) bool {
		if len(s) == 2 {
			return true
		} else {
			result = false
			return false
		}
	})
	return result
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

// Repair finds vertices that are close together and
// combines them into one.
//
// The epsilon argument controls how close points have to
// be. In particular, it sets the approximate maximum
// distance across all dimensions.
func (m *Mesh) Repair(epsilon float64) *Mesh {
	hashToClass := map[Coord]*equivalenceClass{}
	allClasses := map[*equivalenceClass]bool{}
	m.getVertexToFace().KeyRange(func(c Coord) bool {
		hashes := make([]Coord, 0, 4)
		classes := make(map[*equivalenceClass]bool, 8)
		for i := 0.0; i <= 1.0; i += 1.0 {
			for j := 0.0; j <= 1.0; j += 1.0 {
				hash := Coord{
					X: math.Round(c.X/epsilon) + i,
					Y: math.Round(c.Y/epsilon) + j,
				}
				hashes = append(hashes, hash)
				if class, ok := hashToClass[hash]; ok {
					classes[class] = true
				}
			}
		}
		if len(classes) == 0 {
			class := &equivalenceClass{
				Elements:  []Coord{c},
				Hashes:    hashes,
				Canonical: c,
			}
			for _, hash := range hashes {
				hashToClass[hash] = class
			}
			allClasses[class] = true
			return true
		}
		newClass := &equivalenceClass{
			Elements:  []Coord{c},
			Hashes:    hashes,
			Canonical: c,
		}
		for class := range classes {
			delete(allClasses, class)
			newClass.Elements = append(newClass.Elements, class.Elements...)
			for _, hash := range class.Hashes {
				var found bool
				for _, hash1 := range newClass.Hashes {
					if hash1 == hash {
						found = true
						break
					}
				}
				if !found {
					newClass.Hashes = append(newClass.Hashes, hash)
				}
			}
		}
		for _, hash := range newClass.Hashes {
			hashToClass[hash] = newClass
		}
		allClasses[newClass] = true
		return true
	})

	coordToClass := map[Coord]*equivalenceClass{}
	for class := range allClasses {
		for _, c := range class.Elements {
			coordToClass[c] = class
		}
	}

	return m.MapCoords(func(c Coord) Coord {
		return coordToClass[c].Canonical
	})
}

// An equivalenceClass stores a set of points which share
// hashes. It is used for Repair to group vertices.
type equivalenceClass struct {
	Elements  []Coord
	Hashes    []Coord
	Canonical Coord
}

// Decimate repeatedly removes vertices from a mesh until
// it contains maxVertices or fewer vertices with two
// neighbors.
//
// For manifold meshes, maxVertices is a hard-limit on the
// number of resulting vertices.
// For non-manifold meshes, more than maxVertices vertices
// will be retained if all of the remaining vertices are
// not part of exactly two segments.
func (m *Mesh) Decimate(maxVertices int) *Mesh {
	res := NewMesh()
	res.AddMesh(m)
	areas := map[Coord]float64{}
	for _, v := range res.VertexSlice() {
		areas[v] = vertexArea(m, v)
	}

	for len(areas) > maxVertices {
		var next Coord
		nextArea := -1.0
		for v, a := range areas {
			if a == -1 {
				continue
			}
			if a < nextArea || nextArea == -1 {
				nextArea = a
				next = v
			}
		}
		if nextArea == -1 {
			break
		}
		delete(areas, next)
		n1, n2, _ := vertexNeighbors(res, next)
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
	n1, n2, ok := vertexNeighbors(m, c)
	if !ok {
		return -1
	}
	mat := NewMatrix2Columns(n2.Sub(c), n1.Sub(c))
	return math.Abs(mat.Det() / 2)
}

func vertexNeighbors(m *Mesh, c Coord) (n1, n2 Coord, ok bool) {
	neighbors := m.Find(c)
	if len(neighbors) != 2 {
		ok = false
		return
	}
	ok = true
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

// EliminateColinear eliminates vertices that connect
// nearly co-linear edges.
//
// The epsilon argument should be a small positive value
// that is used to approximate co-linearity.
// A good value for very precise results is 1e-8.
func (m *Mesh) EliminateColinear(epsilon float64) *Mesh {
	res := NewMesh()
	res.AddMesh(m)

	eligible := map[Coord]bool{}
	for _, v := range res.VertexSlice() {
		if vertexNormalDifference(m, v) < epsilon {
			eligible[v] = true
		}
	}

	for len(eligible) > 0 {
		var next Coord
		for c := range eligible {
			next = c
			break
		}
		delete(eligible, next)
		n1, n2, _ := vertexNeighbors(m, next)
		for _, seg := range res.Find(next) {
			res.Remove(seg)
		}
		res.Add(&Segment{n1, n2})
		for _, c := range []Coord{n1, n2} {
			if vertexNormalDifference(m, c) < epsilon {
				eligible[c] = true
			} else if eligible[c] {
				delete(eligible, c)
			}
		}
	}

	return res
}

func vertexNormalDifference(m *Mesh, c Coord) float64 {
	segs := m.Find(c)
	if len(segs) != 2 {
		return math.Inf(1)
	}
	n1 := segs[0].Normal()
	n2 := segs[1].Normal()
	return 1 - math.Min(1, n1.Dot(n2))
}
