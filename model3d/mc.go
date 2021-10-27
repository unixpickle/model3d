package model3d

import (
	"math"
	"runtime"
	"sort"
	"sync/atomic"

	"github.com/unixpickle/essentials"
)

// MarchingCubes turns a Solid into a surface mesh using a
// corrected marching cubes algorithm.
func MarchingCubes(s Solid, delta float64) *Mesh {
	if !BoundsValid(s) {
		panic("invalid bounds for solid")
	}

	table := mcLookupTable()
	spacer := newSquareSpacer(s, delta)
	mesh := NewMesh()
	spacer.Scan(s, func(z int, bottomCache, topCache *solidCache) {
		for y := 0; y < len(spacer.Ys)-1; y++ {
			for x := 0; x < len(spacer.Xs)-1; x++ {
				bits := bottomCache.GetSquare(x, y) | (topCache.GetSquare(x, y) << 4)
				triangles := table[bits]
				if len(triangles) > 0 {
					min := spacer.CornerCoord(x, y, z-1)
					max := spacer.CornerCoord(x+1, y+1, z)
					corners := mcCornerCoordinates(min, max)
					for _, t := range triangles {
						mesh.Add(t.Triangle(corners))
					}
				}
			}
		}
	})
	return mesh
}

// MarchingCubesSearch is like MarchingCubes, but applies
// an additional search step to move the vertices along
// the edges of each cube.
//
// The tightness of the triangulation will double for
// every iteration.
func MarchingCubesSearch(s Solid, delta float64, iters int) *Mesh {
	mesh := MarchingCubes(s, delta)
	mcSearch(s, delta, iters, mesh)
	return mesh
}

// MarchingCubesConj is like MarchingCubesSearch, but in a
// transformed space. In particular, it applies a series of
// transformations to the Solid, and then applies the
// inverse to the resulting mesh.
func MarchingCubesConj(s Solid, delta float64, iters int, xforms ...Transform) *Mesh {
	joined := JoinedTransform(xforms)
	solid := TransformSolid(joined, s)
	mesh := MarchingCubesSearch(solid, delta, iters)
	return mesh.Transform(joined.Inverse())
}

// MarchingCubesRC is like MarchingCubes, but eliminates
// rectangular volumes not intersecting a RectCollider.
// This can be much more efficient than MarchingCubes if
// the surface is sparse.
//
// Note that the RectCollider can be conservative and
// possibly report collisions that do not occur.
// However, it should never fail to report collisions,
// since this could cause triangles to be missed.
func MarchingCubesRC(s Solid, rc RectCollider, delta float64) *Mesh {
	if !BoundsValid(s) {
		panic("invalid bounds for solid")
	}

	table := mcLookupTable()
	spacer := newSquareSpacer(s, delta)

	numGos := runtime.GOMAXPROCS(0)
	blockQueue := make(chan *mcBlock, numGos)
	outputs := make(chan *Mesh, numGos)

	blockFilter := func(m *mcBlock) bool {
		// We add a small buffer around bounding boxes to prevent
		// rounding errors from missing rect collisions.
		epsilon := delta * 1e-3

		return rc.RectCollision(m.Bounds(epsilon))
	}

	subDivideVolume := 64
	for i := 0; i < numGos; i++ {
		go func() {
			result := NewMesh()
			cache := newMcBlockCache()
			for block := range blockQueue {
				block.Pieces(subDivideVolume, blockFilter, func(block *mcBlock) {
					cache.Populate(block, s)
					for z := block.min[2]; z < block.max[2]; z++ {
						for y := block.min[1]; y < block.max[1]; y++ {
							for x := block.min[0]; x < block.max[0]; x++ {
								bits := cache.GetCube(x, y, z)
								triangles := table[bits]
								if len(triangles) > 0 {
									min := spacer.CornerCoord(x, y, z)
									max := spacer.CornerCoord(x+1, y+1, z+1)
									corners := mcCornerCoordinates(min, max)
									for _, t := range triangles {
										result.Add(t.Triangle(corners))
									}
								}
							}
						}
					}
				})
			}
			outputs <- result
		}()
	}

	rootBlock := newMcBlock(spacer)

	// Divide the volume up into relatively small pieces,
	// but allow some splitting to be done by the individual
	// Goroutines to increase performance.
	divideVolume := essentials.MaxInt(rootBlock.Volume()/4096, subDivideVolume)
	rootBlock.Pieces(divideVolume, blockFilter, func(m *mcBlock) {
		blockQueue <- m
	})
	close(blockQueue)

	mesh := NewMesh()
	for i := 0; i < numGos; i++ {
		mesh.AddMesh(<-outputs)
	}
	return mesh
}

