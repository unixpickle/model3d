package model2d

type ptrMesh struct {
	first *ptrCoord
}

func newPtrMesh(m *Mesh) *ptrMesh {
	return newPtrMeshSegments(m.SegmentsSlice())
}

func newPtrMeshSegments(segs []*Segment) *ptrMesh {
	res := &ptrMesh{}
	coordMap := map[Coord]*ptrCoord{}
	for _, s := range segs {
		coords := [2]*ptrCoord{}
		for i, c := range s {
			ptrC, ok := coordMap[c]
			if !ok {
				ptrC = &ptrCoord{
					Coord:      c,
					prevCoords: make([]*ptrCoord, 0, 1),
					nextCoords: make([]*ptrCoord, 0, 1),

					listPrev: nil,
					listNext: res.first,
				}
				if res.first != nil {
					res.first.listPrev = ptrC
				}
				res.first = ptrC
				coordMap[c] = ptrC
			}
			coords[i] = ptrC
		}
		coords[0].nextCoords = append(coords[0].nextCoords, coords[1])
		coords[1].prevCoords = append(coords[1].prevCoords, coords[0])
	}
	return res
}

// Mesh converts this into a traditional Mesh object.
func (p *ptrMesh) Mesh() *Mesh {
	m := NewMesh()
	p.IterateCoords(func(c *ptrCoord) {
		for _, c1 := range p.Outgoing(c) {
			m.Add(&Segment{c.Coord, c1.Coord})
		}
	})
	return m
}

// Copy creates a deep copy of the mesh.
func (p *ptrMesh) Copy() *ptrMesh {
	res := &ptrMesh{}
	mapping := map[*ptrCoord]*ptrCoord{}
	p.IterateCoords(func(c *ptrCoord) {
		c1 := &ptrCoord{
			Coord:    c.Coord,
			listNext: p.first,
		}
		mapping[c] = c1
		if res.first != nil {
			res.first.listPrev = c1
		}
		res.first = c1
	})
	p.IterateCoords(func(c *ptrCoord) {
		c1 := mapping[c]
		for _, other := range c.nextCoords {
			c1.nextCoords = append(c1.nextCoords, mapping[other])
		}
		for _, other := range c.prevCoords {
			c1.prevCoords = append(c1.prevCoords, mapping[other])
		}
	})
	return res
}

// IterateCoords iterates over the points in the mesh.
//
// The mesh should not be modified during iteration.
func (p *ptrMesh) IterateCoords(f func(*ptrCoord)) {
	obj := p.first
	for obj != nil {
		f(obj)
		obj = obj.listNext
	}
}

// Peek gets the first coordinate in the mesh.
func (p *ptrMesh) Peek() *ptrCoord {
	return p.first
}

// Outgoing gets all of the coordinates to which edges
// from c connect.
func (p *ptrMesh) Outgoing(c *ptrCoord) []*ptrCoord {
	return c.nextCoords
}

// Incoming gets all of the coordinates connecting to c.
func (p *ptrMesh) Incoming(c *ptrCoord) []*ptrCoord {
	return c.prevCoords
}

// Remove removes an edge segment.
func (p *ptrMesh) Remove(p1, p2 *ptrCoord) {
	found := 0
	for i, c := range p1.nextCoords {
		if c == p2 {
			p1.nextCoords[i] = p1.nextCoords[len(p1.nextCoords)-1]
			p1.nextCoords = p1.nextCoords[:len(p1.nextCoords)-1]
			found++
			break
		}
	}
	for i, c := range p2.prevCoords {
		if c == p1 {
			p2.prevCoords[i] = p2.prevCoords[len(p2.prevCoords)-1]
			p2.prevCoords = p2.prevCoords[:len(p2.prevCoords)-1]
			found++
			break
		}
	}
	if found != 2 {
		panic("edge not found or state was inconsistent")
	}
	for _, c := range []*ptrCoord{p1, p2} {
		if c.NumEdges() == 0 {
			p.RemoveFromList(c)
		}
	}
}

// RemoveFromList removes the point from the ptrMesh, but
// does not remove its connections to neighbors.
//
// This is generally unsafe, and should be used with great
// caution.
func (p *ptrMesh) RemoveFromList(c *ptrCoord) {
	if c.listPrev == nil {
		p.first = c.listNext
	} else {
		c.listPrev.listNext = c.listNext
	}
	if c.listNext != nil {
		c.listNext.listPrev = c.listPrev
	}
	c.listPrev = nil
	c.listNext = nil
}

// Add adds an edge segment.
func (p *ptrMesh) Add(p1, p2 *ptrCoord) {
	for _, c := range p1.nextCoords {
		if c == p2 {
			panic("edge already exists")
		}
	}
	for _, c := range p2.prevCoords {
		if c == p1 {
			panic("inconsistent state (edge partially exists)")
		}
	}
	p1.nextCoords = append(p1.nextCoords, p2)
	p2.prevCoords = append(p2.prevCoords, p1)
	for _, c := range []*ptrCoord{p1, p2} {
		if c.NumEdges() == 1 {
			if c.listPrev != nil || c.listNext != nil {
				panic("bad linked list state.")
			}
			if p.first == nil {
				p.first = c
			} else {
				c.listNext = p.first
				p.first.listPrev = c
				p.first = c
			}
		}
	}
}

type ptrCoord struct {
	Coord

	prevCoords []*ptrCoord
	nextCoords []*ptrCoord

	listPrev *ptrCoord
	listNext *ptrCoord
}

func (p *ptrCoord) NumEdges() int {
	return len(p.prevCoords) + len(p.nextCoords)
}
