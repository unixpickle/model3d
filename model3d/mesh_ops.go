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
	capacity := len(m.faces) * 3
	if v2t := m.getVertexToFaceOrNil(); v2t != nil {
		capacity = v2t.Len()
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
//
// This method is a simpler version of the MeshSmoother
// object, which provides a more controlled smoothing API.
func (m *Mesh) SmoothAreas(stepSize float64, iters int) *Mesh {
	smoother := &MeshSmoother{
		StepSize:   stepSize,
		Iterations: iters,
	}
	return smoother.Smooth(m)
}

// VertexNormals approximates normals for every vertex on
// the mesh. The normals are returned in a mapping from
// vertex coordinates to normals (always of unit length).
//
// If the mesh has self-intersections, degenerate geometry,
// or is non-manifold, the resulting normals may include
// inf or NaN values due to divisions by zero.
//
// This employs the "Mean Weighted by Angle" approach as
// described in
// "A Comparison of Algorithms for Vertex Normal Computations"
// http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.99.2846&rep=rep1&type=pdf.
func (m *Mesh) VertexNormals() *CoordToCoord {
	sums := NewCoordToCoord()
	m.Iterate(func(t *Triangle) {
		edges := [3]Coord3D{
			t[0].Sub(t[1]).Normalize(),
			t[1].Sub(t[2]).Normalize(),
			t[2].Sub(t[0]).Normalize(),
		}
		normal := t.Normal()
		for i, c := range t {
			e1 := edges[(i+2)%3]
			e2 := edges[i]
			theta := math.Asin(math.Min(1.0, e1.Cross(e2).Norm()))
			cur, _ := sums.Load(c)
			sums.Store(c, cur.Add(normal.Scale(theta)))
		}
	})
	normalized := NewCoordToCoord()
	sums.Range(func(k, v Coord3D) bool {
		normalized.Store(k, v.Normalize())
		return true
	})
	return normalized
}

// FlattenBase flattens out the bases of objects for
// printing on an FDM 3D printer. It is intended to be
// used for meshes based on flat-based solids, where the
// base's edges got rounded by smoothing.
//
// The maxAngle argument specifies the maximum angle (in
// radians) to flatten. If left at zero, 45 degrees is
// used since this is the recommended angle on many FDM 3D
// printers.
//
// In some meshes, this may cause triangles to overlap on
// the base of the mesh. Thus, it is only intended to be
// used when the base is clearly defined, and all of the
// triangles touching it are not above any other
// triangles (along the Z-axis).
func (m *Mesh) FlattenBase(maxAngle float64) *Mesh {
	if maxAngle == 0 {
		maxAngle = math.Pi / 4
	}
	minZ := m.Min().Z
	result := NewMesh()
	m.Iterate(func(t *Triangle) {
		t1 := *t
		result.Add(&t1)
	})

	angleZ := math.Cos(maxAngle)
	shouldFlatten := func(t *Triangle) bool {
		var minCount int
		for _, c := range t {
			if c.Z == minZ {
				minCount++
			}
		}
		return minCount == 2 && -t.Normal().Z > angleZ
	}

	pending := map[*Triangle]bool{}
	result.Iterate(func(t *Triangle) {
		if shouldFlatten(t) {
			pending[t] = true
		}
	})

	flattenCoord := func(c Coord3D) {
		newC := c
		newC.Z = minZ
		v2t := result.getVertexToFace()
		for _, t2 := range v2t.Value(c) {
			for i, c1 := range t2 {
				if c1 == c {
					t2[i] = newC
					break
				}
			}
			if shouldFlatten(t2) {
				pending[t2] = true
			} else {
				delete(pending, t2)
			}
		}
		v2t.Store(newC, v2t.Value(c))
		v2t.Delete(c)
	}

	for len(pending) > 0 {
		oldPending := []*Triangle{}
		for t := range pending {
			oldPending = append(oldPending, t)
		}
		pending = map[*Triangle]bool{}
		for _, t := range oldPending {
			for _, c := range t {
				if c.Z != minZ {
					flattenCoord(c)
				}
			}
		}
	}

	return result
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
	m.getVertexToFace().KeyRange(func(c Coord3D) bool {
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
			return true
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
		return true
	})

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
	counts := NewEdgeToInt()
	for face := range m.faces {
		for i := 0; i < 3; i++ {
			seg := NewSegment(face[i], face[(i+1)%3])
			if counts.Add(seg, 1) > 2 {
				return true
			}
		}
	}

	result := false
	counts.ValueRange(func(count int) bool {
		if count == 2 {
			return true
		} else {
			result = true
			return false
		}
	})
	return result
}

// SingularVertices gets the points at which the mesh is
// squeezed to zero volume. In other words, it gets the
// points where two pieces of volume are barely touching
// by a single point.
func (m *Mesh) SingularVertices() []Coord3D {
	var res []Coord3D

	// Queues reused for each vertex to store the current
	// search state, where each entry is a triangle index.
	var unvisited []int
	var visitQueue []int

	m.getVertexToFace().Range(func(vertex Coord3D, tris []*Triangle) bool {
		if len(tris) == 0 {
			return true
		}

		// Start the search with triangle 0 in the queue.
		unvisited = unvisited[:0]
		for i := range tris[1:] {
			unvisited = append(unvisited, i+1)
		}
		visitQueue = append(visitQueue[:0], 0)

		for len(unvisited) > 0 && len(visitQueue) > 0 {
			expandIdx := visitQueue[len(visitQueue)-1]
			expandTri := tris[expandIdx]
			visitQueue = visitQueue[:len(visitQueue)-1]

			for i := 0; i < len(unvisited); i++ {
				visitIdx := unvisited[i]
				if expandTri.SharesEdge(tris[visitIdx]) {
					unvisited[i] = unvisited[len(unvisited)-1]
					unvisited = unvisited[:len(unvisited)-1]
					visitQueue = append(visitQueue, visitIdx)
					i--
				}
			}
		}

		// If we have visited every triangle, then all the
		// triangles for the vertex are connected to each
		// other via edges.
		if len(unvisited) != 0 {
			res = append(res, vertex)
		}
		return true
	})
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

// Orientable checks if the mesh is orientable.
// This will return true for closed, manifold meshes, and
// false for shapes like Mobius strips.
func (m *Mesh) Orientable() bool {
	return m.maybeFaceOrientations() != nil
}

// InconsistentEdges finds edges which are ordered the same
// way for at least two triangles, indicating an
// inconsistent mesh orientation.
//
// For properly connected (i.e. manifold) meshes, this can
// determine if the surface normals are consistent.
func (m *Mesh) InconsistentEdges() [][2]Coord3D {
	edges := map[[2]Coord3D]int{}
	m.Iterate(func(t *Triangle) {
		for _, edge := range triangleEdges(t) {
			edges[edge]++
		}
	})
	var res [][2]Coord3D
	for edge, count := range edges {
		if count > 1 {
			res = append(res, edge)
		}
	}
	return res
}

// FaceOrientations returns, for each group of connected
// faces, the relative orientation of every face.
// For meshes with correct normals, all faces of each group
// should have the same orientation. The values can be
// thought of as flags indicating whether or not each
// triangle should be flipped.
//
// This method should only be called on orientable,
// manifold meshes. If not, it may panic().
// One intended use case is to re-orient faces to repair
// surfaces with incorrect normals.
func (m *Mesh) FaceOrientations() []map[*Triangle]bool {
	orientations := m.maybeFaceOrientations()
	if orientations == nil {
		panic("mesh is not orientable")
	}
	return orientations
}

func (m *Mesh) maybeFaceOrientations() []map[*Triangle]bool {
	remaining := map[*Triangle]bool{}
	m.Iterate(func(t *Triangle) {
		remaining[t] = true
	})

	groups := []map[*Triangle]bool{}
	for len(remaining) > 0 {
		var startTri *Triangle
		for t := range remaining {
			startTri = t
			break
		}
		delete(remaining, startTri)

		group := map[*Triangle]bool{startTri: false}
		seenEdges := map[[2]Coord3D]bool{}
		for _, edge := range triangleEdges(startTri) {
			seenEdges[edge] = true
		}

		queue := m.Neighbors(startTri)
		for _, t := range queue {
			delete(remaining, t)
		}
		for len(queue) > 0 {
			next := queue[0]
			queue = queue[1:]

			foundFlipped := false
			foundUnflipped := false
			for _, edge := range triangleEdges(next) {
				if seenEdges[edge] {
					foundUnflipped = true
				}
				if seenEdges[[2]Coord3D{edge[1], edge[0]}] {
					foundFlipped = true
				}
			}
			if !foundFlipped && !foundUnflipped {
				panic("impossible case detected")
			} else if foundFlipped && foundUnflipped {
				return nil
			}
			group[next] = foundUnflipped
			// The triangle will be flipped, so we register the
			// flipped version of all its edges.
			for _, edge := range triangleEdges(next) {
				if foundUnflipped {
					edge[0], edge[1] = edge[1], edge[0]
				}
				if seenEdges[edge] {
					return nil
				}
				seenEdges[edge] = true
			}
			for _, neighbor := range m.Neighbors(next) {
				if remaining[neighbor] {
					delete(remaining, neighbor)
					queue = append(queue, neighbor)
				}
			}
		}
		groups = append(groups, group)
	}
	return groups
}

// RepairNormalsMajority leverages FaceOrientations to flip
// the fewest faces so that every connected group of faces
// is oriented consistently.
//
// Like FaceOrientations(), this should only be called for
// orientable, manifold meshes.
//
// Note that this makes no guarantees about correctness,
// and will give exactly opposite results when more than
// half of all faces are flipped incorrectly.
// For global correctness, see RepairNormals().
func (m *Mesh) RepairNormalsMajority() (*Mesh, int) {
	orientations := m.FaceOrientations()
	flipFlags := map[*Triangle]bool{}
	for _, group := range orientations {
		numTrue := 0
		for _, value := range group {
			if value {
				numTrue++
			}
		}
		invert := numTrue > len(group)/2
		for k, v := range group {
			flipFlags[k] = v == !invert
		}
	}
	res := NewMesh()
	var count int
	m.Iterate(func(t *Triangle) {
		t1 := *t
		if flipFlags[t] {
			count++
			t1[0], t1[1] = t1[1], t1[0]
		}
		res.Add(&t1)
	})
	return res, count
}

func triangleEdges(t *Triangle) [3][2]Coord3D {
	return [3][2]Coord3D{
		{t[0], t[1]},
		{t[1], t[2]},
		{t[2], t[0]},
	}
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
	solid := NewColliderSolid(collider)
	numFlipped := 0
	newMesh := NewMesh()

	m.Iterate(func(t *Triangle) {
		t1 := *t
		normal := t.Normal()
		center := t[0].Add(t[1]).Add(t[2]).Scale(1.0 / 3)
		movedOut := center.Add(normal.Scale(epsilon))
		if solid.Contains(movedOut) {
			numFlipped++
			t1[0], t1[1] = t1[1], t1[0]
		}
		newMesh.Add(&t1)
	})
	return newMesh, numFlipped
}

// FlipDelaunay "flips" edges in triangle pairs until the
// mesh is Delaunay.
//
// This can be used to prepare a mesh for cotangent
// weights, either for smoothing, deformation, or some
// other mesh operation.
//
// See: https://arxiv.org/abs/math/0503219.
func (m *Mesh) FlipDelaunay() *Mesh {
	res := m.Copy()
	changed := true
	for changed {
		changed = false
		res.Iterate(func(t *Triangle) {
			for _, seg := range t.Segments() {
				tris := res.Find(seg[0], seg[1])
				if len(tris) != 2 {
					return
				}
				var sum float64
				for _, t := range tris {
					other := seg.Other(t)
					v1 := seg[0].Sub(other)
					v2 := seg[1].Sub(other)
					sum += math.Acos(v1.Normalize().Dot(v2.Normalize()))
				}
				if sum < math.Pi+1e-8 {
					continue
				}
				//
				//     p2
				//    /  \
				//  o1 -- o2
				//    \  /
				//     p1
				//
				p1, p2 := seg[0], seg[1]
				o1, o2 := seg.Other(tris[0]), seg.Other(tris[1])
				if (&Triangle{o1, p1, p2}).Normal().Dot(tris[0].Normal()) < 0 {
					p1, p2 = p2, p1
				}
				res.Remove(tris[0])
				res.Remove(tris[1])
				res.Add(&Triangle{o1, o2, p2})
				res.Add(&Triangle{p1, o2, o1})
				changed = true
				break
			}
		})
	}
	return res
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

// EliminateCoplanar eliminates vertices whose neigboring
// triangles are all co-planar.
//
// The epsilon argument controls how close two normals
// must be for the triangles to be considered coplanar.
// A good value for very precise results is 1e-8.
func (m *Mesh) EliminateCoplanar(epsilon float64) *Mesh {
	dec := &decimator{
		FeatureAngle:       math.Acos(1 - epsilon),
		MinimumAspectRatio: 0.01,
		Criterion: &normalDecCriterion{
			CosineEpsilon: epsilon,
		},
	}
	return dec.Decimate(m)
}

// EliminateCoplanarFiltered is like EliminateCoplanar,
// but vertices are only removed if f(vertex) is true.
func (m *Mesh) EliminateCoplanarFiltered(epsilon float64, f func(Coord3D) bool) *Mesh {
	dec := &decimator{
		FeatureAngle:       math.Acos(1 - epsilon),
		MinimumAspectRatio: 0.01,
		Criterion: &normalDecCriterion{
			CosineEpsilon: epsilon,
			FilterFunc:    f,
		},
	}
	return dec.Decimate(m)
}

func canEliminateSegment(m *Mesh, seg Segment) bool {
	// Segment removal must not leave either point
	// from the segment in the mesh.
	if seg[0] == seg[1] {
		return false
	}

	v2t := m.getVertexToFace()
	neighbors1 := v2t.Value(seg[0])
	neighbors2 := v2t.Value(seg[1])
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
	v2t := m.getVertexToFace()
	newNeighbors := []*Triangle{}
	for i, segmentPoint := range segment {
		for _, neighbor := range v2t.Value(segmentPoint) {
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
					delete(m.faces, neighbor)
					for _, p := range neighbor {
						if p != segment[0] && p != segment[1] {
							m.removeFaceFromVertex(v2t, neighbor, p)
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
			v2t.Delete(segmentPoint)
		}
	}
	v2t.Store(mp, newNeighbors)
}