// MarchingCubesSearchRC combines MarchingCubesSearch with
// MarchingCubesRC.
func MarchingCubesSearchRC(s Solid, rc RectCollider, delta float64, iters int) *Mesh {
	mesh := MarchingCubesRC(s, rc, delta)
	mcSearch(s, delta, iters, mesh)
	return mesh
}

func mcSearch(s Solid, delta float64, iters int, mesh *Mesh) {
	if iters == 0 {
		return
	}

	inVertices := mesh.VertexSlice()
	outVertices := make([]Coord3D, len(inVertices))

	min := s.Min().Array()
	essentials.ConcurrentMap(0, len(inVertices), func(i int) {
		outVertices[i] = mcSearchPoint(s, delta, iters, mesh, min, inVertices[i])
	})

	v2t := mesh.getVertexToFace()
	for i, original := range inVertices {
		out := outVertices[i]
		for _, t := range v2t.Value(original) {
			for j, c := range t {
				if c == original {
					t[j] = out
					break
				}
			}
		}
	}

	// We just invalidated the entire v2t cache by
	// replacing the vertices in the triangles.
	mesh.vertexToFace = atomic.Value{}
}

func mcSearchPoint(s Solid, delta float64, iters int, m *Mesh, min [3]float64, c Coord3D) Coord3D {
	arr := c.Array()

	// Figure out which axis the containing edge spans.
	axis := -1
	var falsePoint, truePoint float64
	for i := 0; i < 3; i++ {
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
	if m.Find(c)[0].Normal().Array()[axis] > 0 {
		truePoint, falsePoint = falsePoint, truePoint
	}

	for i := 0; i < iters; i++ {
		midPoint := (falsePoint + truePoint) / 2
		arr[axis] = midPoint
		if s.Contains(NewCoord3DArray(arr)) {
			truePoint = midPoint
		} else {
			falsePoint = midPoint
		}
	}

	arr[axis] = (falsePoint + truePoint) / 2
	return NewCoord3DArray(arr)
}

// mcCorner is a corner index on a cube used for marching
// cubes.
//
// Ordered as:
//
//     (0, 0, 0), (1, 0, 0), (0, 1, 0), (1, 1, 0),
//     (0, 0, 1), (1, 0, 1), (0, 1, 1), (1, 1, 1)
//
// Here is a visualization of the cube indices:
//
//         6 + -----------------------+ 7
//          /|                       /|
//         / |                      / |
//        /  |                     /  |
//     4 +------------------------+ 5 |
//       |   |                    |   |
//       |   |                    |   |
//       |   |                    |   |
//       |   | 2                  |   | 3
//       |   +--------------------|---+
//       |  /                     |  /
//       | /                      | /
//       |/                       |/
//       +------------------------+
//      0                           1
//
type mcCorner uint8

// mcCornerCoordinates gets the coordinates of all eight
// corners for a cube.
func mcCornerCoordinates(min, max Coord3D) [8]Coord3D {
	return [8]Coord3D{
		min,
		{X: max.X, Y: min.Y, Z: min.Z},
		{X: min.X, Y: max.Y, Z: min.Z},
		{X: max.X, Y: max.Y, Z: min.Z},

		{X: min.X, Y: min.Y, Z: max.Z},
		{X: max.X, Y: min.Y, Z: max.Z},
		{X: min.X, Y: max.Y, Z: max.Z},
		max,
	}
}

// mcRotation represents a cube rotation for marching
// cubes.
//
// For corner c, rotation[c] is the new corner at that
// location.
type mcRotation [8]mcCorner

// allMcRotations gets all 24 possible rotations for a
// cube in marching cubes.
func allMcRotations() []mcRotation {
	// Create a generating basis.
	zRotation := mcRotation{2, 0, 3, 1, 6, 4, 7, 5}
	xRotation := mcRotation{2, 3, 6, 7, 0, 1, 4, 5}

	queue := []mcRotation{{0, 1, 2, 3, 4, 5, 6, 7}}
	resMap := map[mcRotation]bool{queue[0]: true}
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		resMap[next] = true
		for _, op := range []mcRotation{zRotation, xRotation} {
			rotated := op.Compose(next)
			if !resMap[rotated] {
				resMap[rotated] = true
				queue = append(queue, rotated)
			}
		}
	}

	var result []mcRotation
	for rotation := range resMap {
		result = append(result, rotation)
	}

	// Make the rotation order deterministic and fairly
	// sensible.
	sort.Slice(result, func(i, j int) bool {
		r1 := result[i]
		r2 := result[j]
		for k := range r1 {
			if r1[k] < r2[k] {
				return true
			} else if r1[k] > r2[k] {
				return false
			}
		}
		return false
	})

	return result
}

