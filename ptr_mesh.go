package model3d

// A ptrMesh is like a Mesh, but it's held together with
// pointers rather than hash maps, allowing for faster
// operations.
type ptrMesh struct {
	First *ptrTriangle
}

// newPtrMesh creates an empty ptrMesh.
func newPtrMesh() *ptrMesh {
	return &ptrMesh{}
}

// newPtrMeshMesh creates a ptrMesh from a Mesh.
func newPtrMeshMesh(m *Mesh) *ptrMesh {
	mapping := newPtrCoordMap()
	res := newPtrMesh()
	m.Iterate(func(t *Triangle) {
		res.Add(mapping.Triangle(t))
	})
	return res
}

// Add adds a triangle to the mesh.
//
// The triangle must not already be in a mesh.
func (p *ptrMesh) Add(t *ptrTriangle) {
	if t.Prev != nil || t.Next != nil {
		panic("triangle is already in a mesh")
	}
	if p.First != nil {
		p.First.Prev = t
	}
	t.Next = p.First
	p.First = t
}

// Remove a triangle t from the mesh.
func (p *ptrMesh) Remove(t *ptrTriangle) {
	prev, next := t.Prev, t.Next
	if prev == nil {
		p.First = next
	} else {
		prev.Next = next
	}
	if next != nil {
		next.Prev = prev
	}
	t.Next, t.Prev = nil, nil
}

// Iterate loops through the triangles in the mesh.
// During iteration, the mesh is immutable.
func (p *ptrMesh) Iterate(f func(t *ptrTriangle)) {
	t := p.First
	for t != nil {
		f(t)
		t = t.Next
	}
}

// Mesh turns the ptrMesh into a Mesh.
func (p *ptrMesh) Mesh() *Mesh {
	m := NewMesh()
	p.Iterate(func(t *ptrTriangle) {
		m.Add(&Triangle{t.Coords[0].Coord3D, t.Coords[1].Coord3D, t.Coords[2].Coord3D})
	})
	return m
}

// A ptrCoordMap stores pointers for Coord3D points.
// It can be used to convert points from a regular mesh
// into pointers for a ptrMesh.
type ptrCoordMap map[Coord3D]*ptrCoord

// newPtrCoordMap creates an empty coordinate map.
func newPtrCoordMap() ptrCoordMap {
	return ptrCoordMap{}
}

// Coord gets or creates a new pointer coordinate.
func (p ptrCoordMap) Coord(c Coord3D) *ptrCoord {
	if ptrC, ok := p[c]; ok {
		return ptrC
	} else {
		ptrC = &ptrCoord{Coord3D: c, Triangles: make([]*ptrTriangle, 0, 1)}
		p[c] = ptrC
		return ptrC
	}
}

// Triangle converts a triangle into a ptrTriangle using
// the pointers in the map.
func (p ptrCoordMap) Triangle(rawTriangle *Triangle) *ptrTriangle {
	t := &ptrTriangle{}
	for i, c := range rawTriangle {
		ptrC := p.Coord(c)
		ptrC.Triangles = append(ptrC.Triangles, t)
		t.Coords[i] = ptrC
	}
	return t
}

type ptrCoord struct {
	Coord3D
	Triangles []*ptrTriangle
}

// RemoveTriangle removes t from p.Triangles.
// It must be the case that t is in p.Triangles.
func (p *ptrCoord) RemoveTriangle(t *ptrTriangle) {
	for i, t1 := range p.Triangles {
		if t1 == t {
			// Unordered set removal with leak avoidance.
			last := len(p.Triangles) - 1
			p.Triangles[i] = p.Triangles[last]
			p.Triangles[last] = nil
			p.Triangles = p.Triangles[:last]
			return
		}
	}
	panic("coordinate not in triangle")
}

// Clusters returns all of the clusters of triangles
// connected to this vertex.
// A cluster is defined as a group of triangles which all
// share p and are all connected to each other by edges.
//
// A non-singular vertex has exactly one cluster.
func (p *ptrCoord) Clusters() [][]*ptrTriangle {
	unvisited := make(map[*ptrTriangle]bool, len(p.Triangles))
	for _, t := range p.Triangles {
		unvisited[t] = true
	}

	var families [][]*ptrTriangle
	for len(unvisited) > 0 {
		var first *ptrTriangle
		for t := range unvisited {
			first = t
			break
		}
		family := make([]*ptrTriangle, 1, len(unvisited))
		family[0] = first
		delete(unvisited, first)
		for queueIdx := 0; queueIdx < len(family); queueIdx++ {
			next := family[queueIdx]
			for _, c := range next.Coords {
				if c == p {
					continue
				}
				for _, t1 := range c.Triangles {
					if unvisited[t1] {
						delete(unvisited, t1)
						family = append(family, t1)
					}
				}
			}
		}
		families = append(families, family)
	}

	return families
}

