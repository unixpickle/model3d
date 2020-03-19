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

// Add translates a Triangle into a ptrTriangle and adds
// it to the mesh.
func (p *ptrMesh) Add(rawTriangle *Triangle) *ptrTriangle {
	t := &ptrTriangle{}
	for i, c := range rawTriangle {
		if ptrC, ok := p.CoordMap[c]; ok {
			ptrC.Triangles = append(ptrC.Triangles, t)
			t.Coords[i] = ptrC
		} else {
			ptrC = &ptrCoord{Coord3D: c, Triangles: []*ptrTriangle{t}}
			t.Coords[i] = ptrC
			p.CoordMap[c] = ptrC
		}
	}
	t.Next = p.First
	p.First = t
	return t
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