// Compose combines two rotations.
func (m mcRotation) Compose(m1 mcRotation) mcRotation {
	var res mcRotation
	for i := range res {
		res[i] = m[m1[i]]
	}
	return res
}

// ApplyCorner applies the rotation to a corner.
func (m mcRotation) ApplyCorner(c mcCorner) mcCorner {
	return m[c]
}

// ApplyTriangle applies the rotation to a triangle.
func (m mcRotation) ApplyTriangle(t mcTriangle) mcTriangle {
	var res mcTriangle
	for i, c := range t {
		res[i] = m.ApplyCorner(c)
	}
	return res
}

// ApplyIntersections applies the rotation to an
// mcIntersections.
func (m mcRotation) ApplyIntersections(i mcIntersections) mcIntersections {
	var res mcIntersections
	for c := mcCorner(0); c < 8; c++ {
		if i.Inside(c) {
			res |= 1 << m.ApplyCorner(c)
		}
	}
	return res
}

// mcTriangle is a triangle constructed out of midpoints
// of edges of a cube.
// There are 6 corners because each pair of two represents
// an edge.
//
// The triangle is ordered in counter-clockwise order when
// looked upon from the outside.
type mcTriangle [6]mcCorner

// Triangle creates a real triangle out of the mcTriangle,
// given the corner coordinates.
func (m mcTriangle) Triangle(corners [8]Coord3D) *Triangle {
	return &Triangle{
		corners[m[0]].Mid(corners[m[1]]),
		corners[m[2]].Mid(corners[m[3]]),
		corners[m[4]].Mid(corners[m[5]]),
	}
}

// mcIntersections represents which corners on a cube are
// inside of a solid.
// Each corner is a bit, and 1 means inside.
type mcIntersections uint8

// newMcIntersections creates an mcIntersections using the
// corners that are inside the solid.
func newMcIntersections(trueCorners ...mcCorner) mcIntersections {
	if len(trueCorners) > 8 {
		panic("expected at most 8 corners")
	}
	var res mcIntersections
	for _, c := range trueCorners {
		res |= (1 << c)
	}
	return res
}

// Inside checks if a corner c is true.
func (m mcIntersections) Inside(c mcCorner) bool {
	return (m & (1 << c)) != 0
}

// mcLookupTable creates a full lookup table that maps
// each mcIntersection to a set of triangles.
func mcLookupTable() [256][]mcTriangle {
	rotations := allMcRotations()
	result := map[mcIntersections][]mcTriangle{}

	for baseInts, baseTris := range baseTriangleTable {
		for _, rot := range rotations {
			newInts := rot.ApplyIntersections(baseInts)
			if _, ok := result[newInts]; !ok {
				newTris := []mcTriangle{}
				for _, t := range baseTris {
					newTris = append(newTris, rot.ApplyTriangle(t))
				}
				result[newInts] = newTris
			}
		}
	}

	var resultArray [256][]mcTriangle
	for key, value := range result {
		resultArray[key] = value
	}
	return resultArray
}

