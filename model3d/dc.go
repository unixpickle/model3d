package model3d

import (
	"math"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/numerical"
)

const DefaultDualContouringBufferSize = 1000000

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

	// BufferSize is a soft-limit on the number of cached
	// vertices that are stored in memory at once.
	// Defaults to DefaultDualContouringBufferSize.
	BufferSize int
}

// Mesh computes a mesh for the surface.
func (d *DualContouring) Mesh() *Mesh {
	if !BoundsValid(d.S.Solid) {
		panic("invalid bounds for solid")
	}
	s := d.S.Solid
	layout := newDcCubeLayout(s.Min(), s.Max(), d.Delta, d.NoJitter, d.BufferSize)
	if len(layout.Zs) < 3 {
		panic("invalid number of z values")
	}

	populateCorners := func() {
		essentials.ConcurrentMap(d.MaxGos, len(layout.Corners), func(i int) {
			corner := layout.Corner(dcCornerIdx(i))
			if !corner.Populated {
				corner.Populated = true
				corner.Value = d.S.Solid.Contains(corner.Coord)
			}
		})
	}

	populateEdges := func() {
		essentials.ConcurrentMap(d.MaxGos, len(layout.Edges), func(i int) {
			edge := layout.Edge(dcEdgeIdx(i))
			if !edge.Populated {
				edge.Populated = true
				corners := layout.EdgeCorners(dcEdgeIdx(i))
				c1 := layout.Corner(corners[0])
				c2 := layout.Corner(corners[1])
				edge.Active = (c1.Value != c2.Value)
				if edge.Active {
					edge.Coord = d.S.Bisect(c1.Coord, c2.Coord)
					edge.Normal = d.S.Normal(edge.Coord)
				}
			}
		})
	}

	populateCubes := func() {
		essentials.ConcurrentMap(d.MaxGos, len(layout.Cubes), func(i int) {
			cube := layout.Cube(dcCubeIdx(i))
			if cube.Populated {
				return
			}
			cube.Populated = true
			if !layout.CubeActive(dcCubeIdx(i)) {
				return
			}
			var massPoint Coord3D
			var count float64
			var active [12]*dcEdge
			for i, edgeIdx := range layout.CubeEdges(dcCubeIdx(i)) {
				if edgeIdx < 0 {
					panic("edge not available for active cube; this likely means the Solid is true outside of bounds")
				}
				edge := layout.Edge(edgeIdx)
				if edge.Active {
					active[i] = edge
					massPoint = massPoint.Add(edge.Coord)
					count++
				}
			}
			if count == 0 {
				panic("no acive edges found")
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
			p := NewCoord3DArray(solution).Add(massPoint)
			minPoint, maxPoint := layout.CubeMinMax(dcCubeIdx(i))
			cube.VertexPosition = p.Max(minPoint).Min(maxPoint)
		})
	}

	mesh := NewMesh()
	appendMesh := func() {
		layout.UsableEdges(func(i dcEdgeIdx) {
			e := layout.Edge(i)
			if e.Triangulated || !e.Active {
				return
			}
			e.Triangulated = true
			var vs [4]Coord3D
			for i, c := range layout.EdgeCubes(i) {
				if c < 0 {
					panic("solid is true outside of bounds")
				}
				vs[i] = layout.Cube(c).VertexPosition
			}
			t1, t2 := &Triangle{vs[0], vs[1], vs[2]}, &Triangle{vs[1], vs[3], vs[2]}
			if t1.Normal().Dot(e.Normal) < 0 {
				t1[0], t1[1] = t1[1], t1[0]
				t2[0], t2[1] = t2[1], t2[0]
			}
			mesh.Add(t1)
			mesh.Add(t2)
		})
	}

	for {
		populateCorners()
		populateEdges()
		populateCubes()
		appendMesh()
		if layout.Remaining() == 0 {
			break
		}
		layout.Shift()
	}

	return mesh
}

type dcCubeIdx int
type dcCornerIdx int
type dcEdgeIdx int

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
	Populated      bool
	VertexPosition Coord3D
}

type dcCorner struct {
	Populated bool
	Value     bool
	Coord     Coord3D
}

type dcEdge struct {
	Populated    bool
	Active       bool
	Triangulated bool
	// Hermite data
	Coord  Coord3D
	Normal Coord3D
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

	// Cache of a subset of the total volume, sequential in
	// the z-axis.
	ZOffset int
	BufRows int
	Corners []dcCorner
	Cubes   []dcCube
	// Edges are stored in the following order, where X/Y/Z
	// represents the axis that a group of edges span:
	// XYZXYZXYZ...XYZXY
	Edges []dcEdge
}

