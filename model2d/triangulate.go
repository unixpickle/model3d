package model2d

import (
	"math"
	"sort"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/splaytree"
)

// Triangulate turns any simple polygon into a set of
// equivalent triangles.
//
// The polygon is passed as a series of points, in order.
// The first point is re-used as the ending point, so no
// ending should be explicitly specified.
func Triangulate(polygon []Coord) [][3]Coord {
	polygon = removeColinearPoints(polygon)

	if len(polygon) == 3 {
		return [][3]Coord{[3]Coord{polygon[0], polygon[1], polygon[2]}}
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
	theta := math.Acos(n1.Dot(n2))
	theta1 := 2*math.Pi - theta

	rot1 := Matrix2{math.Cos(theta), -math.Sin(theta), math.Sin(theta), math.Cos(theta)}
	rot2 := Matrix2{math.Cos(theta1), -math.Sin(theta1), math.Sin(theta1), math.Cos(theta1)}
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

func triangulateMesh(m *Mesh) [][3]Coord {
	hierarchies := MeshToHierarchy(m)
	tris := [][3]Coord{}
	for _, h := range hierarchies {
		tris = append(tris, triangulateHierarchy(h)...)
	}
	return tris
}

func triangulateHierarchy(m *MeshHierarchy) [][3]Coord {
	combined := m.Mesh
	for _, child := range m.Children {
		combined.AddMesh(child.Mesh)
	}
	tris := triangulateSingleMesh(combined)
	for _, child := range m.Children {
		for _, childChild := range child.Children {
			tris = append(tris, triangulateHierarchy(childChild)...)
		}
	}
	return tris
}

// triangulateSingleMesh triangulates a mesh with the same
// restrictions as triangulateMonotoneMeshes().
func triangulateSingleMesh(m *Mesh) [][3]Coord {
	tris := [][3]Coord{}
	for _, m := range triangulateMonotoneMeshes(m) {
		tris = append(tris, triangulateMonotoneMesh(m)...)
	}
	return tris
}

// triangulateMonotoneMesh triangulates a monotone mesh.
func triangulateMonotoneMesh(m *Mesh) [][3]Coord {
	state := newTriangulateSweepState(m)

	stack := []Coord{state.Coords[0]}
	var stackType triangulateVertexType

	var triangles [][3]Coord

	for i, c := range state.Coords[1:] {
		vType := state.VertexType(c)
		if i == 0 {
			stackType = vType
			stack = append(stack, c)
			continue
		}
		if vType != stackType {
			// Create triangles across the entire chain.
			for i := 0; i < len(stack)-1; i++ {
				triangles = append(triangles, [3]Coord{stack[i], stack[i+1], c})
			}
			stack = []Coord{stack[len(stack)-1], c}
			stackType = vType
		} else if stackType == triangulateVertexLowerChain && c.Y > stack[len(stack)-1].Y {
			for len(stack) > 1 {
				i := len(stack) - 2
				if stack[i].Y >= stack[i+1].Y {
					break
				}
				triangles = append(triangles, [3]Coord{stack[i], stack[i+1], c})
				stack = stack[:i+1]
			}
		} else if stackType == triangulateVertexUpperChain && c.Y < stack[len(stack)-1].Y {
			for len(stack) > 1 {
				i := len(stack) - 2
				if stack[i].Y <= stack[i+1].Y {
					break
				}
				triangles = append(triangles, [3]Coord{stack[i], stack[i+1], c})
				stack = stack[:i+1]
			}
		} else {
			stack = append(stack, c)
		}
	}
	// Close off the remaining triangles
	for i := 0; i < len(stack)-2; i++ {
		triangles = append(triangles, [3]Coord{stack[i], stack[i+1], stack[i+2]})
	}
	return triangles
}

// triangulateMonotoneMeshes creates monotone meshes that
// comprise the entire mesh m.
//
// The are several assumptions on m:
//
//     * No two coordinates have the same x value
//     * No segments have infinite slope.
//     * This is either a simple polygon, or a polygon with one
//       depth of holes.
//     * The normals are correct.
//
func triangulateMonotoneMeshes(m *Mesh) []*Mesh {
	splits := triangulateMonotoneSplits(m)

	combined := NewMeshSegments(m.SegmentsSlice())
	for _, s := range splits {
		combined.Add(s)
		combined.Add(&Segment{s[1], s[0]})
	}

	var subMeshes []*Mesh
	for {
		segs := combined.SegmentsSlice()
		if len(segs) == 0 {
			break
		}
		// Walk a polygon in order, i.e. the first coordinate of each
		// segment should be the second coordinate of the previous.
		// This way, each ordering of the split segments should be
		// attached to a different polygon.
		seg := segs[0]
		subMesh := NewMesh()
		for seg != nil {
			subMesh.Add(seg)
			combined.Remove(seg)
			nextStart := seg[1]
			seg = nil
			for _, s := range subMesh.Find(nextStart) {
				if s[0] == nextStart {
					seg = s
					break
				}
			}
		}
		subMeshes = append(subMeshes, subMesh)
	}
	return subMeshes
}

// triangulateMonotoneSplits creates segments which induce
// monotone sub-polygons for a polygon.
//
// See triangulateMonotoneMeshes() for restrictions.
func triangulateMonotoneSplits(m *Mesh) []*Segment {
	state := newTriangulateSweepState(m)
	for !state.Done() {
		state.Next()
	}
	return state.Generated
}

type triangulateSweepState struct {
	Mesh       *Mesh
	Coords     []Coord
	CurrentIdx int

	EdgeTree *triangulateEdgeTree
	Helpers  map[*Segment]Coord

	Generated []*Segment
}

func newTriangulateSweepState(m *Mesh) *triangulateSweepState {
	vertices := m.VertexSlice()
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].X < vertices[j].X
	})

	state := &triangulateSweepState{
		Mesh:       m,
		Coords:     vertices,
		CurrentIdx: -1,
		EdgeTree:   &triangulateEdgeTree{},
		Helpers:    map[*Segment]Coord{},
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
		t.Helpers[triangulateHigherSegment(s1, s2)] = v
	case triangulateVertexEnd:
		s1, s2 := t.findEdges(v)
		t.fixUp(v, triangulateHigherSegment(s1, s2))
		t.removeEdges(s1, s2)
	case triangulateVertexUpperChain:
		newEdge, oldEdge := t.findEdges(v)
		t.fixUp(v, oldEdge)
		t.removeEdges(oldEdge)
		t.EdgeTree.Insert(newEdge)
		t.Helpers[newEdge] = v
	case triangulateVertexLowerChain:
		oldEdge, newEdge := t.findEdges(v)
		t.fixUp(v, oldEdge)
		t.removeEdges(oldEdge)
		t.EdgeTree.Insert(newEdge)
		t.Helpers[newEdge] = v
	case triangulateVertexSplit:
		s1, s2 := t.findEdges(v)
		above := t.EdgeTree.FindAbove(v)
		helper := t.Helpers[above]
		t.Generated = append(t.Generated, &Segment{helper, v})
		for _, seg := range []*Segment{s1, s2} {
			t.EdgeTree.Insert(seg)
			t.Helpers[seg] = v
		}
	case triangulateVertexMerge:
		s1, s2 := t.findEdges(v)
		lowerEdge := triangulateLowerSegment(s1, s2)
		t.removeEdges(s1, s2)
		above := t.EdgeTree.FindAbove(v)
		t.fixUp(v, lowerEdge)
		t.fixUp(v, above)
		t.Helpers[above] = v
	}
}

