package model2d

import (
	"math"

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

type triangulateSweepState struct {
	Coords     []Coord
	CurrentIdx int

	EdgeTree *triangulateEdgeTree
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
