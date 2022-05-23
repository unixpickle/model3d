package model3d

import (
	"math"
	"runtime"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/numerical"
)

// DualContouring is a configurable but simplified version
// of Dual Contouring, a technique for turning a field into
// a mesh.
type DualContouring struct {
	// S specifies the Solid and is used to compute hermite
	// data on line segments.
	S SolidSurfaceEstimator

	// Delta specifies the grid size of the algorithm.
	Delta float64

	// NoJitter, if true, disables a small jitter applied to
	// coordinates. This jitter is enabled by default to
	// avoid common error cases when attempting to estimate
	// normals exactly on the edges of boxy objects.
	NoJitter bool

	// MaxGos, if specified, limits the number of Goroutines
	// for parallel processing. If 0, GOMAXPROCS is used.
	MaxGos int
}

// Mesh computes a mesh for the surface.
func (d *DualContouring) Mesh() *Mesh {
	if !BoundsValid(d.S.Solid) {
		panic("invalid bounds for solid")
	}
	s := d.S.Solid
	layout := newDcCubeLayout(s.Min(), s.Max(), d.Delta, d.NoJitter)
	if len(layout.Zs) < 3 {
		panic("invalid number of z values")
	}

	numWorkers := d.MaxGos
	if numWorkers == 0 {
		numWorkers = runtime.GOMAXPROCS(0)
	}
	heap := &dcHeap{}
	activeCubes := dcScanSolid(numWorkers, layout, heap, d.S.Solid)

	activeEdges := map[dcEdgeIdx]struct{}{}
	for _, cube := range activeCubes {
		for _, edge := range heap.Cube(cube).Edges {
			if heap.Edge(edge).Active(heap) {
				activeEdges[edge] = struct{}{}
			}
		}
	}
	activeEdgeSlice := make([]dcEdgeIdx, 0, len(activeEdges))
	for edge := range activeEdges {
		activeEdgeSlice = append(activeEdgeSlice, edge)
	}

	essentials.ConcurrentMap(d.MaxGos, len(activeEdgeSlice), func(i int) {
		d.computeHermite(heap, activeEdgeSlice[i])
	})
	essentials.ConcurrentMap(d.MaxGos, len(activeCubes), func(i int) {
		d.computeVertexPosition(heap, activeCubes[i])
	})

	mesh := NewMesh()
	for edgeIdx := range activeEdges {
		edge := heap.Edge(edgeIdx)

		var cubes [4]*dcCube
		for i, cubeIdx := range edge.Cubes {
			if cubeIdx < 0 {
				panic("missing cube for edge; this likely means the Solid is true outside of bounds")
			}
			cubes[i] = heap.Cube(cubeIdx)
		}
		vs := [4]Coord3D{
			cubes[0].VertexPosition,
			cubes[1].VertexPosition,
			cubes[2].VertexPosition,
			cubes[3].VertexPosition,
		}
		t1, t2 := &Triangle{vs[0], vs[1], vs[2]}, &Triangle{vs[1], vs[3], vs[2]}
		if t1.Normal().Dot(edge.Normal) < 0 {
			t1[0], t1[1] = t1[1], t1[0]
			t2[0], t2[1] = t2[1], t2[0]
		}
		mesh.Add(t1)
		mesh.Add(t2)
	}

	return mesh
}

func (d *DualContouring) computeHermite(heap *dcHeap, edgeIdx dcEdgeIdx) {
	edge := heap.Edge(edgeIdx)
	edge.Coord = d.S.Bisect(heap.Corner(edge.Corners[0]).Coord, heap.Corner(edge.Corners[1]).Coord)
	edge.Normal = d.S.Normal(edge.Coord)
}

