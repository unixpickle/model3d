package model2d

import (
	"math"
	"sort"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/splaytree"
)

// TriangulateMesh creates a minimal collection of
// triangles that cover the enclosed region of a mesh.
//
// The mesh must be manifold, non-intersecting, and have
// the correct orientation (i.e. correct normals).
// The mesh may have holes, and is assumed to obey the
// even-odd rule for containment.
// If the mesh does not meet the expected criteria, the
// behavior of TriangulateMesh is undefined and may result
// in a panic.
//
// The vertices of the resulting triangles are ordered
// clockwise (assuming a y-axis that points upward).
// This way, each triangle can itself be considered a
// minimal, correctly-oriented mesh.
func TriangulateMesh(m *Mesh) [][3]Coord {
	m, inv := misalignMesh(m)
	hierarchies := uncheckedMeshToHierarchy(m)
	tris := [][3]Coord{}
	for _, h := range hierarchies {
		tris = append(tris, triangulateHierarchy(h)...)
	}
	for i, t := range tris {
		for j, c := range t {
			t[j] = inv(c)
		}
		tris[i] = t
	}
	return tris
}

// Triangulate turns any simple polygon into a set of
// equivalent triangles.
//
// The polygon is passed as a series of points, in order.
// The first point is re-used as the ending point, so no
// ending should be explicitly specified.
//
// Unlike TriangulateMesh, the order of the coordinates
// needn't be clockwise, and the orientation of the
// resulting triangles is undefined.
//
// This should only be used for polygons with several
// vertices. For more complex shapes, use TriangulateMesh.
func Triangulate(polygon []Coord) [][3]Coord {
	polygon = removeColinearPoints(polygon)

	if len(polygon) == 3 {
		return [][3]Coord{{polygon[0], polygon[1], polygon[2]}}
	} else if len(polygon) < 3 {
		panic("polygon does not span a 2-D space")
	}

	for i := range polygon {
		if isVertexEar(polygon, i) {
			p1 := polygon[(i+len(polygon)-1)%len(polygon)]
			p3 := polygon[(i+1)%len(polygon)]
			newPoly := append([]Coord{}, polygon...)
			essentials.OrderedDelete(&newPoly, i)
			return append(Triangulate(newPoly), [3]Coord{p1, polygon[i], p3})
		}
	}
	panic("no ears detected")
}

func isVertexEar(polygon []Coord, vertex int) bool {
	clockwise := isPolygonClockwise(polygon)

	idx1 := (vertex + len(polygon) - 1) % len(polygon)
	idx3 := (vertex + 1) % len(polygon)
	p1 := polygon[idx1]
	p2 := polygon[vertex]
	p3 := polygon[idx3]

	theta := clockwiseAngle(p1, p2, p3)

	if clockwise != (theta <= math.Pi) {
		// This is not an interior corner.
		return false
	}

	inverseMat := (&Matrix2{p1.X - p2.X, p3.X - p2.X, p1.Y - p2.Y, p3.Y - p2.Y}).Inverse()

	for i, p := range polygon {
		if i == idx1 || i == vertex || i == idx3 {
			continue
		}
		coords := inverseMat.MulColumn(p.Sub(p2))
		if coords.X > 0 && coords.Y > 0 && coords.X+coords.Y < 1 {
			// Another point lies inside this triangle.
			return false
		}
	}

	return true
}

// isPolygonClockwise checks if the polygon goes
// clockwise, assuming that the y-axis goes up and the
// x-axis goes to the right.
func isPolygonClockwise(polygon []Coord) bool {
	var sumTheta float64
	for i, p2 := range polygon {
		p1 := polygon[(i+len(polygon)-1)%len(polygon)]
		p3 := polygon[(i+1)%len(polygon)]
		sumTheta += math.Pi - clockwiseAngle(p1, p2, p3)
	}
	return sumTheta > 0
}

func clockwiseAngle(p1, p2, p3 Coord) float64 {
	v1 := p1.Sub(p2)
	v2 := p3.Sub(p2)
	n1 := v1.Normalize()
	n2 := v2.Normalize()
	cosTheta := n1.Dot(n2)

	// Prevent NaN due to rounding error.
	if cosTheta < -1 {
		cosTheta = -1
	} else if cosTheta > 1 {
		cosTheta = 1
	}

	theta := math.Acos(cosTheta)
	theta1 := 2*math.Pi - theta

	cos, sin := math.Cos(theta), math.Sin(theta)
	rot1 := Matrix2{cos, -sin, sin, cos}
	rot2 := Matrix2{cos, sin, -sin, cos}
	dot1 := rot1.MulColumn(n1).Dot(n2)
	dot2 := rot2.MulColumn(n1).Dot(n2)
	if math.Abs(1-dot1) < math.Abs(1-dot2) {
		return theta
	} else {
		return theta1
	}
}

