package model3d

import (
	"math"
)

// Blur creates a new mesh by moving every vertex closer
// to its connected vertices.
//
// The rate argument specifies how much the vertex should
// be moved, 0 being no movement and 1 being the most.
// If multiple rates are passed, then multiple iterations
// of the algorithm are performed in succession.
// If a rate of -1 is passed, then all of neighbors are
// averaged together with each point, and the resulting
// average is used.
func (m *Mesh) Blur(rates ...float64) *Mesh {
	return m.BlurFiltered(nil, rates...)
}

// BlurFiltered is like Blur, but vertices are only
// considered neighbors if f returns true for their
// initial coordinates.
//
// Once vertices are considered neighbors, they will be
// treated as such for every blur iteration, even if the
// coordinates change in such a way that f would no longer
// consider them neighbors.
//
// If f is nil, then this is equivalent to Blur().
func (m *Mesh) BlurFiltered(f func(c1, c2 Coord3D) bool, rates ...float64) *Mesh {
	capacity := len(m.triangles) * 3
	if m.vertexToTriangle != nil {
		capacity = len(m.vertexToTriangle)
	}
	coordToIdx := make(map[Coord3D]int, capacity)
	coords := make([]Coord3D, 0, capacity)
	neighbors := make([][]int, 0, capacity)
	m.Iterate(func(t *Triangle) {
		var indices [3]int
		for i, c := range t {
			if idx, ok := coordToIdx[c]; !ok {
				indices[i] = len(coords)
				coordToIdx[c] = len(coords)
				coords = append(coords, c)
				neighbors = append(neighbors, []int{})
			} else {
				indices[i] = idx
			}
		}
		for _, idx1 := range indices {
			for _, idx2 := range indices {
				if idx1 == idx2 {
					continue
				}
				var found bool
				for _, n := range neighbors[idx1] {
					if n == idx2 {
						found = true
						break
					}
				}
				if !found && (f == nil || f(coords[idx1], coords[idx2])) {
					neighbors[idx1] = append(neighbors[idx1], idx2)
				}
			}
		}
	})

	newCoords := make([]Coord3D, len(coords))
	for _, rate := range rates {
		for i, c := range coords {
			ns := neighbors[i]
			if len(ns) == 0 {
				newCoords[i] = c
				continue
			}

			neighborAvg := Coord3D{}
			for _, c1 := range ns {
				neighborAvg = neighborAvg.Add(coords[c1])
			}

			var newPoint Coord3D
			if rate == -1 {
				newPoint = neighborAvg.Add(c).Scale(1 / float64(len(ns)+1))
			} else {
				neighborAvg = neighborAvg.Scale(1 / float64(len(ns)))
				newPoint = neighborAvg.Scale(rate).Add(c.Scale(1 - rate))
			}

			newCoords[i] = newPoint
		}
		copy(coords, newCoords)
	}

	m1 := NewMesh()
	m.Iterate(func(t *Triangle) {
		t1 := *t
		for i, c := range t1 {
			t1[i] = coords[coordToIdx[c]]
		}
		m1.Add(&t1)
	})
	return m1
}

// SmoothAreas uses gradient descent to iteratively smooth
// out the surface by moving every vertex in the direction
// that minimizes the area of its adjacent triangles.
//
// The stepSize argument specifies how much the vertices
// are moved at each iteration. Good values depend on the
// mesh, but a good start is on the order of 0.1.
//
// The iters argument specifies how many gradient steps
// are taken.
//
// This algorithm can produce very smooth objects, but it
// is much less efficient than Blur().
// Consider Blur() when perfect smoothness is not
// required.
func (m *Mesh) SmoothAreas(stepSize float64, iters int) *Mesh {
	capacity := len(m.triangles) * 3
	if m.vertexToTriangle != nil {
		capacity = len(m.vertexToTriangle)
	}

	coordToIdx := make(map[Coord3D]int, capacity)
	coords := make([]Coord3D, 0, capacity)
	triangles := make([][3]int, 0, len(m.triangles))

	m.Iterate(func(t *Triangle) {
		var triangle [3]int
		for i, c := range t {
			idx, ok := coordToIdx[c]
			if !ok {
				idx = len(coords)
				coordToIdx[c] = idx
				coords = append(coords, c)
			}
			triangle[i] = idx
		}
		triangles = append(triangles, triangle)
	})

	newCoords := append([]Coord3D{}, coords...)
	for i := 0; i < iters; i++ {
		for _, triangle := range triangles {
			var t Triangle
			for i, j := range triangle {
				t[i] = coords[j]
			}
			for i, grad := range t.AreaGradient() {
				j := triangle[i]
				newCoords[j] = newCoords[j].Add(grad.Scale(-stepSize))
			}
		}
		copy(coords, newCoords)
	}

	m1 := NewMesh()
	for _, triangle := range triangles {
		var t Triangle
		for i, j := range triangle {
			t[i] = coords[j]
		}
		m1.Add(&t)
	}
	return m1
}

