package model2d

/***************************************
 * NOTE: based off of model3d/mc.go on *
 * Apr 18, 2020.                       *
 ***************************************/

// MarchingSquares turns a Solid into a mesh using a 2D
// version of the marching cubes algorithm.
func MarchingSquares(s Solid, delta float64) *Mesh {
	table := msLookupTable()

	spacer := newSquareSpacer(s, delta)
	bottomCache := newSolidCache(s, spacer)
	topCache := newSolidCache(s, spacer)
	topCache.FetchY(0)

	mesh := NewMesh()

	for y := 1; y < len(spacer.Ys); y++ {
		bottomCache, topCache = topCache, bottomCache
		topCache.FetchY(y)

		for x := 0; x < len(spacer.Xs)-1; x++ {
			bits := bottomCache.GetSegment(x) | (topCache.GetSegment(x) << 2)
			segments := table[bits]
			if len(segments) > 0 {
				min := spacer.CornerCoord(x, y-1)
				max := spacer.CornerCoord(x+1, y)
				corners := msCornerCoordinates(min, max)
				for _, s := range segments {
					mesh.Add(s.Segment(corners))
				}
			}
		}
	}

	return mesh
}

// msCorner represents a corner on a square.
// The x-axis is the first bit, and y-axis is the second
// bit, so for example (0, 1) is 0b10.
//
//     2 +-----+ 3
//       |     |
//       |     |
//     0 +-----+ 1
//
type msCorner uint8

// msCornerCoordinates gets the coordinates of all four
// corners for a square.
func msCornerCoordinates(min, max Coord) [4]Coord {
	return [4]Coord{
		min,
		{X: max.X, Y: min.Y},
		{X: min.X, Y: max.Y},
		max,
	}
}

// msIntersections stores a bitmap of which corners of a
// square are contained in an object.
//
// The bits are:
//
//     1: the corner (0, 0)
//     2: the corner (1, 0)
//     4: the corner (0, 1)
//     8: the corner (1, 1)
//
type msIntersections uint8

func newMsIntersections(corners ...msCorner) msIntersections {
	var res msIntersections
	for _, x := range corners {
		res |= 1 << x
	}
	return res
}

// msSegment is a segment constructed out of midpoints of
// edges of a square.
// There are 4 corners because each pair of two represents
// an edge.
//
// The segment is ordered in clockwise order circling
// around the outside of the mesh.
type msSegment [4]msCorner

// Segment creates a real segment out of the msSegment,
// given the corner coordinates.
func (m msSegment) Segment(corners [4]Coord) *Segment {
	return &Segment{
		corners[m[0]].Mid(corners[m[1]]),
		corners[m[2]].Mid(corners[m[3]]),
	}
}

// msLookupTable creates a lookup table for the marching
// squares algorithm.
func msLookupTable() [16][]msSegment {
	mapping := map[msIntersections][]msSegment{
		newMsIntersections():     []msSegment{},
		newMsIntersections(0):    []msSegment{{0, 2, 0, 1}},
		newMsIntersections(1):    []msSegment{{0, 1, 1, 3}},
		newMsIntersections(2):    []msSegment{{2, 3, 0, 2}},
		newMsIntersections(3):    []msSegment{{1, 3, 2, 3}},
		newMsIntersections(0, 1): []msSegment{{0, 2, 1, 3}},
		newMsIntersections(0, 2): []msSegment{{2, 3, 0, 1}},
		// Resolve ambiguities and don't create holes
		// by covering both cases explicitly.
		newMsIntersections(0, 3): []msSegment{{0, 2, 0, 1}, {1, 3, 2, 3}},
		newMsIntersections(1, 2): []msSegment{{0, 1, 1, 3}, {2, 3, 0, 2}},
	}

	// Add inverses to complete the table.
	for ints, segs := range mapping {
		invInts := 0xf ^ ints
		if _, ok := mapping[invInts]; !ok {
			revSegs := make([]msSegment, len(segs))
			for i, seg := range segs {
				revSegs[i] = msSegment{seg[2], seg[3], seg[0], seg[1]}
			}
			mapping[invInts] = revSegs
		}
	}
	if len(mapping) != 16 {
		panic("incorrect number of cases")
	}

	res := [16][]msSegment{}
	for x, y := range mapping {
		res[x] = y
	}
	return res
}

type squareSpacer struct {
	Xs []float64
	Ys []float64
}

func newSquareSpacer(s Solid, delta float64) *squareSpacer {
	var xs, ys []float64
	min := s.Min()
	max := s.Max()
	for x := min.X - delta; x <= max.X+delta; x += delta {
		xs = append(xs, x)
	}
	for y := min.Y - delta; y <= max.Y+delta; y += delta {
		ys = append(ys, y)
	}
	return &squareSpacer{Xs: xs, Ys: ys}
}

func (s *squareSpacer) CornerCoord(x, y int) Coord {
	return Coord{X: s.Xs[x], Y: s.Ys[y]}
}

type solidCache struct {
	spacer *squareSpacer
	solid  Solid
	values []bool
}

func newSolidCache(solid Solid, spacer *squareSpacer) *solidCache {
	return &solidCache{
		spacer: spacer,
		solid:  solid,
		values: make([]bool, len(spacer.Xs)),
	}
}

func (s *solidCache) FetchY(y int) {
	maxX := len(s.spacer.Xs) - 1
	onEdge := y == 0 || y == len(s.spacer.Ys)-1

	var idx int
	for i := 0; i < len(s.spacer.Xs); i++ {
		b := s.solid.Contains(s.spacer.CornerCoord(i, y))
		s.values[idx] = b
		idx++
		if b && (onEdge || i == 0 || i == maxX) {
			panic("solid is true outside of bounds")
		}
	}
}

func (s *solidCache) GetSegment(x int) msIntersections {
	var result msIntersections
	if s.values[x] {
		result |= 1
	}
	if s.values[x+1] {
		result |= 2
	}
	return result
}