func removeColinearPoints(poly []Coord) []Coord {
	var res []Coord
	for i, p2 := range poly {
		p1 := poly[(i+len(poly)-1)%len(poly)]
		p3 := poly[(i+1)%len(poly)]
		theta := clockwiseAngle(p1, p2, p3)
		if math.Abs(math.Sin(theta)) > 1e-8 {
			res = append(res, p2)
		}
	}
	return res
}

type triangulateVertexType int

const (
	triangulateVertexSplit triangulateVertexType = iota
	triangulateVertexMerge
	triangulateVertexStart
	triangulateVertexEnd
	triangulateVertexUpperChain
	triangulateVertexLowerChain
)

func triangulateHierarchy(m *MeshHierarchy) [][3]Coord {
	segments := m.Mesh.SegmentsSlice()
	for _, child := range m.Children {
		segments = append(segments, child.Mesh.SegmentsSlice()...)
	}
	tris := triangulateSingleMesh(newPtrMeshSegments(segments))
	for _, child := range m.Children {
		for _, childChild := range child.Children {
			tris = append(tris, triangulateHierarchy(childChild)...)
		}
	}
	return tris
}

// triangulateSingleMesh triangulates a mesh with the same
// restrictions as triangulateMonotoneDecomp().
func triangulateSingleMesh(m *ptrMesh) [][3]Coord {
	tris := [][3]Coord{}
	for _, m := range triangulateMonotoneDecomp(m) {
		tris = append(tris, triangulateMonotoneMesh(m)...)
	}
	return tris
}

// triangulateMonotoneMesh triangulates a monotone mesh.
func triangulateMonotoneMesh(m *ptrMesh) [][3]Coord {
	state := newTriangulateSweepState(m)

	stack := []*ptrCoord{state.Coords[0]}
	var stackType triangulateVertexType

	var triangles [][3]Coord

	for i, c := range state.Coords[1:] {
		vType := state.VertexType(c)
		if i == 0 {
			stackType = vType
			stack = append(stack, c)
			continue
		}
		if vType == triangulateVertexEnd {
			// Close off the remaining triangles
			for i := 0; i < len(stack)-1; i++ {
				tri := [3]Coord{stack[i].Coord, stack[i+1].Coord, c.Coord}
				if !isPolygonClockwise(tri[:]) {
					tri[0], tri[1] = tri[1], tri[0]
				}
				triangles = append(triangles, tri)
			}
			if i != len(state.Coords)-2 {
				panic("polygon was not monotone")
			}
			stack = nil
		} else if vType != stackType {
			// Create triangles across the entire chain.
			for i := 0; i < len(stack)-1; i++ {
				tri := [3]Coord{stack[i].Coord, stack[i+1].Coord, c.Coord}
				if !isPolygonClockwise(tri[:]) {
					tri[0], tri[1] = tri[1], tri[0]
				}
				triangles = append(triangles, tri)
			}
			stack = []*ptrCoord{stack[len(stack)-1], c}
			stackType = vType
		} else if stackType == triangulateVertexUpperChain ||
			stackType == triangulateVertexLowerChain {
			for len(stack) > 1 {
				i := len(stack) - 2
				interiorAngle := clockwiseAngle(stack[i].Coord, stack[i+1].Coord, c.Coord)
				if stackType == triangulateVertexLowerChain {
					interiorAngle = math.Pi*2 - interiorAngle
				}
				if interiorAngle >= math.Pi {
					// Once we hit a reflex vertex, we cannot
					// create any more interior triangles.
					break
				}
				if stackType == triangulateVertexUpperChain {
					triangles = append(triangles, [3]Coord{stack[i].Coord, stack[i+1].Coord, c.Coord})
				} else {
					triangles = append(triangles, [3]Coord{stack[i+1].Coord, stack[i].Coord, c.Coord})
				}
				stack = stack[:i+1]
			}
			stack = append(stack, c)
		} else {
			stack = append(stack, c)
		}
	}
	if len(stack) != 0 {
		panic("polygon was not monotone")
	}
	return triangles
}