// baseTriangleTable encodes the marching cubes lookup
// table (up to rotations) from:
// "A survey of the marching cubes algorithm" (2006).
// https://cg.informatik.uni-freiburg.de/intern/seminar/surfaceReconstruction_survey%20of%20marching%20cubes.pdf
var baseTriangleTable = map[mcIntersections][]mcTriangle{
	// Case 0-5
	newMcIntersections(): {},
	newMcIntersections(0): {
		{0, 1, 0, 2, 0, 4},
	},
	newMcIntersections(0, 1): {
		{0, 4, 1, 5, 0, 2},
		{1, 5, 1, 3, 0, 2},
	},
	newMcIntersections(0, 5): {
		{0, 1, 0, 2, 0, 4},
		{5, 7, 1, 5, 4, 5},
	},
	newMcIntersections(0, 7): {
		{0, 1, 0, 2, 0, 4},
		{6, 7, 3, 7, 5, 7},
	},
	newMcIntersections(1, 2, 3): {
		{0, 1, 1, 5, 0, 2},
		{0, 2, 1, 5, 2, 6},
		{2, 6, 1, 5, 3, 7},
	},

	// Case 6-11
	newMcIntersections(0, 1, 7): {
		// Case 2.
		{0, 4, 1, 5, 0, 2},
		{1, 5, 1, 3, 0, 2},
		// End of case 4
		{6, 7, 3, 7, 5, 7},
	},
	newMcIntersections(1, 4, 7): {
		{4, 6, 4, 5, 0, 4},
		{1, 5, 1, 3, 0, 1},
		// End of case 4.
		{6, 7, 3, 7, 5, 7},
	},
	newMcIntersections(0, 1, 2, 3): {
		{0, 4, 1, 5, 3, 7},
		{0, 4, 3, 7, 2, 6},
	},
	newMcIntersections(0, 2, 3, 6): {
		{0, 1, 4, 6, 0, 4},
		{0, 1, 6, 7, 4, 6},
		{0, 1, 1, 3, 6, 7},
		{1, 3, 3, 7, 6, 7},
	},
	newMcIntersections(1, 2, 5, 6): {
		{0, 2, 2, 3, 6, 7},
		{0, 2, 6, 7, 4, 6},
		{0, 1, 4, 5, 5, 7},
		{5, 7, 1, 3, 0, 1},
	},
	newMcIntersections(0, 2, 3, 7): {
		{0, 4, 0, 1, 2, 6},
		{0, 1, 5, 7, 2, 6},
		{2, 6, 5, 7, 6, 7},
		{0, 1, 1, 3, 5, 7},
	},

	// Case 12-17
	newMcIntersections(1, 2, 3, 4): {
		{0, 1, 1, 5, 0, 2},
		{0, 2, 1, 5, 2, 6},
		{2, 6, 1, 5, 3, 7},
		{4, 5, 0, 4, 4, 6},
	},
	newMcIntersections(1, 2, 4, 7): {
		{0, 1, 1, 5, 1, 3},
		{0, 2, 2, 3, 2, 6},
		{4, 5, 0, 4, 4, 6},
		{5, 7, 6, 7, 3, 7},
	},
	newMcIntersections(1, 2, 3, 6): {
		{0, 2, 0, 1, 4, 6},
		{0, 1, 3, 7, 4, 6},
		{0, 1, 1, 5, 3, 7},
		{4, 6, 3, 7, 6, 7},
	},
	newMcIntersections(0, 2, 3, 5, 6): {
		// Case 9
		{0, 1, 4, 6, 0, 4},
		{0, 1, 6, 7, 4, 6},
		{0, 1, 1, 3, 6, 7},
		{1, 3, 3, 7, 6, 7},
		// End of case 3
		{5, 7, 1, 5, 4, 5},
	},
	newMcIntersections(2, 3, 4, 5, 6): {
		{5, 7, 1, 5, 0, 4},
		{0, 4, 6, 7, 5, 7},
		{0, 2, 6, 7, 0, 4},
		{0, 2, 3, 7, 6, 7},
		{0, 2, 1, 3, 3, 7},
	},
	newMcIntersections(0, 4, 5, 6, 7): {
		{1, 5, 0, 1, 0, 2},
		{0, 2, 2, 6, 1, 5},
		{1, 5, 2, 6, 3, 7},
	},

	// Case 18-22
	newMcIntersections(1, 2, 3, 4, 5, 6): {
		// Inverse of case 4.
		{0, 2, 0, 1, 0, 4},
		{3, 7, 6, 7, 5, 7},
	},
	newMcIntersections(1, 2, 3, 4, 6, 7): {
		{0, 2, 4, 5, 0, 4},
		{0, 2, 5, 7, 4, 5},
		{0, 2, 1, 5, 5, 7},
		{0, 1, 1, 5, 0, 2},
	},
	newMcIntersections(2, 3, 4, 5, 6, 7): {
		// Inverse of case 2.
		{1, 5, 0, 4, 0, 2},
		{1, 3, 1, 5, 0, 2},
	},
	newMcIntersections(1, 2, 3, 4, 5, 6, 7): {
		// Inverse of case 1.
		{0, 2, 0, 1, 0, 4},
	},
	newMcIntersections(0, 1, 2, 3, 4, 5, 6, 7): {},
}

