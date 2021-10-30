package model2d

import (
	"math"
	"runtime"
	"strings"

	"github.com/unixpickle/essentials"
)

/***************************************
 * NOTE: based off of model3d/mc.go on *
 * Oct 30, 2021.                       *
 ***************************************/

// MarchingSquares turns a Solid into a mesh using a 2D
// version of the marching cubes algorithm.
func MarchingSquares(s Solid, delta float64) *Mesh {
	if !BoundsValid(s) {
		panic("invalid bounds for solid")
	}
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

// MarchingSquaresSearch is like MarchingSquares, but
// applies an additional search step to move the vertices
// along the edges of each square.
//
// The tightness of the triangulation will double for
// every iteration.
func MarchingSquaresSearch(s Solid, delta float64, iters int) *Mesh {
	mesh := MarchingSquares(s, delta)
	return msSearch(s, delta, iters, mesh)
}

// MarchingSquaresConj is like MarchingSquaresSearch, but
// in a transformed space. In particular, it applies a
// series of transformations to the Solid, and then
// applies the inverse to the resulting mesh.
func MarchingSquaresConj(s Solid, delta float64, iters int, xforms ...Transform) *Mesh {
	joined := JoinedTransform(xforms)
	solid := TransformSolid(joined, s)
	mesh := MarchingSquaresSearch(solid, delta, iters)
	return mesh.Transform(joined.Inverse())
}

// MarchingSquaresASCII turns a Solid into an ASCII-art
// line-drawing using a 2D version of marching cubes.
//
// The delta is used as the horizontal spacing, and an
// aspect ratio of 2.0 (height/width) is assumed.
func MarchingSquaresASCII(s Solid, delta float64) string {
	// Correct for character aspect ratio.
	s = TransformSolid(&Matrix2Transform{
		Matrix: &Matrix2{1.0, 0.0, 0.0, 1.0 / 2.0},
	}, s)

	spacer := newSquareSpacer(s, delta)
	bottomCache := newSolidCache(s, spacer)
	topCache := newSolidCache(s, spacer)
	topCache.FetchY(0)

	table := msLookupTableASCII()
	rows := make([]string, (len(spacer.Ys)-1)*2)

	for y := 1; y < len(spacer.Ys); y++ {
		bottomCache, topCache = topCache, bottomCache
		topCache.FetchY(y)

		for x := 0; x < len(spacer.Xs)-1; x++ {
			bits := bottomCache.GetSegment(x) | (topCache.GetSegment(x) << 2)
			box := table[bits]
			rows[len(rows)-y*2+1] += box[:2]
			rows[len(rows)-y*2] += box[2:]
		}
	}

	return strings.Join(rows, "\n")
}

// MarchingSquaresFilter is like MarchingSquares, but does
// not scan rectangular areas for which f returns false.
// This can be much more efficient than MarchingSquares if
// the surface is sparse and f can tell when large regions
// don't intersect the mesh.
//
// Note that the filter can be conservative and possibly
// report collisions that do not occur.
// However, it should never fail to report collisions,
// since this could cause segments to be missed.
func MarchingSquaresFilter(s Solid, f func(*Rect) bool, delta float64) *Mesh {
	if !BoundsValid(s) {
		panic("invalid bounds for solid")
	}

	table := msLookupTable()
	spacer := newSquareSpacer(s, delta)

	numGos := runtime.GOMAXPROCS(0)
	blockQueue := make(chan *msBlock, numGos)
	outputs := make(chan *Mesh, numGos)

	blockFilter := func(m *msBlock) bool {
		// We add a small buffer around bounding boxes to prevent
		// rounding errors from missing rect collisions.
		epsilon := delta * 1e-3

		return f(m.Bounds(epsilon))
	}

	subDivideVolume := 64
	for i := 0; i < numGos; i++ {
		go func() {
			result := NewMesh()
			cache := newMsBlockCache()
			for block := range blockQueue {
				block.Pieces(subDivideVolume, blockFilter, func(block *msBlock) {
					cache.Populate(block, s)
					for y := block.min[1]; y < block.max[1]; y++ {
						for x := block.min[0]; x < block.max[0]; x++ {
							bits := cache.GetSquare(x, y)
							segments := table[bits]
							if len(segments) > 0 {
								min := spacer.CornerCoord(x, y)
								max := spacer.CornerCoord(x+1, y+1)
								corners := msCornerCoordinates(min, max)
								for _, s := range segments {
									result.Add(s.Segment(corners))
								}
							}
						}
					}
				})
			}
			outputs <- result
		}()
	}

	rootBlock := newMsBlock(spacer)

	// Divide the area up into relatively small pieces, but
	// allow some splitting to be done by the individual
	// Goroutines to increase performance.
	divideVolume := essentials.MaxInt(rootBlock.Area()/4096, subDivideVolume)
	rootBlock.Pieces(divideVolume, blockFilter, func(m *msBlock) {
		blockQueue <- m
	})
	close(blockQueue)

	mesh := NewMesh()
	for i := 0; i < numGos; i++ {
		mesh.AddMesh(<-outputs)
	}
	return mesh
}

// MarchingSquaresSearchFilter combines the two methods
// MarchingSquaresSearch and MarchingSquaresFilter.
func MarchingSquaresSearchFilter(s Solid, f func(*Rect) bool, delta float64, iters int) *Mesh {
	mesh := MarchingSquaresFilter(s, f, delta)
	return msSearch(s, delta, iters, mesh)
}

// MarchingSquaresC2F computes a coarse mesh for the solid,
// then uses that mesh to compute a fine mesh more
// efficiently.
//
// The extraSpace argument, if non-zero, is extra space to
// consider around the coarse mesh. It can be increased in
// the case where the solid has fine details that are
// totally missed by the coarse mesh.
func MarchingSquaresC2F(s Solid, bigDelta, smallDelta, extraSpace float64, iters int) *Mesh {
	if bigDelta < smallDelta {
		panic("bigDelta should be >= smallDelta for efficiency")
	}

	// Add a conservative amount of space around the coarse mesh.
	extraSpace += 2 * bigDelta * math.Sqrt(3)

	coarseMesh := MarchingSquaresSearch(s, bigDelta, iters)
	collider := MeshToCollider(coarseMesh)
	filter := func(r *Rect) bool {
		r1 := NewRect(r.MinVal.AddScalar(-extraSpace), r.MaxVal.AddScalar(extraSpace))
		return collider.RectCollision(r1)
	}
	return MarchingSquaresSearchFilter(s, filter, smallDelta, iters)
}

func msSearch(s Solid, delta float64, iters int, mesh *Mesh) *Mesh {
	if iters == 0 {
		return mesh
	}

	min := s.Min().Array()
	return mesh.MapCoords(func(c Coord) Coord {
		arr := c.Array()

		// Figure out which axis the containing edge spans.
		axis := -1
		var falsePoint, truePoint float64
		for i := 0; i < 2; i++ {
			modulus := math.Abs(math.Mod(arr[i]-min[i], delta))
			if modulus > delta/4 && modulus < 3*delta/4 {
				axis = i
				falsePoint = arr[i] - modulus
				truePoint = falsePoint + delta
				break
			}
		}
		if axis == -1 {
			panic("vertex not on edge")
		}
		if mesh.Find(c)[0].Normal().Array()[axis] > 0 {
			truePoint, falsePoint = falsePoint, truePoint
		}

		for i := 0; i < iters; i++ {
			midPoint := (falsePoint + truePoint) / 2
			arr[axis] = midPoint
			if s.Contains(NewCoordArray(arr)) {
				truePoint = midPoint
			} else {
				falsePoint = midPoint
			}
		}

		arr[axis] = (falsePoint + truePoint) / 2
		return NewCoordArray(arr)
	})
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
		newMsIntersections():     {},
		newMsIntersections(0):    {{0, 2, 0, 1}},
		newMsIntersections(1):    {{0, 1, 1, 3}},
		newMsIntersections(2):    {{2, 3, 0, 2}},
		newMsIntersections(3):    {{1, 3, 2, 3}},
		newMsIntersections(0, 1): {{0, 2, 1, 3}},
		newMsIntersections(0, 2): {{2, 3, 0, 1}},
		// Resolve ambiguities and don't create holes
		// by covering both cases explicitly.
		newMsIntersections(0, 3): {{0, 2, 0, 1}, {1, 3, 2, 3}},
		newMsIntersections(1, 2): {{0, 1, 1, 3}, {2, 3, 0, 2}},
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

// msLookupTable creates a lookup table which maps cases
// to four-character ASCII line-drawings, in row-major
// order.
func msLookupTableASCII() [16]string {
	mapping := map[msIntersections]string{
		newMsIntersections():     "    ",
		newMsIntersections(0):    "\\   ",
		newMsIntersections(1):    " /  ",
		newMsIntersections(2):    "  / ",
		newMsIntersections(3):    "   \\",
		newMsIntersections(0, 1): "  __",
		newMsIntersections(0, 2): " | |",
		// Resolve ambiguities and don't create holes
		// by covering both cases explicitly.
		newMsIntersections(0, 3): "\\  \\",
		newMsIntersections(1, 2): " // ",
	}

	// Add inverses to complete the table.
	for ints, chars := range mapping {
		invInts := 0xf ^ ints
		if _, ok := mapping[invInts]; !ok {
			mapping[invInts] = chars
		}
	}
	if len(mapping) != 16 {
		panic("incorrect number of cases")
	}

	res := [16]string{}
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

// msBlock is used to track a rectangular area of maching
// squares cells.
type msBlock struct {
	spacer *squareSpacer
	min    [2]int
	max    [2]int
}

func newMsBlock(spacer *squareSpacer) msBlock {
	return msBlock{
		spacer: spacer,
		min:    [2]int{0, 0},
		max:    [2]int{len(spacer.Xs) - 1, len(spacer.Ys) - 1},
	}
}

func (m *msBlock) Bounds(epsilon float64) *Rect {
	min := XY(m.spacer.Xs[m.min[0]], m.spacer.Ys[m.min[1]])
	max := XY(m.spacer.Xs[m.max[0]], m.spacer.Ys[m.max[1]])
	return NewRect(min.AddScalar(-epsilon), max.AddScalar(epsilon))
}

func (m *msBlock) Area() int {
	res := 1
	for i, min := range m.min {
		res *= m.max[i] - min
	}
	return res
}

func (m *msBlock) NumPoints() int {
	res := 1
	for i, min := range m.min {
		res *= (m.max[i] - min) + 1
	}
	return res
}

func (m *msBlock) Pieces(minArea int, g func(*msBlock) bool, f func(*msBlock)) {
	if !g(m) {
		return
	}
	if m.Area()/2 < minArea {
		f(m)
	} else {
		b1, b2 := m.Split()
		b1.Pieces(minArea, g, f)
		b2.Pieces(minArea, g, f)
	}
}

func (m *msBlock) Split() (msBlock, msBlock) {
	lenX := m.max[0] - m.min[0]
	lenY := m.max[1] - m.min[1]
	splitAxis := 0
	if lenY >= lenX {
		splitAxis = 1
	}

	min1 := m.min
	max1 := m.max
	min2 := m.min
	max2 := m.max
	max1[splitAxis] = (m.max[splitAxis] + m.min[splitAxis]) / 2
	min2[splitAxis] = max1[splitAxis]

	s1 := msBlock{
		spacer: m.spacer,
		min:    min1,
		max:    max1,
	}
	s2 := msBlock{
		spacer: m.spacer,
		min:    min2,
		max:    max2,
	}
	return s1, s2
}

type msBlockCache struct {
	block  *msBlock
	values []bool
}

func newMsBlockCache() *msBlockCache {
	return &msBlockCache{values: []bool{}}
}

func (m *msBlockCache) Populate(block *msBlock, s Solid) {
	m.block = block
	m.values = m.values[:0]
	for yIdx := block.min[1]; yIdx <= block.max[1]; yIdx++ {
		edgeY := yIdx == 0 || yIdx == len(m.block.spacer.Ys)-1
		y := block.spacer.Ys[yIdx]
		for xIdx := block.min[0]; xIdx <= block.max[0]; xIdx++ {
			edgeX := xIdx == 0 || xIdx == len(m.block.spacer.Xs)-1
			x := block.spacer.Xs[xIdx]
			value := s.Contains(XY(x, y))
			if value && (edgeX || edgeY) {
				panic("solid is true outside of bounds")
			}
			m.values = append(m.values, value)
		}
	}
}

func (m *msBlockCache) GetSquare(x, y int) msIntersections {
	x -= m.block.min[0]
	y -= m.block.min[1]
	return m.getSegment(x, y) | (m.getSegment(x, y+1) << 2)
}

func (m *msBlockCache) getSegment(x, y int) msIntersections {
	yStride := (1 + m.block.max[0] - m.block.min[0])
	offset := yStride*y + x
	result := msIntersections(0)
	if m.values[offset] {
		result |= 1
	}
	if m.values[offset+1] {
		result |= 2
	}
	return result
}
