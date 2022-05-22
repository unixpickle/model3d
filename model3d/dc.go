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
	s := d.S.Solid
	layout := newDcCubeLayout(s.Min(), s.Max(), d.Delta, d.NoJitter)
	if len(layout.Zs) < 3 {
		panic("invalid number of z values")
	}

	numWorkers := d.MaxGos
	if numWorkers == 0 {
		numWorkers = runtime.GOMAXPROCS(0)
	}
	activeCubes := dcScanSolid(numWorkers, layout, d.S.Solid)

	activeEdges := map[*dcEdge]struct{}{}
	for _, cube := range activeCubes {
		for _, edge := range cube.Edges {
			if edge.Active() {
				activeEdges[edge] = struct{}{}
			}
		}
	}
	activeEdgeSlice := make([]*dcEdge, 0, len(activeEdges))
	for edge := range activeEdges {
		activeEdgeSlice = append(activeEdgeSlice, edge)
	}

	essentials.ConcurrentMap(d.MaxGos, len(activeEdgeSlice), func(i int) {
		d.computeHermite(activeEdgeSlice[i])
	})
	essentials.ConcurrentMap(d.MaxGos, len(activeCubes), func(i int) {
		d.computeVertexPosition(activeCubes[i])
	})

	mesh := NewMesh()
	for edge := range activeEdges {
		if edge.Cubes[0] == nil || edge.Cubes[1] == nil || edge.Cubes[2] == nil ||
			edge.Cubes[3] == nil {
			panic("missing cube for edge; this likely means the Solid is true outside of bounds")
		}
		vs := [4]Coord3D{
			edge.Cubes[0].VertexPosition,
			edge.Cubes[1].VertexPosition,
			edge.Cubes[2].VertexPosition,
			edge.Cubes[3].VertexPosition,
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

func (d *DualContouring) computeHermite(edge *dcEdge) {
	edge.Coord = d.S.Bisect(edge.Corners[0].Coord, edge.Corners[1].Coord)
	edge.Normal = d.S.Normal(edge.Coord)
}

func (d *DualContouring) computeVertexPosition(cube *dcCube) {
	// Based on Dual Contouring: "The Secret Sauce"
	// https://people.eecs.berkeley.edu/~jrs/meshpapers/SchaeferWarren2.pdf
	var massPoint Coord3D
	var count float64
	for _, edge := range cube.Edges {
		if edge.Active() {
			massPoint = massPoint.Add(edge.Coord)
			count++
		}
	}
	massPoint = massPoint.Scale(1 / count)

	var matA []numerical.Vec3
	var matB []float64
	for _, edge := range cube.Edges {
		if edge.Active() {
			v := edge.Coord.Sub(massPoint)
			matA = append(matA, edge.Normal.Array())
			matB = append(matB, v.Dot(edge.Normal))
		}
	}
	solution := numerical.LeastSquares3(matA, matB, 0.1)
	minPoint := cube.Corners[0].Coord
	maxPoint := cube.Corners[7].Coord
	cube.VertexPosition = NewCoord3DArray(solution).Add(massPoint).Max(minPoint).Min(maxPoint)
}

type dcCornerRequest struct {
	Index   int
	Corners []*dcCorner
}

func dcScanSolid(maxProcs int, layout *dcCubeLayout, s Solid) []*dcCube {
	maxProcs = essentials.MinInt(maxProcs, len(layout.Zs))

	requests := make(chan *dcCornerRequest, maxProcs)
	results := make(chan int, maxProcs)
	defer close(requests)
	for i := 0; i < maxProcs; i++ {
		go func() {
			for req := range requests {
				for _, corner := range req.Corners {
					corner.Value = s.Contains(corner.Coord)
				}
				results <- req.Index
			}
		}()
	}

	pending := map[int][]*dcCube{}
	completed := map[int]struct{}{}
	var firstIncomplete int
	var firstUnqueued int
	var activeCubes []*dcCube

	var handleResponse func(idx int)
	handleResponse = func(idx int) {
		completed[idx] = struct{}{}
		for idx == firstIncomplete {
			if idx > 0 {
				for _, cube := range pending[idx-1] {
					if cube.Active() {
						activeCubes = append(activeCubes, cube)
					} else {
						cube.Unlink()
					}
				}
				pending[idx-1] = nil
			}
			firstIncomplete++
			// If rows were completed out of order, we can now
			// process all of the results.
			if _, ok := completed[firstIncomplete]; ok {
				handleResponse(idx + 1)
			}
		}
	}

	previous, corners := layout.FirstRow()
	pending[0] = previous

	requests <- &dcCornerRequest{Corners: corners, Index: 0}
	firstUnqueued = 1

	for firstUnqueued < len(layout.Zs)-1 {
		if firstUnqueued-firstIncomplete < maxProcs {
			previous, corners = layout.NextRow(previous, firstUnqueued-1)
			pending[firstUnqueued] = previous
			requests <- &dcCornerRequest{Corners: corners, Index: firstUnqueued}
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
	Corners [8]*dcCorner
	Edges   [12]*dcEdge

	VertexPosition Coord3D
}

// Active returns true if the corners do not all share the
// same solid value.
func (d *dcCube) Active() bool {
	first := d.Corners[0].Value
	for _, x := range d.Corners[1:] {
		if x.Value != first {
			return true
		}
	}
	return false
}

// Unlink removes this cube's references to its edges and
// corners and removes the edges' references to itself.
func (d *dcCube) Unlink() {
	for _, e := range d.Edges {
		for i, c := range e.Cubes {
			if c == d {
				e.Cubes[i] = nil
				break
			}
		}
	}
	d.Edges = [12]*dcEdge{}
	d.Corners = [8]*dcCorner{}
}

type dcCorner struct {
	Value bool
	Coord Coord3D
}

type dcEdge struct {
	Corners [2]*dcCorner
	Cubes   [4]*dcCube

	// Hermite data
	Coord  Coord3D
	Normal Coord3D
}

// Active returns true if the solid value changes across
// the edge.
func (d *dcEdge) Active() bool {
	return d.Corners[0].Value != d.Corners[1].Value
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
func (d *dcCubeLayout) FirstRow() ([]*dcCube, []*dcCorner) {
	cubes := make([]*dcCube, (len(d.Xs)-1)*(len(d.Ys)-1))
	corners := make([]*dcCorner, len(d.Xs)*len(d.Ys)*2)
	edgeX := make([]*dcEdge, 0, (len(d.Xs)-1)*len(d.Ys)*2)
	edgeY := make([]*dcEdge, 0, (len(d.Ys)-1)*len(d.Xs)*2)
	edgeZ := make([]*dcEdge, 0, len(d.Xs)*len(d.Ys))

	cornerAt := func(x, y, z int) *dcCorner {
		return corners[x+y*len(d.Xs)+z*(len(corners)/2)]
	}
	cubeAt := func(x, y, z int) *dcCube {
		if z != 0 || x < 0 || y < 0 || x >= len(d.Xs)-1 || y >= len(d.Ys)-1 {
			return nil
		}
		return cubes[x+y*(len(d.Xs)-1)]
	}

	for i := range cubes {
		cubes[i] = &dcCube{}
	}

	// Create corners and z edges.
	for i, y := range d.Ys {
		for j, x := range d.Xs {
			idx1 := i*len(d.Xs) + j
			idx2 := i*len(d.Xs) + j + len(corners)/2
			corners[idx1] = &dcCorner{Coord: XYZ(x, y, d.Zs[0])}
			corners[idx2] = &dcCorner{Coord: XYZ(x, y, d.Zs[1])}
			edgeZ = append(edgeZ, &dcEdge{
				Corners: [2]*dcCorner{corners[idx1], corners[idx2]},
				Cubes: [4]*dcCube{
					cubeAt(j-1, i-1, 0),
					cubeAt(j, i-1, 0),
					cubeAt(j-1, i, 0),
					cubeAt(j, i, 0),
				},
			})
		}
	}

	// Create x and y edges and point them to corners and cubes.
	for z := 0; z < 2; z++ {
		for y := 0; y < len(d.Ys); y++ {
			for x := 0; x < len(d.Xs)-1; x++ {
				edgeX = append(edgeX, &dcEdge{
					Corners: [2]*dcCorner{
						cornerAt(x, y, z),
						cornerAt(x+1, y, z),
					},
					Cubes: [4]*dcCube{
						cubeAt(x, y-1, z-1),
						cubeAt(x, y, z-1),
						cubeAt(x, y-1, z),
						cubeAt(x, y, z),
					},
				})
			}
		}
		for y := 0; y < len(d.Ys)-1; y++ {
			for x := 0; x < len(d.Xs); x++ {
				edgeY = append(edgeY, &dcEdge{
					Corners: [2]*dcCorner{
						cornerAt(x, y, z),
						cornerAt(x, y+1, z),
					},
					Cubes: [4]*dcCube{
						cubeAt(x-1, y, z-1),
						cubeAt(x, y, z-1),
						cubeAt(x-1, y, z),
						cubeAt(x, y, z),
					},
				})
			}
		}
	}

	// Link cubes to corners and edges.
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			cube := cubeAt(x, y, 0)
			cornerIdx := 0
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					for k := 0; k < 2; k++ {
						cube.Corners[cornerIdx] = cornerAt(x+k, y+j, i)
						cornerIdx++
					}
				}
			}
			cube.Edges = [12]*dcEdge{
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
func (d *dcCubeLayout) NextRow(previous []*dcCube, prevIdx int) ([]*dcCube, []*dcCorner) {
	cubes := make([]*dcCube, (len(d.Xs)-1)*(len(d.Ys)-1))
	corners := make([]*dcCorner, len(d.Xs)*len(d.Ys))
	edgeX := make([]*dcEdge, 0, (len(d.Xs)-1)*len(d.Ys))
	edgeY := make([]*dcEdge, 0, (len(d.Ys)-1)*len(d.Xs))
	edgeZ := make([]*dcEdge, 0, len(d.Xs)*len(d.Ys))

	// Reconstruct previous corners in the correct order.
	previousCorners := make([]*dcCorner, len(d.Xs)*len(d.Ys))
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			cube := previous[x+y*(len(d.Xs)-1)]
			previousCorners[x+y*(len(d.Xs))] = cube.Corners[4]
			if x+1 == len(d.Xs)-1 || y+1 == len(d.Ys)-1 {
				previousCorners[1+x+y*(len(d.Xs))] = cube.Corners[5]
				previousCorners[x+(y+1)*(len(d.Xs))] = cube.Corners[6]
				previousCorners[1+x+(y+1)*(len(d.Xs))] = cube.Corners[7]
			}
		}
	}

	cornerAt := func(x, y, z int) *dcCorner {
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
	cubeAt := func(x, y, z int) *dcCube {
		if z > 0 || x < 0 || y < 0 || x >= len(d.Xs)-1 || y >= len(d.Ys)-1 {
			return nil
		}
		idx := x + y*(len(d.Xs)-1)
		if z == -1 {
			return previous[idx]
		} else if z == 0 {
			return cubes[idx]
		}
		return nil
	}

	for i := range cubes {
		cubes[i] = &dcCube{}
	}

	// Create corners and z edges.
	for i, y := range d.Ys {
		for j, x := range d.Xs {
			idx := i*len(d.Xs) + j
			corners[idx] = &dcCorner{Coord: XYZ(x, y, d.Zs[prevIdx+2])}
			edgeZ = append(edgeZ, &dcEdge{
				Corners: [2]*dcCorner{previousCorners[idx], corners[idx]},
				Cubes: [4]*dcCube{
					cubeAt(j-1, i-1, 0),
					cubeAt(j, i-1, 0),
					cubeAt(j-1, i, 0),
					cubeAt(j, i, 0),
				},
			})
		}
	}

	// Create bottom x and y edges and point them to corners and
	// cubes.
	for y := 0; y < len(d.Ys); y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			edgeX = append(edgeX, &dcEdge{
				Corners: [2]*dcCorner{
					cornerAt(x, y, 1),
					cornerAt(x+1, y, 1),
				},
				Cubes: [4]*dcCube{
					cubeAt(x, y-1, 0),
					cubeAt(x, y, 0),
					cubeAt(x, y-1, 1),
					cubeAt(x, y, 1),
				},
			})
		}
	}
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs); x++ {
			edgeY = append(edgeY, &dcEdge{
				Corners: [2]*dcCorner{
					cornerAt(x, y, 1),
					cornerAt(x, y+1, 1),
				},
				Cubes: [4]*dcCube{
					cubeAt(x-1, y, 0),
					cubeAt(x, y, 0),
					cubeAt(x-1, y, 1),
					cubeAt(x, y, 1),
				},
			})
		}
	}

	// Link cubes to corners and edges.
	for y := 0; y < len(d.Ys)-1; y++ {
		for x := 0; x < len(d.Xs)-1; x++ {
			above := cubeAt(x, y, -1)
			// Link above cube's bottom edges to the new row of cubes.
			above.Edges[8].Cubes[2] = cubeAt(x, y-1, 0)
			above.Edges[8].Cubes[3] = cubeAt(x, y, 0)
			above.Edges[9].Cubes[2] = cubeAt(x-1, y, 0)
			above.Edges[9].Cubes[3] = cubeAt(x, y, 0)
			above.Edges[10].Cubes[2] = cubeAt(x, y, 0)
			above.Edges[10].Cubes[3] = cubeAt(x+1, y, 0)
			above.Edges[11].Cubes[2] = cubeAt(x, y, 0)
			above.Edges[11].Cubes[3] = cubeAt(x, y+1, 0)

			// Link new cube to its edges and corners.
			cube := cubeAt(x, y, 0)
			cornerIdx := 0
			for i := 0; i < 2; i++ {
				for j := 0; j < 2; j++ {
					for k := 0; k < 2; k++ {
						cube.Corners[cornerIdx] = cornerAt(x+k, y+j, i)
						cornerIdx++
					}
				}
			}
			cube.Edges = [12]*dcEdge{
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