func (d *DualContouring) computeVertexPosition(heap *dcHeap, cubeIdx dcCubeIdx) {
	cube := heap.Cube(cubeIdx)

	// Based on Dual Contouring: "The Secret Sauce"
	// https://people.eecs.berkeley.edu/~jrs/meshpapers/SchaeferWarren2.pdf
	var massPoint Coord3D
	var count float64
	var active [12]*dcEdge
	for i, edgeIdx := range cube.Edges {
		if edgeIdx < 0 {
			panic("edge not available for active cube; this likely means the Solid is true outside of bounds")
		}
		edge := heap.Edge(edgeIdx)
		if edge.Active(heap) {
			active[i] = edge
			massPoint = massPoint.Add(edge.Coord)
			count++
		}
	}
	massPoint = massPoint.Scale(1 / count)

	var matA []numerical.Vec3
	var matB []float64
	for _, edge := range active {
		if edge != nil {
			v := edge.Coord.Sub(massPoint)
			matA = append(matA, edge.Normal.Array())
			matB = append(matB, v.Dot(edge.Normal))
		}
	}
	solution := numerical.LeastSquares3(matA, matB, 0.1)
	minPoint := heap.Corner(cube.Corners[0]).Coord
	maxPoint := heap.Corner(cube.Corners[7]).Coord
	cube.VertexPosition = NewCoord3DArray(solution).Add(massPoint).Max(minPoint).Min(maxPoint)
}

type dcCornerRequest struct {
	Index   int
	Coords  []Coord3D
	Corners []dcCornerIdx
	Cubes   []dcCubeIdx
	Values  []bool
}

func dcScanSolid(maxProcs int, layout *dcCubeLayout, heap *dcHeap, s Solid) []dcCubeIdx {
	maxProcs = essentials.MinInt(maxProcs, len(layout.Zs))

	requests := make(chan *dcCornerRequest, maxProcs)
	results := make(chan *dcCornerRequest, maxProcs)
	defer close(requests)
	for i := 0; i < maxProcs; i++ {
		go func() {
			for req := range requests {
				req.Values = make([]bool, len(req.Coords))
				for i, c := range req.Coords {
					req.Values[i] = s.Contains(c)
				}
				results <- req
			}
		}()
	}

	completed := map[int]struct{}{}
	pending := map[int]*dcCornerRequest{}
	var firstIncomplete int
	var firstUnqueued int
	var activeCubes []dcCubeIdx

	handleResponse := func(resp *dcCornerRequest) {
		for i, c := range resp.Corners {
			heap.Corner(c).Value = resp.Values[i]
		}
		completed[resp.Index] = struct{}{}
		idx := resp.Index
		for idx == firstIncomplete {
			if idx > 0 {
				for _, cube := range pending[idx-1].Cubes {
					if heap.Cube(cube).Active(heap) {
						activeCubes = append(activeCubes, cube)
					} else {
						heap.UnlinkCube(cube)
					}
				}
				delete(pending, idx-1)
			}
			firstIncomplete++
			// If rows were completed out of order, we can now
			// process all of the results.
			if _, ok := completed[firstIncomplete]; ok {
				idx++
			}
		}
	}

	previous, corners := layout.FirstRow(heap)
	req := &dcCornerRequest{
		Index:   0,
		Coords:  heap.CornerCoords(corners),
		Corners: corners,
		Cubes:   previous,
	}
	pending[0] = req
	requests <- req
	firstUnqueued = 1

	for firstUnqueued < len(layout.Zs)-1 {
		if firstUnqueued-firstIncomplete < maxProcs {
			previous, corners = layout.NextRow(heap, previous, firstUnqueued-1)
			req = &dcCornerRequest{
				Index:   firstUnqueued,
				Coords:  heap.CornerCoords(corners),
				Corners: corners,
				Cubes:   previous,
			}
			pending[firstUnqueued] = req
			requests <- req
			firstUnqueued++
		} else {
			handleResponse(<-results)
		}
	}
	for firstIncomplete < len(layout.Zs)-1 {
		handleResponse(<-results)
	}

	return activeCubes
}

type dcCubeIdx int
type dcCornerIdx int
type dcEdgeIdx int

type dcHeap struct {
	freeCubes   []int
	freeCorners []int
	freeEdges   []int

	cubes   []dcCube
	corners []dcCorner
	edges   []dcEdge
}

