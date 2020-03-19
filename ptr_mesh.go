package model3d

// A ptrMesh is like a Mesh, but it's held together with
// pointers rather than hash maps, allowing for faster
// operations.
type ptrMesh struct {
	CoordMap map[Coord3D]*ptrCoord
	First    *ptrTriangle
}

// newPtrMesh creates an empty ptrMesh.
func newPtrMesh() *ptrMesh {
	return &ptrMesh{CoordMap: map[Coord3D]*ptrCoord{}}
}

// Add adds a triangle to the mesh.
//
// The triangle must not already be in a mesh.
func (p *ptrMesh) Add(t *ptrTriangle) {
	if t.Prev != nil || t.Next != nil {
		panic("triangle is already in a mesh")
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
	for _, c := range t.Coords {
		c.RemoveTriangle(t)
	}
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

type ptrTriangle struct {
	Coords [3]*ptrCoord
	Prev   *ptrTriangle
	Next   *ptrTriangle
}