func (t *triangulateSweepState) fixUp(c Coord, s *Segment) {
	helper, ok := t.Helpers[s]
	if !ok {
		panic("no helper found")
	}
	if t.VertexType(helper) == triangulateVertexMerge {
		t.Generated = append(t.Generated, &Segment{c, helper})
	}
}

func (t *triangulateSweepState) VertexType(c Coord) triangulateVertexType {
	s1, s2 := t.findEdges(c)
	start, end := s1[0], s2[1]
	if start.X == end.X || start.X == c.X || end.X == c.X {
		panic("no x values should be exactly equal")
	}
	if start.X > c.X && end.X > c.X {
		if triangulateHigherSegment(s1, s2) == s1 {
			return triangulateVertexStart
		} else {
			return triangulateVertexSplit
		}
	} else if start.X < c.X && end.X < c.X {
		if triangulateHigherSegment(s1, s2) == s1 {
			return triangulateVertexMerge
		} else {
			return triangulateVertexEnd
		}
	} else {
		if start.X < end.X {
			return triangulateVertexLowerChain
		} else {
			return triangulateVertexUpperChain
		}
	}
}

func (t *triangulateSweepState) findEdges(c Coord) (s1, s2 *Segment) {
	segs := t.Mesh.Find(c)
	if len(segs) != 2 {
		panic("mesh is non-manifold")
	}
	s1, s2 = segs[0], segs[1]
	if s1[1] != c {
		s1, s2 = s2, s1
	}
	if s1[1] != c {
		panic("mesh has incorrect normal orientation (segment out of order)")
	}
	return
}