func (d *dcHeap) UnlinkCube(idx dcCubeIdx) {
	for _, eIdx := range d.cubes[idx].Edges {
		e := &d.edges[eIdx]
		var remainingCubes int
		for i, c := range e.Cubes {
			if c == idx {
				e.Cubes[i] = -1
			} else if c >= 0 {
				remainingCubes++
			}
		}
		if remainingCubes == 0 {
			d.freeEdges = append(d.freeEdges, int(eIdx))
		}
	}
	for _, cIdx := range d.cubes[idx].Corners {
		c := &d.corners[cIdx]
		c.Refs--
		if c.Refs == 0 {
			d.freeCorners = append(d.freeCorners, int(cIdx))
		}
	}
	d.freeCubes = append(d.freeCubes, int(idx))
}

func (d *dcHeap) AllocateCube(value dcCube) dcCubeIdx {
	idx := d.allocate(&d.freeCubes, func() int {
		d.cubes = append(d.cubes, dcCube{})
		return len(d.cubes) - 1
	})
	d.cubes[idx] = value
	return dcCubeIdx(idx)
}

func (d *dcHeap) AllocateEdge(value dcEdge) dcEdgeIdx {
	idx := d.allocate(&d.freeEdges, func() int {
		d.edges = append(d.edges, dcEdge{})
		return len(d.edges) - 1
	})
	d.edges[idx] = value
	return dcEdgeIdx(idx)
}

func (d *dcHeap) AllocateCorner(value dcCorner) dcCornerIdx {
	idx := d.allocate(&d.freeCorners, func() int {
		d.corners = append(d.corners, dcCorner{})
		return len(d.corners) - 1
	})
	d.corners[idx] = value
	return dcCornerIdx(idx)
}

func (d *dcHeap) Cube(idx dcCubeIdx) *dcCube {
	return &d.cubes[idx]
}

func (d *dcHeap) Edge(idx dcEdgeIdx) *dcEdge {
	return &d.edges[idx]
}

func (d *dcHeap) Corner(idx dcCornerIdx) *dcCorner {
	return &d.corners[idx]
}

func (d *dcHeap) CornerCoords(corners []dcCornerIdx) []Coord3D {
	res := make([]Coord3D, len(corners))
	for i, c := range corners {
		res[i] = d.corners[c].Coord
	}
	return res
}

func (d *dcHeap) allocate(free *[]int, create func() int) int {
	if len(*free) > 0 {
		idx := (*free)[len(*free)-1]
		*free = (*free)[:len(*free)-1]
		return idx
	} else {
		return create()
	}
}

// a dcCube represents a cube in Dual Contouring.
//
// Each cube has 8 corners, laid out like so:
//
// 0 --------- 1
// |\          |\
// | \         | \
// |  2 --------- 3
// 4 -|------ 5   |
//  \ |        \  |
//   \|         \ |
//    6 --------- 7
//
// Where 0 is at (0, 0, 0), 1 is at (1, 0, 0), 2 is at
// (0, 1, 0), and 4 is at (0, 0, 1) in terms of XYZ.
//
// Each cube has 12 edges, laid out like so:
//
// +-----0-----+
// |\          |\
// | 1         | 2
// 4  \        5  \
// |   +-----3-+---+
// +---+--8----+   |
//  \  6        \  7
//   9 |        10 |
//    \|          \|
//     +---- 11----+
//
// Note that, under the above setup, the last 4 edges are
// equal to the top 4 edges moved down one cube.
// This makes it easier to generate new rows of cubes.
type dcCube struct {
	Corners [8]dcCornerIdx
	Edges   [12]dcEdgeIdx

	VertexPosition Coord3D
}

// Active returns true if the corners do not all share the
// same solid value.
func (d *dcCube) Active(heap *dcHeap) bool {
	first := heap.Corner(d.Corners[0]).Value
	for _, x := range d.Corners[1:] {
		if heap.Corner(x).Value != first {
			return true
		}
	}
	return false
}

type dcCorner struct {
	Value bool
	Coord Coord3D

	// Refs counts the number of allocated cubes this
	// corner still belongs to.
	Refs int
}

type dcEdge struct {
	Corners [2]dcCornerIdx
	Cubes   [4]dcCubeIdx

	// Hermite data
	Coord  Coord3D
	Normal Coord3D
}

// Active returns true if the solid value changes across
// the edge.
func (d *dcEdge) Active(heap *dcHeap) bool {
	return heap.Corner(d.Corners[0]).Value != heap.Corner(d.Corners[1]).Value
}