// triangulateMonotoneDecomp creates monotone meshes that
// cover the entire mesh m.
//
// The are several assumptions on m:
//
//     * No two coordinates have the same x value
//     * No segments have infinite slope.
//     * This is either a simple polygon, or a polygon with one
//       depth of holes.
//     * The normals are correct.
//
// The mesh m will be destroyed in the process.
func triangulateMonotoneDecomp(m *ptrMesh) []*ptrMesh {
	splits := triangulateMonotoneSplits(m)
	for _, split := range splits {
		m.Add(split[0], split[1])
		m.Add(split[1], split[0])
	}

	var subMeshes []*ptrMesh

	// Reset every inner loop, but re-used to prevent
	// extra allocations.
	var subSegs []*Segment
	for m.Peek() != nil {
		subSegs = subSegs[:0]

		// Walk a polygon in order, and always chose the smallest
		// possible interior angle for each vertex.
		startPoint := m.Peek()
		seg := [2]*ptrCoord{startPoint, m.Outgoing(startPoint)[0]}
		for {
			subSegs = append(subSegs, &Segment{seg[0].Coord, seg[1].Coord})
			m.Remove(seg[0], seg[1])
			if seg[1] == startPoint {
				break
			}
			potentialSegs := [][2]*ptrCoord{}
			for _, c1 := range m.Outgoing(seg[1]) {
				if c1 != seg[0] {
					potentialSegs = append(potentialSegs, [2]*ptrCoord{seg[1], c1})
				}
			}
			seg = segmentWithSmallestAngle(seg[0], potentialSegs)
		}
		subMeshes = append(subMeshes, newPtrMeshSegments(subSegs))
	}
	return subMeshes
}

func segmentWithSmallestAngle(p1 *ptrCoord, potential [][2]*ptrCoord) [2]*ptrCoord {
	if len(potential) == 0 {
		panic("mesh was non-manifold")
	}
	bestAngle := 10.0
	bestSegment := potential[0]
	for _, s := range potential {
		theta := clockwiseAngle(p1.Coord, s[0].Coord, s[1].Coord)
		if theta < bestAngle {
			bestAngle = theta
			bestSegment = s
		}
	}
	return bestSegment
}

// triangulateMonotoneSplits creates segments which induce
// monotone sub-polygons for a polygon.
//
// See triangulateMonotoneDecomp() for restrictions.
func triangulateMonotoneSplits(m *ptrMesh) [][2]*ptrCoord {
	state := newTriangulateSweepState(m)
	for !state.Done() {
		state.Next()
	}
	return state.Generated
}

// triangulateSweepState tracks the process of a sweep
// along the x-axis while decomposing a polygon into
// monotone polygons.
//
// The implementation is based on the analysis from
// "CMSC 754: Lecture 5 - Polygon Triangulation"
// (http://www.cs.umd.edu/class/spring2020/cmsc754/Lects/lect05-triangulate.pdf).
type triangulateSweepState struct {
	Mesh       *ptrMesh
	Coords     []*ptrCoord
	CurrentIdx int

	EdgeTree *triangulateEdgeTree

	// Helpers maps upper segments to a coordinate that can
	// be connected to any point under that segment.
	// However, since segments are uniquely defined by
	// their first point (in a manifold mesh), this simply
	// maps first vertices to helper vertices.
	Helpers map[*ptrCoord]*ptrCoord

	Generated [][2]*ptrCoord
}

func newTriangulateSweepState(m *ptrMesh) *triangulateSweepState {
	var vertices []*ptrCoord
	m.IterateCoords(func(c *ptrCoord) {
		vertices = append(vertices, c)
	})
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].Coord.X < vertices[j].Coord.X
	})

	state := &triangulateSweepState{
		Mesh:       m,
		Coords:     vertices,
		CurrentIdx: -1,
		EdgeTree:   &triangulateEdgeTree{},
		Helpers:    map[*ptrCoord]*ptrCoord{},
	}
	if state.VertexType(state.Coords[0]) != triangulateVertexStart {
		panic("invalid initial vertex type")
	}
	return state
}

func (t *triangulateSweepState) Done() bool {
	return t.CurrentIdx+1 >= len(t.Coords)
}

