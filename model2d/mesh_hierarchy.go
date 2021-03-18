// Generated from templates/mesh_hierarchy.template

package model2d

import "sort"

var arbitraryAxis Coord = Coord{X: 0.95177695, Y: 0.26858931}

// A MeshHierarchy is a tree structure where each node is
// a closed, simple polygon, and children are contained
// inside their parents.
//
// Only manifold meshes with no self-intersections can be
// converted into a MeshHierarchy.
type MeshHierarchy struct {
	// Mesh is the root shape of this (sub-)hierarchy.
	Mesh *Mesh

	// MeshSolid is a solid indicating which points are
	// contained in the mesh.
	MeshSolid Solid

	Children []*MeshHierarchy
}

// MeshToHierarchy creates a MeshHierarchy for each
// exterior mesh contained in m.
//
// The mesh m must be manifold and have no
// self-intersections.
func MeshToHierarchy(m *Mesh) []*MeshHierarchy {
	if !m.Manifold() {
		panic("mesh must be manifold")
	}
	return uncheckedMeshToHierarchy(m)
}

func uncheckedMeshToHierarchy(m *Mesh) []*MeshHierarchy {
	pm := newPtrMesh(m)
	sorted := newSortedCoords(pm)

	var result []*MeshHierarchy

ClosedMeshLoop:
	for {
		minVertex := sorted.Next()
		if minVertex == nil {
			break
		}
		stripped := removeAllConnected(pm, minVertex)
		GroupSegments(stripped)
		solid := NewColliderSolid(GroupedSegmentsToCollider(stripped))
		strippedMesh := NewMeshSegments(stripped)
		for _, x := range result {
			if x.MeshSolid.Contains(minVertex.Coord) {
				// We know the mesh is a leaf, because if it contained
				// any other mesh, that mesh would have to have a higher
				// minVertex along an arbitrary axis, and would not have
				// been added yet.
				x.insertLeaf(strippedMesh, solid, minVertex.Coord)
				continue ClosedMeshLoop
			}
		}
		// If we are here, this is a root mesh.
		result = append(result, &MeshHierarchy{
			Mesh:      strippedMesh,
			MeshSolid: solid,
		})
	}

	return result
}

// insertLeaf inserts a mesh into the hierarchy, knowing
// that the mesh is a leaf in the current hierarchy.
func (m *MeshHierarchy) insertLeaf(mesh *Mesh, solid Solid, c Coord) {
	v := mesh.VertexSlice()[0]
	for _, child := range m.Children {
		if child.MeshSolid.Contains(v) {
			child.insertLeaf(mesh, solid, c)
			return
		}
	}
	m.Children = append(m.Children, &MeshHierarchy{
		Mesh:      mesh,
		MeshSolid: solid,
	})
}

// FullMesh re-combines the root mesh with all of its
// children.
func (m *MeshHierarchy) FullMesh() *Mesh {
	res := NewMeshSegments(m.Mesh.SegmentSlice())
	for _, child := range m.Children {
		res.AddMesh(child.FullMesh())
	}
	return res
}

// MapCoords creates a new MeshHierarchy by applying f to
// every coordinate in every mesh.
func (m *MeshHierarchy) MapCoords(f func(Coord) Coord) *MeshHierarchy {
	res := &MeshHierarchy{
		Mesh: m.Mesh.MapCoords(f),
	}
	res.MeshSolid = NewColliderSolid(MeshToCollider(res.Mesh))
	for _, child := range m.Children {
		res.Children = append(res.Children, child.MapCoords(f))
	}
	return res
}

// Min gets the minimum point of the outer mesh's
// bounding box.
func (m *MeshHierarchy) Min() Coord {
	return m.MeshSolid.Min()
}

// Max gets the maximum point of the outer mesh's
// bounding box.
func (m *MeshHierarchy) Max() Coord {
	return m.MeshSolid.Max()
}

// Contains checks if c is inside the hierarchy using the
// even-odd rule.
func (m *MeshHierarchy) Contains(c Coord) bool {
	if !m.MeshSolid.Contains(c) {
		return false
	}
	for _, child := range m.Children {
		if child.Contains(c) {
			return false
		}
	}
	return true
}

// removeAllConnected strips all segments connected to c
// out of m and returns them as segments.
func removeAllConnected(m *ptrMesh, c *ptrCoord) []*Segment {
	var result []*Segment
	first := c
	for c != nil {
		if len(m.Outgoing(c)) != 1 || len(m.Incoming(c)) != 1 {
			panic("mesh is non-manifold")
		}
		next := m.Outgoing(c)[0]
		result = append(result, &Segment{c.Coord, next.Coord})
		m.RemoveFromList(c)
		if next == first {
			break
		}
		c = next
	}
	return result
}

// coordInMesh checks if a ptrCoord is in a ptrMesh.
// This assumes that a coordinate will be removed from the
// mesh if any of its corresponding faces are removed,
// which will always happen if the mesh is manifold.
func coordInMesh(m *ptrMesh, c *ptrCoord) bool {
	return c.listPrev != nil || m.first == c
}

type sortedCoords struct {
	dots   []float64
	coords []*ptrCoord

	mesh   *ptrMesh
	curIdx int
}

func newSortedCoords(m *ptrMesh) *sortedCoords {
	var coords []*ptrCoord
	var dots []float64
	m.Iterate(func(c *ptrCoord) {
		coords = append(coords, c)
		dots = append(dots, c.Dot(arbitraryAxis))
	})
	res := &sortedCoords{
		dots:   dots,
		coords: coords,
		mesh:   m,
	}
	sort.Sort(res)
	return res
}

func (s *sortedCoords) Next() *ptrCoord {
	for s.curIdx < s.Len() {
		c := s.coords[s.curIdx]
		s.curIdx++
		if coordInMesh(s.mesh, c) {
			return c
		}
	}
	return nil
}

func (s *sortedCoords) Len() int {
	return len(s.dots)
}

func (s *sortedCoords) Less(i, j int) bool {
	return s.dots[i] < s.dots[j]
}

func (s *sortedCoords) Swap(i, j int) {
	s.dots[i], s.dots[j] = s.dots[j], s.dots[i]
	s.coords[i], s.coords[j] = s.coords[j], s.coords[i]
}