// Repair finds vertices that are close together and
// combines them into one.
//
// The epsilon argument controls how close points have to
// be. In particular, it sets the approximate maximum
// distance across all dimensions.
func (m *Mesh) Repair(epsilon float64) *Mesh {
	hashToClass := map[Coord3D]*equivalenceClass{}
	allClasses := map[*equivalenceClass]bool{}
	for c := range m.getVertexToTriangle() {
		hashes := make([]Coord3D, 0, 8)
		classes := make(map[*equivalenceClass]bool, 8)
		for i := 0.0; i <= 1.0; i += 1.0 {
			for j := 0.0; j <= 1.0; j += 1.0 {
				for k := 0.0; k <= 1.0; k += 1.0 {
					hash := Coord3D{
						X: math.Round(c.X/epsilon) + i,
						Y: math.Round(c.Y/epsilon) + j,
						Z: math.Round(c.Z/epsilon) + k,
					}
					hashes = append(hashes, hash)
					if class, ok := hashToClass[hash]; ok {
						classes[class] = true
					}
				}
			}
		}
		if len(classes) == 0 {
			class := &equivalenceClass{
				Elements:  []Coord3D{c},
				Hashes:    hashes,
				Canonical: c,
			}
			for _, hash := range hashes {
				hashToClass[hash] = class
			}
			allClasses[class] = true
			continue
		}
		newClass := &equivalenceClass{
			Elements:  []Coord3D{c},
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
	}

	coordToClass := map[Coord3D]*equivalenceClass{}
	for class := range allClasses {
		for _, c := range class.Elements {
			coordToClass[c] = class
		}
	}

	return m.MapCoords(func(c Coord3D) Coord3D {
		return coordToClass[c].Canonical
	})
}

// An equivalenceClass stores a set of points which share
// hashes. It is used for Repair to group vertices.
type equivalenceClass struct {
	Elements  []Coord3D
	Hashes    []Coord3D
	Canonical Coord3D
}

// NeedsRepair checks if every edge touches exactly two
// triangles. If not, NeedsRepair returns true.
func (m *Mesh) NeedsRepair() bool {
	for t := range m.triangles {
		for i := 0; i < 3; i++ {
			p1 := t[i]
			p2 := t[(i+1)%3]
			if len(m.Find(p1, p2)) != 2 {
				return true
			}
		}
	}
	return false
}

// SingularVertices gets the points at which the mesh is
// squeezed to zero volume. In other words, it gets the
// points where two pieces of volume are barely touching
// by a single point.
func (m *Mesh) SingularVertices() []Coord3D {
	var res []Coord3D
	for vertex, tris := range m.getVertexToTriangle() {
		if len(tris) == 0 {
			continue
		}

		// Queue used for a breadth-first search.
		// One entry per triangle. A 0 means unvisited.
		// A 1 means visited but not expanded. A 2 means
		// visited and expanded.
		queue := make([]int, len(tris))

		// Start at first triangle and check that all
		// others are connected.
		queue[0] = 1
		changed := true
		numVisited := 1
		for changed {
			changed = false
			for i, status := range queue {
				if status != 1 {
					continue
				}
				t := tris[i]
				for j, t1 := range tris {
					if queue[j] == 0 && t.SharesEdge(t1) {
						queue[j] = 1
						numVisited++
						changed = true
					}
				}
				queue[i] = 2
			}
		}
		if numVisited != len(tris) {
			res = append(res, vertex)
		}
	}
	return res
}

// SelfIntersections counts the number of times the mesh
// intersects itself.
// In an ideal mesh, this would be 0.
func (m *Mesh) SelfIntersections() int {
	var res int
	collider := MeshToCollider(m)
	m.Iterate(func(t *Triangle) {
		res += len(collider.TriangleCollisions(t))
	})
	return res
}

// RepairNormals flips normals when they point within the
// solid defined by the mesh, as determined by the
// even-odd rule.
//
// The repaired mesh is returned, along with the number of
// modified triangles.
//
// The check is performed by adding the normal, scaled by
// epsilon, to the center of the triangle, and then
// counting the number of ray collisions from this point
// in the direction of the normal.
func (m *Mesh) RepairNormals(epsilon float64) (*Mesh, int) {
	collider := MeshToCollider(m)
	numFlipped := 0
	newMesh := NewMesh()
	m.Iterate(func(t *Triangle) {
		t1 := *t
		normal := t.Normal()
		center := t[0].Add(t[1]).Add(t[2]).Scale(1.0 / 3)
		origin := center.Add(normal.Scale(epsilon))
		if collider.RayCollisions(&Ray{Origin: origin, Direction: normal})%2 == 1 {
			numFlipped++
			t1[0], t1[1] = t1[1], t1[0]
		}
		newMesh.Add(&t1)
	})
	return newMesh, numFlipped
}

// EliminateEdges creates a new mesh by iteratively
// removing edges according to the function f.
//
// The f function takes the current new mesh and a line
// segment, and returns true if the segment should be
// removed.
func (m *Mesh) EliminateEdges(f func(tmp *Mesh, segment Segment) bool) *Mesh {
	result := NewMesh()
	m.Iterate(func(t *Triangle) {
		t1 := *t
		result.Add(&t1)
	})
	changed := true
	for changed {
		changed = false
		remainingSegments := map[Segment]bool{}
		result.Iterate(func(t *Triangle) {
			for _, seg := range t.Segments() {
				remainingSegments[seg] = true
			}
		})
		for len(remainingSegments) > 0 {
			var segment Segment
			for seg := range remainingSegments {
				segment = seg
				delete(remainingSegments, seg)
				break
			}
			if !canEliminateSegment(result, segment) || !f(result, segment) {
				continue
			}
			eliminateSegment(result, segment, remainingSegments)
			changed = true
		}
	}
	return result
}

// EliminateCoplanar eliminates line segments which are
// touching a collection of coplanar triangles.
//
// The epsilon argument controls how close two normals
// must be for the triangles to be considered coplanar.
func (m *Mesh) EliminateCoplanar(epsilon float64) *Mesh {
	return m.EliminateEdges(func(m *Mesh, s Segment) bool {
		isFirst := true
		var normal Coord3D
		for _, p := range s {
			for _, neighbor := range m.getVertexToTriangle()[p] {
				if isFirst {
					normal = neighbor.Normal()
					isFirst = false
				} else if math.Abs(neighbor.Normal().Dot(normal)-1) > epsilon {
					return false
				}
			}
		}
		return true
	})
}

func canEliminateSegment(m *Mesh, seg Segment) bool {
	// Segment removal must not leave either point
	// from the segment in the mesh.
	if seg[0] == seg[1] {
		return false
	}

	v2t := m.getVertexToTriangle()
	neighbors1 := v2t[seg[0]]
	neighbors2 := v2t[seg[1]]
	otherSegs := make([]Segment, 0, len(neighbors1)+len(neighbors2))
	for i, neighbors := range [][]*Triangle{neighbors1, neighbors2} {
		for _, t := range neighbors {
			p1, p2 := t[0], t[1]
			if p1 == seg[0] || p1 == seg[1] {
				p1, p2 = p2, t[2]
			} else if p2 == seg[0] || p2 == seg[1] {
				p2 = t[2]
			}
			if p2 == seg[0] || p2 == seg[1] {
				// This triangle contains the segment and
				// will definitely be eliminated.
				continue
			}
			otherSeg := NewSegment(p1, p2)
			if i == 1 {
				for _, s := range otherSegs {
					if s == otherSeg {
						// Two triangles will become duplicates.
						return false
					}
				}
			}
			otherSegs = append(otherSegs, otherSeg)

			t1 := Triangle{p1, p2, seg[0]}
			t2 := Triangle{p1, p2, seg[1]}
			if t1.Normal().Dot(t2.Normal()) < 0 {
				return false
			}
		}
	}
	return true
}

func eliminateSegment(m *Mesh, segment Segment, remaining map[Segment]bool) {
	mp := segment.Mid()
	v2t := m.getVertexToTriangle()
	newNeighbors := []*Triangle{}
	for i, segmentPoint := range segment {
		for _, neighbor := range v2t[segmentPoint] {
			var removedSegs int
			for _, seg := range neighbor.Segments() {
				if seg[0] == segment[0] || seg[0] == segment[1] || seg[1] == segment[0] ||
					seg[1] == segment[1] {
					delete(remaining, seg)
					removedSegs++
				}
			}

			if removedSegs == 3 {
				if i == 0 {
					// This triangle contains the segment,
					// so it must be fully removed.
					delete(m.triangles, neighbor)
					for _, p := range neighbor {
						if p != segment[0] && p != segment[1] {
							m.removeTriangleFromVertex(neighbor, p)
							break
						}
					}
				}
				continue
			}

			for i, p := range neighbor {
				if p == segment[0] || p == segment[1] {
					neighbor[i] = mp
					seg1 := NewSegment(mp, neighbor[(i+1)%3])
					seg2 := NewSegment(mp, neighbor[(i+2)%3])
					remaining[seg1] = true
					remaining[seg2] = true
					break
				}
			}

			newNeighbors = append(newNeighbors, neighbor)
			delete(v2t, segmentPoint)
		}
	}
	v2t[mp] = newNeighbors
}