func (t *triangulateSweepState) Next() {
	t.CurrentIdx++
	v := t.Coords[t.CurrentIdx]

	switch t.VertexType(v) {
	case triangulateVertexStart:
		s1, s2 := t.findEdges(v)
		t.EdgeTree.Insert(s1)
		t.EdgeTree.Insert(s2)
		t.Helpers[triangulateHigherSegment(s1, s2)[0]] = v
	case triangulateVertexEnd:
		s1, s2 := t.findEdges(v)
		t.fixUp(v, triangulateHigherSegment(s1, s2))
		t.removeEdges(s1, s2)
	case triangulateVertexUpperChain:
		oldEdge, newEdge := t.findEdges(v)
		t.fixUp(v, oldEdge)
		t.removeEdges(oldEdge)
		t.EdgeTree.Insert(newEdge)
		t.Helpers[newEdge[0]] = v
	case triangulateVertexLowerChain:
		newEdge, oldEdge := t.findEdges(v)
		t.removeEdges(oldEdge)
		aboveEdge := t.EdgeTree.FindAbove(v.Coord)
		t.fixUp(v, aboveEdge)
		t.EdgeTree.Insert(newEdge)
		t.Helpers[aboveEdge[0]] = v
	case triangulateVertexSplit:
		s1, s2 := t.findEdges(v)
		above := t.EdgeTree.FindAbove(v.Coord)
		helper := t.Helpers[above[0]]
		t.Generated = append(t.Generated, [2]*ptrCoord{helper, v})
		for _, seg := range [][2]*ptrCoord{s1, s2} {
			t.EdgeTree.Insert(seg)
		}
		lower := triangulateLowerSegment(s1, s2)
		t.Helpers[lower[0]] = v
		t.Helpers[above[0]] = v
	case triangulateVertexMerge:
		s1, s2 := t.findEdges(v)
		lowerEdge := triangulateLowerSegment(s1, s2)
		t.fixUp(v, lowerEdge)
		t.removeEdges(s1, s2)
		above := t.EdgeTree.FindAbove(v.Coord)
		t.fixUp(v, above)
		t.Helpers[above[0]] = v
	}
}

func (t *triangulateSweepState) VertexType(v *ptrCoord) triangulateVertexType {
	s1, s2 := t.findEdges(v)
	c := v.Coord
	start, end := s1[0].Coord, s2[1].Coord
	if start.X == end.X || start.X == c.X || end.X == c.X {
		panic("no x values should be exactly equal")
	}
	if start.X > c.X && end.X > c.X {
		if triangulateHigherSegment(s1, s2) == s2 {
			return triangulateVertexStart
		} else {
			return triangulateVertexSplit
		}
	} else if start.X < c.X && end.X < c.X {
		if triangulateHigherSegment(s1, s2) == s2 {
			return triangulateVertexMerge
		} else {
			return triangulateVertexEnd
		}
	} else {
		if start.X > end.X {
			return triangulateVertexLowerChain
		} else {
			return triangulateVertexUpperChain
		}
	}
}

func (t *triangulateSweepState) fixUp(c *ptrCoord, s [2]*ptrCoord) {
	helper, ok := t.Helpers[s[0]]
	if !ok {
		panic("no helper found")
	}
	if t.VertexType(helper) == triangulateVertexMerge {
		t.Generated = append(t.Generated, [2]*ptrCoord{c, helper})
	}
}

func (t *triangulateSweepState) findEdges(c *ptrCoord) (s1, s2 [2]*ptrCoord) {
	s1 = [2]*ptrCoord{firstOfExactlyOne(t.Mesh.Incoming(c)), c}
	s2 = [2]*ptrCoord{c, firstOfExactlyOne(t.Mesh.Outgoing(c))}
	return
}

func (t *triangulateSweepState) removeEdges(segs ...[2]*ptrCoord) {
	for _, seg := range segs {
		t.EdgeTree.Delete(seg)
		delete(t.Helpers, seg[0])
	}
}

// triangulateEdgeTree sorts edges in a mesh along the
// y-axis, for efficient modification and lookup.
//
// Edges can be sorted in this way so long as they do not
// intersect, and all overlap in their x-axis projections.
type triangulateEdgeTree struct {
	Tree splaytree.Tree
}

func (t *triangulateEdgeTree) Insert(s [2]*ptrCoord) {
	t.Tree.Insert(newSortedEdge(s))
}

func (t *triangulateEdgeTree) Delete(s [2]*ptrCoord) {
	t.Tree.Delete(newSortedEdge(s))
}

func (t *triangulateEdgeTree) FindAbove(c Coord) [2]*ptrCoord {
	res, _ := t.findAbove(t.Tree.Root, c)
	return res
}

func (t *triangulateEdgeTree) findAbove(n *splaytree.Node, c Coord) ([2]*ptrCoord, bool) {
	if n == nil {
		return [2]*ptrCoord{}, false
	}
	comp := n.Value.(*sortedEdge).ComparePoint(c)
	if comp == -1 {
		// This node is below the vertex.
		return t.findAbove(n.Right, c)
	} else if comp == 1 {
		res, ok := t.findAbove(n.Left, c)
		if !ok {
			return n.Value.(*sortedEdge).Segment, true
		}
		return res, true
	} else {
		panic("vertex intersects an edge in the state")
	}
}