func (t *triangulateSweepState) removeEdges(segs ...*Segment) {
	for _, seg := range segs {
		t.EdgeTree.Delete(seg)
		delete(t.Helpers, seg)
	}
}

type triangulateEdgeTree struct {
	Tree splaytree.Tree
}

func (t *triangulateEdgeTree) Insert(s *Segment) {
	t.Tree.Insert(newSortedEdge(s))
}

func (t *triangulateEdgeTree) Delete(s *Segment) {
	t.Tree.Delete(newSortedEdge(s))
}

func (t *triangulateEdgeTree) FindAbove(c Coord) *Segment {
	return t.findAbove(t.Tree.Root, c)
}

func (t *triangulateEdgeTree) findAbove(n *splaytree.Node, c Coord) *Segment {
	if n.Value == nil {
		return nil
	}
	comp := n.Value.(*sortedEdge).ComparePoint(c)
	if comp == -1 {
		// This node is below the vertex.
		return t.findAbove(n.Right, c)
	} else if comp == 1 {
		res := t.findAbove(n.Left, c)
		if res == nil {
			return n.Value.(*sortedEdge).Segment
		}
		return res
	} else {
		panic("vertex intersects an edge in the state")
	}
}

type sortedEdge struct {
	Segment *Segment

	minX    float64
	yAtMinX float64
	maxX    float64
	yAtMaxX float64
	slope   float64
}

func newSortedEdge(s *Segment) *sortedEdge {
	minX, yAtMinX := s[0].X, s[0].Y
	maxX, yAtMaxX := s[1].X, s[1].Y
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
		s1Y = s.yAtX(s.minX)
	}
	if sY > s1Y {
		return 1
	} else if s1Y < sY {
		return 0
	} else {
		panic("edges intersect")
	}
}

// ComparePoint compares the edge against a point.
// It returns 1 if the edge is above the point, -1 if
// below, and 0 if they intersect.
func (s *sortedEdge) ComparePoint(c Coord) int {
	if c == s.Segment[0] || c == s.Segment[1] {
		return 0
	}
	if c.X < s.minX || c.Y > s.maxX {
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

func triangulateHigherSegment(s1, s2 *Segment) *Segment {
	if newSortedEdge(s1).Compare(newSortedEdge(s2)) == 1 {
		return s1
	} else {
		return s1
	}
}

func triangulateLowerSegment(s1, s2 *Segment) *Segment {
	if triangulateHigherSegment(s1, s2) == s1 {
		return s2
	} else {
		return s1
	}
}