// dcCubeLayout helps locate cubes in space and arrange
// their vertices and edges.
type dcCubeLayout struct {
	// Xs, Ys, and Zs are the x/y/z-axis values of vertices
	// spaced along the grid.
	// The size of each slice is one greater than th number
	// of cubes along that axis.
	Xs []float64
	Ys []float64
	Zs []float64
}

func newDcCubeLayout(min, max Coord3D, delta float64, noJitter bool) *dcCubeLayout {
	jitter := delta * 0.012923982
	if noJitter {
		jitter = 0
	}

	min = min.AddScalar(-delta)
	max = max.AddScalar(delta)
	count := max.Sub(min).Scale(1 / delta)
	countX := int(math.Round(count.X)) + 1
	countY := int(math.Round(count.Y)) + 1
	countZ := int(math.Round(count.Z)) + 1

	res := &dcCubeLayout{
		Xs: make([]float64, countX),
		Ys: make([]float64, countY),
		Zs: make([]float64, countZ),
	}
	for i := 0; i < countX; i++ {
		res.Xs[i] = min.X + float64(i)*delta + jitter
	}
	for i := 0; i < countY; i++ {
		res.Ys[i] = min.Y + float64(i)*delta + jitter
	}
	for i := 0; i < countZ; i++ {
		res.Zs[i] = min.Z + float64(i)*delta + jitter
	}
	return res
}

// FirstRow creates a row of cubes, returning the cubes and
// all of the created corners.
func (d *dcCubeLayout) FirstRow(heap *dcHeap) (cubes []dcCubeIdx, corners []dcCornerIdx) {
	cubes = make([]dcCubeIdx, (len(d.Xs)-1)*(len(d.Ys)-1))
	corners = make([]dcCornerIdx, len(d.Xs)*len(d.Ys)*2)
	edgeX := make([]dcEdgeIdx, 0, (len(d.Xs)-1)*len(d.Ys)*2)
	edgeY := make([]dcEdgeIdx, 0, (len(d.Ys)-1)*len(d.Xs)*2)
	edgeZ := make([]dcEdgeIdx, 0, len(d.Xs)*len(d.Ys))

	cornerAt := func(x, y, z int) dcCornerIdx {
		return corners[x+y*len(d.Xs)+z*(len(corners)/2)]
	}
	cubeAt := func(x, y, z int) dcCubeIdx {
		if z != 0 || x < 0 || y < 0 || x >= len(d.Xs)-1 || y >= len(d.Ys)-1 {
			return -1
		}
		return cubes[x+y*(len(d.Xs)-1)]
	}

	for i := range cubes {
		cubes[i] = heap.AllocateCube(dcCube{})
	}
	for i := range corners {
		corners[i] = heap.AllocateCorner(dcCorner{})
	}

	// Create corners and z edges.
	for i, y := range d.Ys {
		for j, x := range d.Xs {
			idx1 := i*len(d.Xs) + j
			idx2 := i*len(d.Xs) + j + len(corners)/2
			heap.Corner(corners[idx1]).Coord = XYZ(x, y, d.Zs[0])
			heap.Corner(corners[idx2]).Coord = XYZ(x, y, d.Zs[1])
			edgeZ = append(edgeZ, heap.AllocateEdge(dcEdge{
				Corners: [2]dcCornerIdx{corners[idx1], corners[idx2]},
				Cubes: [4]dcCubeIdx{
					cubeAt(j-1, i-1, 0),
					cubeAt(j, i-1, 0),
					cubeAt(j-1, i, 0),
					cubeAt(j, i, 0),
				},
			}))
		}
	}

	// Create x and y edges and point them to corners and cubes.
	for z := 0; z < 2; z++ {
		for y := 0; y < len(d.Ys); y++ {
			for x := 0; x < len(d.Xs)-1; x++ {
				edgeX = append(edgeX, heap.AllocateEdge(dcEdge{
					Corners: [2]dcCornerIdx{
						cornerAt(x, y, z),
						cornerAt(x+1, y, z),
					},
					Cubes: [4]dcCubeIdx{
						cubeAt(x, y-1, z-1),
						cubeAt(x, y, z-1),
						cubeAt(x, y-1, z),
						cubeAt(x, y, z),
					},
				}))
			}
		}
		for y := 0; y < len(d.Ys)-1; y++ {
			for x := 0; x < len(d.Xs); x++ {
				edgeY = append(edgeY, heap.AllocateEdge(dcEdge{
					Corners: [2]dcCornerIdx{
						cornerAt(x, y, z),
						cornerAt(x, y+1, z),
					},
					Cubes: [4]dcCubeIdx{
						cubeAt(x-1, y, z-1),
						cubeAt(x, y, z-1),
						cubeAt(x-1, y, z),
						cubeAt(x, y, z),
					},
				}))
			}
		}
	}

	// Link cubes to corners and edges.
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			cube := heap.Cube(cubeAt(x, y, 0))
			dcCornerIdx := 0
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					for k := 0; k < 2; k++ {
						c := cornerAt(x+k, y+j, i)
						cube.Corners[dcCornerIdx] = c
						heap.Corner(c).Refs++
						dcCornerIdx++
					}
				}
			}
			cube.Edges = [12]dcEdgeIdx{
				// Top edges.
				edgeX[x+y*(len(d.Xs)-1)],
				edgeY[x+y*len(d.Xs)],
				edgeY[1+x+y*len(d.Xs)],
				edgeX[x+(y+1)*(len(d.Xs)-1)],
				// Vertical edges.
				edgeZ[x+y*len(d.Xs)],
				edgeZ[1+x+y*len(d.Xs)],
				edgeZ[x+(y+1)*len(d.Xs)],
				edgeZ[1+x+(y+1)*len(d.Xs)],
				// Bottom edges.
				edgeX[len(edgeX)/2+x+y*(len(d.Xs)-1)],
				edgeY[len(edgeY)/2+x+y*len(d.Xs)],
				edgeY[len(edgeY)/2+1+x+y*len(d.Xs)],
				edgeX[len(edgeX)/2+x+(y+1)*(len(d.Xs)-1)],
			}
		}
	}
	return cubes, corners
}