func newDcCubeLayout(min, max Coord3D, delta float64, noJitter bool, bufSize int) *dcCubeLayout {
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

	if bufSize == 0 {
		bufSize = DefaultDualContouringBufferSize
	}
	bufRows := bufSize / (len(res.Xs) * len(res.Ys))
	bufRows = essentials.MinInt(essentials.MaxInt(bufRows, 4), len(res.Zs))
	res.BufRows = bufRows
	res.Corners = make([]dcCorner, 0, len(res.Xs)*len(res.Ys)*bufRows)
	res.Cubes = make([]dcCube, (len(res.Xs)-1)*(len(res.Ys)-1)*(bufRows-1))

	xCount := (len(res.Xs) - 1) * len(res.Ys)
	yCount := (len(res.Ys) - 1) * len(res.Xs)
	zCount := len(res.Xs) * len(res.Ys)
	res.Edges = make([]dcEdge, (xCount+yCount)*bufRows+zCount*(bufRows-1))

	for i := 0; i < bufRows; i++ {
		for j := 0; j < len(res.Ys); j++ {
			for k := 0; k < len(res.Xs); k++ {
				res.Corners = append(res.Corners, dcCorner{
					Coord: XYZ(res.Xs[k], res.Ys[j], res.Zs[i]),
				})
			}
		}
	}

	return res
}

func (d *dcCubeLayout) Remaining() int {
	return len(d.Zs) - (d.BufRows + d.ZOffset)
}

// Shift moves up all of the buffer by the number of rows.
func (d *dcCubeLayout) Shift() {
	rows := essentials.MinInt(d.Remaining(), d.BufRows-2)

	cubeRowSize := (len(d.Xs) - 1) * (len(d.Ys) - 1)
	cornerRowSize := len(d.Xs) * len(d.Ys)
	edgeRowSize := (len(d.Xs)-1)*len(d.Ys) + (len(d.Ys)-1)*len(d.Xs) + cornerRowSize

	copy(d.Cubes, d.Cubes[rows*cubeRowSize:])
	copy(d.Corners, d.Corners[rows*cornerRowSize:])
	copy(d.Edges, d.Edges[rows*edgeRowSize:])

	d.ZOffset += rows

	for i := len(d.Cubes) - rows*cubeRowSize; i < len(d.Cubes); i++ {
		d.Cubes[i] = dcCube{}
	}
	for i := len(d.Edges) - rows*edgeRowSize; i < len(d.Edges); i++ {
		d.Edges[i] = dcEdge{}
	}
	for i := len(d.Corners) - rows*cornerRowSize; i < len(d.Corners); i++ {
		x := i % len(d.Xs)
		y := (i / len(d.Xs)) % len(d.Ys)
		z := (i / len(d.Xs)) / len(d.Ys)
		d.Corners[i] = dcCorner{
			Coord: XYZ(
				d.Xs[x],
				d.Ys[y],
				d.Zs[z+d.ZOffset],
			),
		}
	}
}

func (d *dcCubeLayout) Cube(c dcCubeIdx) *dcCube {
	return &d.Cubes[int(c)]
}

func (d *dcCubeLayout) Corner(c dcCornerIdx) *dcCorner {
	return &d.Corners[int(c)]
}

func (d *dcCubeLayout) Edge(e dcEdgeIdx) *dcEdge {
	return &d.Edges[int(e)]
}

func (d *dcCubeLayout) UsableEdges(f func(dcEdgeIdx)) {
	atBottom := d.ZOffset+d.BufRows == len(d.Zs)
	xCount, yCount, _ := d.edgeCounts()
	endIdx := len(d.Edges)
	if !atBottom {
		endIdx -= xCount + yCount
	}
	for i := 0; i < endIdx; i++ {
		f(dcEdgeIdx(i))
	}
}

func (d *dcCubeLayout) CubeActive(c dcCubeIdx) bool {
	var value, result bool
	for i, idx := range d.CubeCorners(c) {
		thisValue := d.Corner(idx).Value
		if i == 0 {
			value = thisValue
		} else if value != thisValue {
			result = true
		}
	}
	return result
}

func (d *dcCubeLayout) CubeMinMax(c dcCubeIdx) (min, max Coord3D) {
	for i, idx := range d.CubeCorners(c) {
		coord := d.Corner(idx).Coord
		if i == 0 {
			min, max = coord, coord
		} else {
			min = min.Min(coord)
			max = max.Max(coord)
		}
	}
	return
}

func (d *dcCubeLayout) CubeCorners(c dcCubeIdx) [8]dcCornerIdx {
	x, y, z := d.cubeCoord(c)
	var result [8]dcCornerIdx
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				cornerIdx := (x + k) + ((y+j)+(z+i)*len(d.Ys))*len(d.Xs)
				result[k+j*2+i*4] = dcCornerIdx(cornerIdx)
			}
		}
	}
	return result
}

func (d *dcCubeLayout) CubeEdges(c dcCubeIdx) [12]dcEdgeIdx {
	x, y, z := d.cubeCoord(c)

	return [12]dcEdgeIdx{
		d.xEdgeIdx(x, y, z),
		d.yEdgeIdx(x, y, z),
		d.yEdgeIdx(x+1, y, z),
		d.xEdgeIdx(x, y+1, z),
		d.zEdgeIdx(x, y, z),
		d.zEdgeIdx(x+1, y, z),
		d.zEdgeIdx(x, y+1, z),
		d.zEdgeIdx(x+1, y+1, z),
		d.xEdgeIdx(x, y, z+1),
		d.yEdgeIdx(x, y, z+1),
		d.yEdgeIdx(x+1, y, z+1),
		d.xEdgeIdx(x, y+1, z+1),
	}
}