// SortTriangles re-orders p.Triangles so that all of the
// triangles are connected to the next triangle in the
// list by an edge.
//
// If there are multiple clusters, it sorts each cluster
// separately.
func (p *ptrCoord) SortTriangles() {
	if len(p.Triangles) < 2 {
		return
	}

	var firstCorner, nextCorner *ptrCoord
	for _, c := range p.Triangles[0].Coords {
		if c != p {
			firstCorner = c
			nextCorner = c
			break
		}
	}

	for i := 1; i < len(p.Triangles)-1; i++ {
		var found bool
		for j := i; j < len(p.Triangles); j++ {
			t := p.Triangles[j]
			if !t.Contains(nextCorner) {
				continue
			}
			p.Triangles[i], p.Triangles[j] = p.Triangles[j], p.Triangles[i]
			nextCorner = newPtrSegment(p, nextCorner).Other(t)
			if nextCorner == firstCorner {
				return
			}
			found = true
			break
		}
		if !found {
			// Start sorting the next cluster.
			for _, c := range p.Triangles[i].Coords {
				if c != p {
					firstCorner = c
					nextCorner = c
					break
				}
			}
		}
	}
}

// A ptrTriangle is a triangle in a ptrMesh.
//
// The triangle's coordinates contain a pointer to it.
// When the triangle is done being used, call RemoveCoords
// on it to remove it from its coordinates.
// This is different from removing the triangle from its
// mesh.
type ptrTriangle struct {
	Coords [3]*ptrCoord
	Prev   *ptrTriangle
	Next   *ptrTriangle
}

// newPtrTriangle creates a triangle and adds it to all of
// its coordinates.
func newPtrTriangle(c1, c2, c3 *ptrCoord) *ptrTriangle {
	res := &ptrTriangle{Coords: [3]*ptrCoord{c1, c2, c3}}
	for _, c := range res.Coords {
		c.Triangles = append(c.Triangles, res)
	}
	return res
}

// Triangle creates a geometric triangle for p.
func (p *ptrTriangle) Triangle() *Triangle {
	return &Triangle{p.Coords[0].Coord3D, p.Coords[1].Coord3D, p.Coords[2].Coord3D}
}

// Contains checks if p contains a point c.
func (p *ptrTriangle) Contains(c *ptrCoord) bool {
	return p.Coords[0] == c || p.Coords[1] == c || p.Coords[2] == c
}

// RemoveCoords removes p from its coordinates.
func (p *ptrTriangle) RemoveCoords() {
	for _, c := range p.Coords {
		c.RemoveTriangle(p)
	}
}

// AddCoords un-does p.RemoveCoords().
func (p *ptrTriangle) AddCoords() {
	for _, c := range p.Coords {
		c.Triangles = append(c.Triangles, p)
	}
}

// A ptrSegment is a line segment.
type ptrSegment [2]*ptrCoord

// newPtrSegment creates a ptrSegment that is unique for
// the un-ordered pair c1, c2.
func newPtrSegment(c1, c2 *ptrCoord) ptrSegment {
	p1, p2 := c1.Coord3D, c2.Coord3D
	if p1.X < p2.X || (p1.X == p2.X && p1.Y < p2.Y) ||
		(p1.X == p2.X && p1.Y == p2.Y && p1.Z < p2.Z) {
		return ptrSegment{c1, c2}
	} else {
		return ptrSegment{c2, c1}
	}
}

// Triangles gets the triangles touching the segment.
func (p ptrSegment) Triangles() []*ptrTriangle {
	res := make([]*ptrTriangle, 0, 2)
	for _, t := range p[0].Triangles {
		if t.Contains(p[1]) {
			res = append(res, t)
		}
	}
	return res
}

// Mid gets the segment's midpoint.
func (p ptrSegment) Mid() Coord3D {
	return p[0].Coord3D.Mid(p[1].Coord3D)
}

// Other gets the point in t which is not in p.
func (p ptrSegment) Other(t *ptrTriangle) *ptrCoord {
	if t.Coords[0] != p[0] && t.Coords[0] != p[1] {
		return t.Coords[0]
	} else if t.Coords[1] != p[0] && t.Coords[1] != p[1] {
		return t.Coords[1]
	} else {
		return t.Coords[2]
	}
}