type squareSpacer struct {
	Xs []float64
	Ys []float64
	Zs []float64
}

func newSquareSpacer(s Solid, delta float64) *squareSpacer {
	var xs, ys, zs []float64
	min := s.Min()
	max := s.Max()
	for x := min.X - delta; x <= max.X+delta; x += delta {
		xs = append(xs, x)
	}
	for y := min.Y - delta; y <= max.Y+delta; y += delta {
		ys = append(ys, y)
	}
	for z := min.Z - delta; z <= max.Z+delta; z += delta {
		zs = append(zs, z)
	}
	return &squareSpacer{Xs: xs, Ys: ys, Zs: zs}
}

func (s *squareSpacer) CornerCoord(x, y, z int) Coord3D {
	return XYZ(s.Xs[x], s.Ys[y], s.Zs[z])
}

func (s *squareSpacer) Scan(solid Solid, f func(z int, bottom, top *solidCache)) {
	numGos := runtime.GOMAXPROCS(0)

	// Prevent edge case where we are making a very
	// flat object on a multi-core machine.
	if numGos > len(s.Zs)-1 {
		numGos = len(s.Zs) - 1
	}

	caches := make([]*asyncSolidCache, numGos+1)
	for i := range caches {
		caches[i] = &asyncSolidCache{
			Cache: newSolidCache(solid, s),
			Done:  make(chan struct{}, 1),
		}
		caches[i].FetchZ(i)
	}

	<-caches[0].Done
	for nextZ := 1; nextZ < len(s.Zs); nextZ++ {
		prevIdx := (nextZ - 1) % len(caches)
		curIdx := nextZ % len(caches)

		<-caches[curIdx].Done

		f(nextZ, caches[prevIdx].Cache, caches[curIdx].Cache)

		if nextZ+len(caches)-1 < len(s.Zs) {
			caches[prevIdx].FetchZ(nextZ + len(caches) - 1)
		}
	}
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
		values: make([]bool, len(spacer.Xs)*len(spacer.Ys)),
	}
}

func (s *solidCache) FetchZ(z int) {
	maxX := len(s.spacer.Xs) - 1
	maxY := len(s.spacer.Ys) - 1
	onEdge := z == 0 || z == len(s.spacer.Zs)-1

	var idx int
	for i := 0; i < len(s.spacer.Ys); i++ {
		for j := 0; j < len(s.spacer.Xs); j++ {
			b := s.solid.Contains(s.spacer.CornerCoord(j, i, z))
			s.values[idx] = b
			idx++
			if b && (onEdge || i == 0 || j == 0 || i == maxY || j == maxX) {
				panic("solid is true outside of bounds")
			}
		}
	}
}

func (s *solidCache) Get(x, y int) bool {
	return s.values[x+y*len(s.spacer.Xs)]
}

func (s *solidCache) GetSquare(x, y int) mcIntersections {
	result := mcIntersections(0)
	mask := mcIntersections(1)
	for y1 := y; y1 < y+2; y1++ {
		for x1 := x; x1 < x+2; x1++ {
			if s.Get(x1, y1) {
				result |= mask
			}
			mask <<= 1
		}
	}
	return result
}

type asyncSolidCache struct {
	Cache *solidCache
	Done  chan struct{}
}

func (a *asyncSolidCache) FetchZ(z int) {
	go func() {
		a.Cache.FetchZ(z)
		a.Done <- struct{}{}
	}()
}

// mcBlock is used to track a rectangular volume of maching
// cubes cells.
type mcBlock struct {
	spacer *squareSpacer
	min    [3]int
	max    [3]int
}