func (d *dcCubeLayout) EdgeCorners(e dcEdgeIdx) [2]dcCornerIdx {
	xCount, yCount, zCount := d.edgeCounts()

	edgeIdx := int(e)
	z := edgeIdx / (xCount + yCount + zCount)
	edgeIdx = edgeIdx % (xCount + yCount + zCount)
	if edgeIdx < xCount {
		x := edgeIdx % (len(d.Xs) - 1)
		y := edgeIdx / (len(d.Xs) - 1)
		return [2]dcCornerIdx{d.cornerIdx(x, y, z), d.cornerIdx(x+1, y, z)}
	} else if edgeIdx < xCount+yCount {
		edgeIdx -= xCount
		x := edgeIdx % len(d.Xs)
		y := edgeIdx / len(d.Xs)
		return [2]dcCornerIdx{d.cornerIdx(x, y, z), d.cornerIdx(x, y+1, z)}
	} else {
		edgeIdx -= xCount + yCount
		x := edgeIdx % len(d.Xs)
		y := edgeIdx / len(d.Xs)
		return [2]dcCornerIdx{d.cornerIdx(x, y, z), d.cornerIdx(x, y, z+1)}
	}
}

func (d *dcCubeLayout) EdgeCubes(e dcEdgeIdx) [4]dcCubeIdx {
	xCount, yCount, zCount := d.edgeCounts()

	cubeAt := func(x, y, z int) dcCubeIdx {
		if z < 0 || x < 0 || y < 0 || x >= len(d.Xs)-1 || y >= len(d.Ys)-1 || z >= d.BufRows-1 {
			return -1
		}
		return dcCubeIdx(x + (y+z*(len(d.Ys)-1))*(len(d.Xs)-1))
	}

	edgeIdx := int(e)
	z := edgeIdx / (xCount + yCount + zCount)

	edgeIdx = edgeIdx % (xCount + yCount + zCount)
	if edgeIdx < xCount {
		x := edgeIdx % (len(d.Xs) - 1)
		y := edgeIdx / (len(d.Xs) - 1)
		return [4]dcCubeIdx{
			cubeAt(x, y-1, z-1),
			cubeAt(x, y, z-1),
			cubeAt(x, y-1, z),
			cubeAt(x, y, z),
		}
	} else if edgeIdx < xCount+yCount {
		edgeIdx -= xCount
		x := edgeIdx % len(d.Xs)
		y := edgeIdx / len(d.Xs)
		return [4]dcCubeIdx{
			cubeAt(x-1, y, z-1),
			cubeAt(x, y, z-1),
			cubeAt(x-1, y, z),
			cubeAt(x, y, z),
		}
	} else {
		edgeIdx -= xCount + yCount
		x := edgeIdx % len(d.Xs)
		y := edgeIdx / len(d.Xs)
		return [4]dcCubeIdx{
			cubeAt(x-1, y-1, z),
			cubeAt(x, y-1, z),
			cubeAt(x-1, y, z),
			cubeAt(x, y, z),
		}
	}
}

func (d *dcCubeLayout) cubeCoord(c dcCubeIdx) (x, y, z int) {
	cubeIdx := int(c)
	x = cubeIdx % (len(d.Xs) - 1)
	cubeIdx /= (len(d.Xs) - 1)
	y = cubeIdx % (len(d.Ys) - 1)
	cubeIdx /= (len(d.Ys) - 1)
	z = cubeIdx
	return
}

func (d *dcCubeLayout) edgeCounts() (xCount, yCount, zCount int) {
	xCount = (len(d.Xs) - 1) * len(d.Ys)
	yCount = (len(d.Ys) - 1) * len(d.Xs)
	zCount = len(d.Xs) * len(d.Ys)
	return
}

func (d *dcCubeLayout) cornerIdx(x, y, z int) dcCornerIdx {
	return dcCornerIdx(x + (y+z*len(d.Ys))*len(d.Xs))
}

func (d *dcCubeLayout) xEdgeIdx(x, y, z int) dcEdgeIdx {
	xCount, yCount, zCount := d.edgeCounts()
	return dcEdgeIdx(z*(xCount+yCount+zCount) + (len(d.Xs)-1)*y + x)
}

func (d *dcCubeLayout) yEdgeIdx(x, y, z int) dcEdgeIdx {
	xCount, yCount, zCount := d.edgeCounts()
	return dcEdgeIdx(z*(xCount+yCount+zCount) + xCount + len(d.Xs)*y + x)
}

func (d *dcCubeLayout) zEdgeIdx(x, y, z int) dcEdgeIdx {
	xCount, yCount, zCount := d.edgeCounts()
	return dcEdgeIdx(z*(xCount+yCount+zCount) + xCount + yCount + len(d.Xs)*y + x)
}