type sortedEdge struct {
	Segment [2]*ptrCoord

	minX    float64
	yAtMinX float64
	maxX    float64
	yAtMaxX float64
	slope   float64
}

func newSortedEdge(s [2]*ptrCoord) *sortedEdge {
	minX, yAtMinX := s[0].Coord.X, s[0].Coord.Y
	maxX, yAtMaxX := s[1].Coord.X, s[1].Coord.Y
	if minX > maxX {
		minX, maxX = maxX, minX
		yAtMinX, yAtMaxX = yAtMaxX, yAtMinX
	}
	return &sortedEdge{
		Segment: s,
		minX:    minX,
		yAtMinX: yAtMinX,
		maxX:    maxX,
		yAtMaxX: yAtMaxX,
		slope:   (yAtMaxX - yAtMinX) / (maxX - minX),
	}
}

// Compare compares two edges in terms of y value,
// assuming the edges overlap in the x axis but do not
// intersect.
func (s *sortedEdge) Compare(other splaytree.Value) int {
	s1 := other.(*sortedEdge)
	if s1.Segment == s.Segment {
		return 0
	}

	// Deal with segments that have the same endpoint.
	if s1.minX == s.minX {
		if s.slope == s1.slope {
			panic("segments overlap with same slope")
		} else if s.slope > s1.slope {
			return 1
		} else {
			return -1
		}
	} else if s1.maxX == s.maxX {
		if s.slope == s1.slope {
			panic("segments overlap with same slope")
		} else if s.slope < s1.slope {
			return 1
		} else {
			return -1
		}
	}

	var sY, s1Y float64
	if s1.minX > s.minX {
		sY = s.yAtX(s1.minX)
		s1Y = s1.yAtMinX
	} else {
		sY = s.yAtMinX
		s1Y = s1.yAtX(s.minX)
	}
	if sY > s1Y {
		return 1
	} else if sY < s1Y {
		return -1
	} else {
		panic("edges intersect")
	}
}

// ComparePoint compares the edge against a point.
// It returns 1 if the edge is above the point, -1 if
// below, and 0 if they intersect.
func (s *sortedEdge) ComparePoint(c Coord) int {
	if c == s.Segment[0].Coord || c == s.Segment[1].Coord {
		return 0
	}
	if c.X < s.minX || c.X > s.maxX {
		panic("coordinate out of bounds")
	}
	y := s.yAtX(c.X)
	if y > c.Y {
		return 1
	} else if y < c.Y {
		return -1
	} else {
		return 0
	}
}

func (s *sortedEdge) yAtX(x float64) float64 {
	// Return exact correct values at endpoints to avoid
	// edge cases from rounding error.
	if x == s.minX {
		return s.yAtMinX
	} else if x == s.maxX {
		return s.yAtMaxX
	}
	fraction := (x - s.minX) / (s.maxX - s.minX)
	return fraction*s.yAtMaxX + (1-fraction)*s.yAtMinX
}

func triangulateHigherSegment(s1, s2 [2]*ptrCoord) [2]*ptrCoord {
	if newSortedEdge(s1).Compare(newSortedEdge(s2)) == 1 {
		return s1
	} else {
		return s2
	}
}

func triangulateLowerSegment(s1, s2 [2]*ptrCoord) [2]*ptrCoord {
	if triangulateHigherSegment(s1, s2) == s1 {
		return s2
	} else {
		return s1
	}
}

func firstOfExactlyOne(coords []*ptrCoord) *ptrCoord {
	if len(coords) != 1 {
		panic("mesh is non-manifold")
	}
	return coords[0]
}

// misalignMesh rotates the mesh by a random angle to
// prevent vertices from directly aligning on the x or
// y axes.
func misalignMesh(m *Mesh) (misaligned *Mesh, inv func(Coord) Coord) {
	invMapping := map[Coord]Coord{}
	xAxis := NewCoordPolar(0.5037616150469717, 1.0)
	yAxis := XY(-xAxis.Y, xAxis.X)
	misaligned = m.MapCoords(func(c Coord) Coord {
		c1 := XY(xAxis.Dot(c), yAxis.Dot(c))
		invMapping[c1] = c
		return c1
	})
	inv = func(c Coord) Coord {
		if res, ok := invMapping[c]; ok {
			return res
		} else {
			panic("coordinate was not in the misaligned output")
		}
	}
	return misaligned, inv
}