// NextRow generates a new row of cubes below the previous
// row. For the first call, prevIdx should be 0, and for
// subsequent calls it should increase by 1 each time.
//
// The previous cubes will be linked to the new cubes.
//
// Returns the newly created cubes and newly created row of
// corners.
func (d *dcCubeLayout) NextRow(heap *dcHeap, previous []dcCubeIdx,
	prevIdx int) (cubes []dcCubeIdx, corners []dcCornerIdx) {
	cubes = make([]dcCubeIdx, (len(d.Xs)-1)*(len(d.Ys)-1))
	corners = make([]dcCornerIdx, len(d.Xs)*len(d.Ys))
	edgeX := make([]dcEdgeIdx, 0, (len(d.Xs)-1)*len(d.Ys))
	edgeY := make([]dcEdgeIdx, 0, (len(d.Ys)-1)*len(d.Xs))
	edgeZ := make([]dcEdgeIdx, 0, len(d.Xs)*len(d.Ys))

	for i := range cubes {
		cubes[i] = heap.AllocateCube(dcCube{})
	}

	// Reconstruct previous corners in the correct order.
	previousCorners := make([]dcCornerIdx, len(d.Xs)*len(d.Ys))
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			cube := heap.Cube(previous[x+y*(len(d.Xs)-1)])
			previousCorners[x+y*(len(d.Xs))] = cube.Corners[4]
			if x+1 == len(d.Xs)-1 || y+1 == len(d.Ys)-1 {
				previousCorners[1+x+y*(len(d.Xs))] = cube.Corners[5]
				previousCorners[x+(y+1)*(len(d.Xs))] = cube.Corners[6]
				previousCorners[1+x+(y+1)*(len(d.Xs))] = cube.Corners[7]
			}
		}
	}

	cornerAt := func(x, y, z int) dcCornerIdx {
		if z < 0 || z > 1 {
			panic("z out of bounds")
		}
		idx := x + y*len(d.Xs)
		if z == 0 {
			return previousCorners[idx]
		} else {
			return corners[idx]
		}
	}
	cubeAt := func(x, y, z int) dcCubeIdx {
		if z > 0 || x < 0 || y < 0 || x >= len(d.Xs)-1 || y >= len(d.Ys)-1 {
			return -1
		}
		idx := x + y*(len(d.Xs)-1)
		if z == -1 {
			return previous[idx]
		} else if z == 0 {
			return cubes[idx]
		}
		return -1
	}

	// Create corners and z edges.
	for i, y := range d.Ys {
		for j, x := range d.Xs {
			idx := i*len(d.Xs) + j
			corners[idx] = heap.AllocateCorner(dcCorner{Coord: XYZ(x, y, d.Zs[prevIdx+2])})
			edgeZ = append(edgeZ, heap.AllocateEdge(dcEdge{
				Corners: [2]dcCornerIdx{previousCorners[idx], corners[idx]},
				Cubes: [4]dcCubeIdx{
					cubeAt(j-1, i-1, 0),
					cubeAt(j, i-1, 0),
					cubeAt(j-1, i, 0),
					cubeAt(j, i, 0),
				},
			}))
		}
	}

	// Create bottom x and y edges and point them to corners and
	// cubes.
	for y := 0; y < len(d.Ys); y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			edgeX = append(edgeX, heap.AllocateEdge(dcEdge{
				Corners: [2]dcCornerIdx{
					cornerAt(x, y, 1),
					cornerAt(x+1, y, 1),
				},
				Cubes: [4]dcCubeIdx{
					cubeAt(x, y-1, 0),
					cubeAt(x, y, 0),
					cubeAt(x, y-1, 1),
					cubeAt(x, y, 1),
				},
			}))
		}
	}
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs); x++ {
			edgeY = append(edgeY, heap.AllocateEdge(dcEdge{
				Corners: [2]dcCornerIdx{
					cornerAt(x, y, 1),
					cornerAt(x, y+1, 1),
				},
				Cubes: [4]dcCubeIdx{
					cubeAt(x-1, y, 0),
					cubeAt(x, y, 0),
					cubeAt(x-1, y, 1),
					cubeAt(x, y, 1),
				},
			}))
		}
	}

	// Link cubes to corners and edges.
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			above := heap.Cube(cubeAt(x, y, -1))
			// Link above cube's bottom edges to the new row of cubes.
			heap.Edge(above.Edges[8]).Cubes[2] = cubeAt(x, y-1, 0)
			heap.Edge(above.Edges[8]).Cubes[3] = cubeAt(x, y, 0)
			heap.Edge(above.Edges[9]).Cubes[2] = cubeAt(x-1, y, 0)
			heap.Edge(above.Edges[9]).Cubes[3] = cubeAt(x, y, 0)
			heap.Edge(above.Edges[10]).Cubes[2] = cubeAt(x, y, 0)
			heap.Edge(above.Edges[10]).Cubes[3] = cubeAt(x+1, y, 0)
			heap.Edge(above.Edges[11]).Cubes[2] = cubeAt(x, y, 0)
			heap.Edge(above.Edges[11]).Cubes[3] = cubeAt(x, y+1, 0)

			// Link new cube to its edges and corners.
			cube := heap.Cube(cubeAt(x, y, 0))
			dcCornerIdx := 0
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					for k := 0; k < 2; k++ {
						c := cornerAt(x+k, y+j, i)
						cube.Corners[dcCornerIdx] = c
						heap.Corner(c).Refs++
						dcCornerIdx++
					}
				}
			}
			cube.Edges = [12]dcEdgeIdx{
				// Top edges.
				above.Edges[8],
				above.Edges[9],
				above.Edges[10],
				above.Edges[11],
				// Vertical edges.
				edgeZ[x+y*len(d.Xs)],
				edgeZ[1+x+y*len(d.Xs)],
				edgeZ[x+(y+1)*len(d.Xs)],
				edgeZ[1+x+(y+1)*len(d.Xs)],
				// Bottom edges.
				edgeX[x+y*(len(d.Xs)-1)],
				edgeY[x+y*len(d.Xs)],
				edgeY[1+x+y*len(d.Xs)],
				edgeX[x+(y+1)*(len(d.Xs)-1)],
			}
		}
	}

	return cubes, corners
}