func newMcBlock(spacer *squareSpacer) mcBlock {
	return mcBlock{
		spacer: spacer,
		min:    [3]int{0, 0, 0},
		max:    [3]int{len(spacer.Xs) - 1, len(spacer.Ys) - 1, len(spacer.Zs) - 1},
	}
}

func (m *mcBlock) Bounds(epsilon float64) *Rect {
	min := XYZ(m.spacer.Xs[m.min[0]], m.spacer.Ys[m.min[1]], m.spacer.Zs[m.min[2]])
	max := XYZ(m.spacer.Xs[m.max[0]], m.spacer.Ys[m.max[1]], m.spacer.Zs[m.max[2]])
	return NewRect(min.AddScalar(-epsilon), max.AddScalar(epsilon))
}

func (m *mcBlock) Volume() int {
	res := 1
	for i, min := range m.min {
		res *= m.max[i] - min
	}
	return res
}

func (m *mcBlock) NumPoints() int {
	res := 1
	for i, min := range m.min {
		res *= (m.max[i] - min) + 1
	}
	return res
}

func (m *mcBlock) Pieces(minVolume int, g func(*mcBlock) bool, f func(*mcBlock)) {
	if !g(m) {
		return
	}
	if m.Volume()/2 < minVolume {
		f(m)
	} else {
		b1, b2 := m.Split()
		b1.Pieces(minVolume, g, f)
		b2.Pieces(minVolume, g, f)
	}
}

func (m *mcBlock) Split() (mcBlock, mcBlock) {
	lenX := m.max[0] - m.min[0]
	lenY := m.max[1] - m.min[1]
	lenZ := m.max[2] - m.min[2]
	splitAxis := 0
	if lenY >= lenX && lenY >= lenZ {
		splitAxis = 1
	} else if lenZ >= lenX && lenZ >= lenY {
		splitAxis = 2
	}

	min1 := m.min
	max1 := m.max
	min2 := m.min
	max2 := m.max
	max1[splitAxis] = (m.max[splitAxis] + m.min[splitAxis]) / 2
	min2[splitAxis] = max1[splitAxis]

	s1 := mcBlock{
		spacer: m.spacer,
		min:    min1,
		max:    max1,
	}
	s2 := mcBlock{
		spacer: m.spacer,
		min:    min2,
		max:    max2,
	}
	return s1, s2
}

type mcBlockCache struct {
	block  *mcBlock
	values []bool
}

func newMcBlockCache() *mcBlockCache {
	return &mcBlockCache{values: []bool{}}
}

func (m *mcBlockCache) Populate(block *mcBlock, s Solid) {
	m.block = block
	m.values = m.values[:0]
	for zIdx := block.min[2]; zIdx <= block.max[2]; zIdx++ {
		edgeZ := zIdx == 0 || zIdx == len(m.block.spacer.Zs)-1
		z := block.spacer.Zs[zIdx]
		for yIdx := block.min[1]; yIdx <= block.max[1]; yIdx++ {
			edgeY := yIdx == 0 || yIdx == len(m.block.spacer.Ys)-1
			y := block.spacer.Ys[yIdx]
			for xIdx := block.min[0]; xIdx <= block.max[0]; xIdx++ {
				edgeX := xIdx == 0 || xIdx == len(m.block.spacer.Xs)-1
				x := block.spacer.Xs[xIdx]
				value := s.Contains(XYZ(x, y, z))
				if value && (edgeX || edgeY || edgeZ) {
					panic("solid is true outside of bounds")
				}
				m.values = append(m.values, value)
			}
		}
	}
}

func (m *mcBlockCache) GetCube(x, y, z int) mcIntersections {
	x -= m.block.min[0]
	y -= m.block.min[1]
	z -= m.block.min[2]
	return m.getSquare(x, y, z) | (m.getSquare(x, y, z+1) << 4)
}

func (m *mcBlockCache) getSquare(x, y, z int) mcIntersections {
	yStride := (1 + m.block.max[0] - m.block.min[0])
	zStride := yStride * (1 + m.block.max[1] - m.block.min[1])
	result := mcIntersections(0)
	mask := mcIntersections(1)
	for y1 := y; y1 < y+2; y1++ {
		for x1 := x; x1 < x+2; x1++ {
			if m.values[z*zStride+y1*yStride+x1] {
				result |= mask
			}
			mask <<= 1
		}
	}
	return result
}
