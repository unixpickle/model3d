package model3d

import (
	"math"
	"sort"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/numerical"
)

const (
	DefaultDualContouringBufferSize           = 1000000
	DefaultDualContouringRepairEpsilon        = 0.01
	DefaultDualContouringCubeMargin           = 0.001
	DefaultDualContouringSingularValueEpsilon = 0.1
)

type DualContouringTriangleMode int

const (
	DualContouringTriangleModeMaxMinArea DualContouringTriangleMode = iota
	DualContouringTriangleModeSharpest
	DualContouringTriangleModeFlattest
)

// DualContour is a shortcut for creating a DualContouring
// instance and calling Mesh() on it.
func DualContour(s Solid, delta float64, repair, clip bool) *Mesh {
	dc := &DualContouring{
		S: SolidSurfaceEstimator{
			Solid: s,
		},
		Delta:  delta,
		Repair: repair,
		Clip:   clip,
	}
	return dc.Mesh()
}

// DualContourSDF constructs an SDF of a Solid using dual
// contouring. This approach accounts for the fact that
// dual contouring produces self-intersecting meshes and
// will therefore not obey the even-odd rule.
func DualContourSDF(s Solid, delta float64) FaceSDF {
	mesh := DualContour(s, delta, false, false)
	res := MeshToSDF(mesh).(*meshSDF)

	// The mesh might have a larger bounding box than the
	// original solid due to sharp edges.
	res.Solid = ForceSolidBounds(s, s.Min().Min(mesh.Min()), s.Max().Max(mesh.Max()))

	return res
}

// DualContouring is a configurable but simplified version
// of Dual Contouring, a technique for turning a field into
// a mesh.
//
// By default, DualContouring does not attempt to produce
// manifold meshes, since doing so can result in ugly edge
// artifacts, reducing the primary benefit of DC. To attempt
// manifold meshes, set Clip and Repair to true.
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

	// Repair, if true, attempts to make non-manifold
	// meshes manifold. It is not guaranteed to work, even
	// if Clip is true.
	Repair bool

	// Clip, if true, clips vertices to their cubes.
	// The CubeMargin is used to add a small buffer to the
	// edge of the cubes.
	// When paired with Repair, this can helps meshes be
	// manifold, although it still will not guarantee
	// orientability.
	Clip bool

	// CubeMargin is space around the edges of each cube
	// that a vertex is not allowed to fall into. This can
	// prevent various types of non-manifold geometry.
	//
	// This size is relative to Delta.
	//
	// Defaults to DefaultDualContouringCubeMargin.
	// Only is used if CLip is true.
	CubeMargin float64

	// RepairEpsilon is a small value indicating the amount
	// to move vertices when fixing singularities.
	// It will be scaled relative to Delta to prevent large
	// changes relative to the grid size.
	//
	// Defaults to DefaultDualContouringRepairEpsilon.
	// Only is used if Repair is true.
	RepairEpsilon float64

	// SingularValueEpsilon is the smallest singular value
	// to allow for pseudoinverse calculations.
	//
	// Defaults to DefaultDualContouringSingularValueEpsilon.
	SingularValueEpsilon float64

	// L2Penalty is the regularization coefficient imposed
	// on solutions to the QEF.
	L2Penalty float64

	// TriangleMode controls how quads are triangulated.
	TriangleMode DualContouringTriangleMode
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

	mesh := NewMesh()
	for {
		d.populateCorners(layout)
		d.populateEdges(layout)
		d.populateCubes(layout)
		d.appendMesh(layout, mesh)
		if layout.Remaining() == 0 {
			break
		}
		layout.Shift()
	}

	if d.Repair {
		orig := d.repairSingularEdges(mesh, layout)
		d.repairSingularVertices(mesh, layout, orig)
		mesh.clearVertexToFace()
	}

	return mesh
}

func (d *DualContouring) populateCorners(layout *dcCubeLayout) {
	essentials.ConcurrentMap(d.MaxGos, len(layout.Corners), func(i int) {
		corner := layout.Corner(dcCornerIdx(i))
		if !corner.Populated {
			corner.Populated = true
			corner.Value = d.S.Solid.Contains(corner.Coord)
		}
	})
}

func (d *DualContouring) populateEdges(layout *dcCubeLayout) {
	essentials.ConcurrentMap(d.MaxGos, len(layout.Edges), func(i int) {
		edge := layout.Edge(dcEdgeIdx(i))
		if edge.Populated {
			return
		}
		edge.Populated = true
		corners := layout.EdgeCorners(dcEdgeIdx(i))
		c1 := layout.Corner(corners[0])
		c2 := layout.Corner(corners[1])
		edge.Active = (c1.Value != c2.Value)
		if edge.Active {
			edge.Coord = d.S.Bisect(c1.Coord, c2.Coord)
			edge.Normal = d.S.Normal(edge.Coord)
		}
	})
}

func (d *DualContouring) populateCubes(layout *dcCubeLayout) {
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
		solution := numerical.LeastSquaresReg3(matA, matB, d.L2Penalty, d.singularValueEpsilon())
		p := NewCoord3DArray(solution).Add(massPoint)
		if d.Clip {
			minPoint, maxPoint := layout.CubeMinMax(dcCubeIdx(i))
			margin := d.CubeMargin
			if margin == 0 {
				margin = DefaultDualContouringCubeMargin
			}
			margin = margin * d.Delta
			minPoint = minPoint.AddScalar(margin)
			maxPoint = maxPoint.AddScalar(-margin)
			p = p.Max(minPoint).Min(maxPoint)
		}

		cube.VertexPosition = p
	})
}

func (d *DualContouring) appendMesh(layout *dcCubeLayout, mesh *Mesh) {
	numEdges := layout.UsableEdges()
	essentials.ReduceConcurrentMap(d.MaxGos, numEdges, func() (func(i int), func()) {
		subMesh := NewMesh()
		addEdge := func(idx int) {
			i := dcEdgeIdx(idx)
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

			// Flip the normals based on the topology.
			// The return value of EdgeCubes is ordered in a
			// consistent way so that normals can be computed
			// like this.
			cs := layout.EdgeCorners(i)
			if layout.Corner(cs[0]).Value {
				vs[0], vs[1], vs[2], vs[3] = vs[3], vs[2], vs[1], vs[0]
			}

			t1, t2 := d.triangulateQuad(vs)

			subMesh.Add(t1)
			subMesh.Add(t2)
		}
		reduce := func() {
			mesh.AddMesh(subMesh)
		}
		return addEdge, reduce
	})
}

func (d *DualContouring) triangulateQuad(vs [4]Coord3D) (t1, t2 *Triangle) {
	t1a, t2a := &Triangle{vs[0], vs[1], vs[2]}, &Triangle{vs[0], vs[2], vs[3]}
	vs[0], vs[1], vs[2], vs[3] = vs[1], vs[2], vs[3], vs[0]
	t1b, t2b := &Triangle{vs[0], vs[1], vs[2]}, &Triangle{vs[0], vs[2], vs[3]}

	if d.TriangleMode == DualContouringTriangleModeSharpest ||
		d.TriangleMode == DualContouringTriangleModeFlattest {
		dotA := t1a.Normal().Dot(t2a.Normal())
		dotB := t1b.Normal().Dot(t2b.Normal())
		if d.TriangleMode == DualContouringTriangleModeFlattest {
			dotA, dotB = dotB, dotA
		}
		if dotA < dotB {
			t1, t2 = t1a, t2a
		} else {
			t1, t2 = t1b, t2b
		}
	} else if d.TriangleMode == DualContouringTriangleModeMaxMinArea {
		areaA := math.Min(t1a.Area(), t2a.Area())
		areaB := math.Min(t1b.Area(), t2b.Area())
		if areaA > areaB {
			t1, t2 = t1a, t2a
		} else {
			t1, t2 = t1b, t2b
		}
	}

	return
}

func (d *DualContouring) repairSingularEdges(m *Mesh, layout *dcCubeLayout) *CoordMap[bool] {
	groups := singularEdgeGroups(m)
	if len(groups) == 0 {
		origPoints := NewCoordMap[bool]()
		m.IterateVertices(func(c Coord3D) {
			origPoints.Store(c, true)
		})
		return origPoints
	}
	epsilon := d.repairEpsilon() * 0.49

	if d.Clip {
		// Constrain vertices to be within a margin of the cube
		// so that moving/creating vertices will not cause
		// self-intersections.
		mapping := NewCoordMap[Coord3D]()
		for _, group := range groups {
			group.Constrain(m, epsilon, layout).Range(func(k, v Coord3D) bool {
				mapping.Store(k, v)
				return true
			})
		}
		mapInPlace(m, mapping)
		for _, group := range groups {
			group.Map(mapping)
		}
	}
	origPoints := NewCoordMap[bool]()
	m.IterateVertices(func(c Coord3D) {
		origPoints.Store(c, true)
	})
	for _, group := range groups {
		group.Repair(m, epsilon)
	}
	return origPoints
}

func (d *DualContouring) repairSingularVertices(m *Mesh, layout *dcCubeLayout, orig *CoordMap[bool]) {
	groups := singularVertexGroups(m)
	if len(groups) == 0 {
		return
	}
	epsilon := d.repairEpsilon() * 0.49

	if d.Clip {
		// Constrain vertices to be within a margin of the cube
		// so that moving singular vertices will not cause
		// self-intersections.
		//
		// Note that the previous step of repairing singular
		// edges might have caused vertices to become singular,
		// but all of these now-singular vertices were originally
		// generated within a cube. The extra vertices added to
		// the topology by singular edge repair will never be
		// singular themselves.
		mapping := NewCoordMap[Coord3D]()
		for _, group := range groups {
			group.Constrain(m, epsilon, layout, orig).Range(func(k, v Coord3D) bool {
				mapping.Store(k, v)
				return true
			})
		}
		mapInPlace(m, mapping)
		for _, group := range groups {
			group.Map(mapping)
		}
	}
	for _, group := range groups {
		group.Repair(m, epsilon)
	}
}

func (d *DualContouring) repairEpsilon() float64 {
	if d.RepairEpsilon == 0 {
		return DefaultDualContouringRepairEpsilon * d.Delta
	}
	return d.RepairEpsilon * d.Delta
}

func (d *DualContouring) singularValueEpsilon() float64 {
	if d.SingularValueEpsilon != 0 {
		return d.SingularValueEpsilon
	}
	return DefaultDualContouringSingularValueEpsilon
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

func (d *dcCubeLayout) PointCubeMinMax(c Coord3D) (min, max Coord3D) {
	arrs := [3][]float64{d.Xs, d.Ys, d.Zs}
	var result [3]int
	for i, axisValue := range c.Array() {
		arr := arrs[i]
		idx := sort.SearchFloat64s(arr, axisValue)
		if idx <= 0 {
			idx = 1
		} else if idx == len(arr) {
			idx -= 1
		}
		result[i] = idx - 1
	}
	min = XYZ(d.Xs[result[0]], d.Ys[result[1]], d.Zs[result[2]])
	max = XYZ(d.Xs[result[0]+1], d.Ys[result[1]+1], d.Zs[result[2]+1])
	return
}

func (d *dcCubeLayout) UsableEdges() int {
	atBottom := d.ZOffset+d.BufRows == len(d.Zs)
	xCount, yCount, _ := d.edgeCounts()
	endIdx := len(d.Edges)
	if !atBottom {
		endIdx -= xCount + yCount
	}
	return endIdx
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
			cubeAt(x, y, z-1),
			cubeAt(x, y-1, z-1),
			cubeAt(x, y-1, z),
			cubeAt(x, y, z),
		}
	} else if edgeIdx < xCount+yCount {
		edgeIdx -= xCount
		x := edgeIdx % len(d.Xs)
		y := edgeIdx / len(d.Xs)
		return [4]dcCubeIdx{
			cubeAt(x-1, y, z),
			cubeAt(x-1, y, z-1),
			cubeAt(x, y, z-1),
			cubeAt(x, y, z),
		}
	} else {
		edgeIdx -= xCount + yCount
		x := edgeIdx % len(d.Xs)
		y := edgeIdx / len(d.Xs)
		return [4]dcCubeIdx{
			cubeAt(x, y-1, z),
			cubeAt(x-1, y-1, z),
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

type singularEdgeGroup struct {
	Groups     [][2]*Triangle
	InwardDirs []Coord3D
	Edge       Segment
}

func newSingularEdgeGroup(m *Mesh, s Segment, tris []*Triangle) *singularEdgeGroup {
	if len(tris)%2 != 0 {
		panic("invalid triangle count")
	}
	axis := s[0].Sub(s[1]).Normalize()
	b1, b2 := axis.OrthoBasis()
	mp := s.Mid()
	thetas := make([]float64, len(tris))
	normalDirs := make([]bool, len(tris))
	for i, t := range tris {
		triVec := s.Other(t).Sub(mp).Normalize()
		x, y := b1.Dot(triVec), b2.Dot(triVec)
		thetas[i] = math.Atan2(y, x)

		// The normal should be tangent to the circle,
		// and we record whether it is facing clockwise
		// or counterclockwise, since this indicates
		// containment.
		normal := t.Normal()
		nx, ny := b1.Dot(normal), b2.Dot(normal)
		normalDirs[i] = (nx*y - ny*x) > 0
	}
	essentials.VoodooSort(thetas, func(i, j int) bool {
		return thetas[i] < thetas[j]
	}, tris, normalDirs)

	if len(tris) > 2 && normalDirs[0] {
		// To be consistent across edges, we always move points
		// away from the solid to create volume where there was
		// previously none.
		t := tris[0]
		copy(tris, tris[1:])
		tris[len(tris)-1] = t

		backup := thetas[0] + math.Pi*2
		copy(thetas, thetas[1:])
		thetas[len(thetas)-1] = backup
	}

	// Pair every triangle with the next triangle that has the
	// opposite orientation. This is only necessary if the mesh
	// is self-intersecting; otherwise, the pairs should
	// already be correctly oriented.
	for i := 0; i < len(tris); i += 2 {
		t1 := tris[i]
		for j := i + 1; j < len(tris); j++ {
			t2 := tris[j]
			if segmentOrientation(t1, s) != segmentOrientation(t2, s) {
				if j != i+1 {
					tris[i+1], tris[j] = tris[j], tris[i+1]
					thetas[i+1], thetas[j] = thetas[j], thetas[i+1]
				}
				break
			}
		}
	}

	groups := make([][2]*Triangle, 0, len(tris)/2)
	dirs := make([]Coord3D, 0, len(tris)/2)
	for i := 0; i < len(tris); i += 2 {
		groups = append(groups, [2]*Triangle{tris[i], tris[i+1]})

		// Store the direction facing towards the middle of the
		// triangle pair.
		theta := (thetas[i] + thetas[i+1]) / 2
		dirs = append(dirs, b1.Scale(math.Cos(theta)).Add(b2.Scale(math.Sin(theta))))
	}
	return &singularEdgeGroup{
		Groups:     groups,
		InwardDirs: dirs,
		Edge:       s,
	}
}

func (s *singularEdgeGroup) Constrain(m *Mesh, epsilon float64, layout *dcCubeLayout) *CoordMap[Coord3D] {
	points := NewCoordMap[Coord3D]()
	for _, g := range s.Groups {
		for _, t := range g {
			for _, c := range t {
				if _, ok := points.Load(c); !ok {
					min, max := layout.PointCubeMinMax(c)
					min = min.AddScalar(epsilon)
					max = max.AddScalar(-epsilon)
					points.Store(c, c.Min(max).Max(min))
				}
			}
		}
	}
	return points
}

func (s *singularEdgeGroup) Map(mapping *CoordMap[Coord3D]) {
	s.Edge[0] = mapping.Value(s.Edge[0])
	s.Edge[1] = mapping.Value(s.Edge[1])
}

func (s *singularEdgeGroup) RecomputeGroups(m *Mesh) {
	*s = *newSingularEdgeGroup(m, s.Edge, m.Find(s.Edge[0], s.Edge[1]))
}

func (s *singularEdgeGroup) Repair(m *Mesh, epsilon float64) {
	// Might be necessary if one of our triangles was
	// removed and replaced by a previous repair.
	s.RecomputeGroups(m)

	mp := s.Edge.Mid()
	for i, group := range s.Groups {
		inwardDir := s.InwardDirs[i]
		newMp := mp.Add(inwardDir.Scale(epsilon))
		if len(m.Find(newMp)) > 0 {
			panic("repair point already exists; this should have very low probability.")
		}
		for _, t := range group {
			other := s.Edge.Other(t)
			t1 := &Triangle{other, s.Edge[0], newMp}
			t2 := &Triangle{other, newMp, s.Edge[1]}
			sharedSeg := NewSegment(other, s.Edge[0])
			if segmentOrientation(t1, sharedSeg) != segmentOrientation(t, sharedSeg) {
				t1[0], t1[1] = t1[1], t1[0]
				t2[0], t2[1] = t2[1], t2[0]
			}
			m.Remove(t)
			m.Add(t1)
			m.Add(t2)
		}
	}
}

func singularEdgeGroups(m *Mesh) []*singularEdgeGroup {
	counts := NewEdgeToSlice[*Triangle]()
	var results []*singularEdgeGroup
	m.Iterate(func(t *Triangle) {
		for _, s := range t.Segments() {
			counts.Append(s, t)
		}
	})
	counts.Range(func(key [2]Coord3D, tris []*Triangle) bool {
		if len(tris) > 2 {
			results = append(results, newSingularEdgeGroup(m, Segment(key), tris))
		}
		return true
	})
	return results
}

func segmentOrientation(t *Triangle, s Segment) bool {
	for i, x := range t {
		if x == s[0] {
			return t[(i+2)%3] == s[1]
		}
	}
	panic("first segment point not in triangle")
}

type singularVertexGroup struct {
	Groups [][]*Triangle
	Vertex Coord3D
}

func (s *singularVertexGroup) Constrain(m *Mesh, epsilon float64, layout *dcCubeLayout,
	origPoints *CoordMap[bool]) *CoordMap[Coord3D] {
	points := NewCoordMap[Coord3D]()
	for _, g := range s.Groups {
		for _, t := range g {
			for _, c := range t {
				if _, ok := points.Load(c); !ok {
					if _, mask := origPoints.Load(c); !mask {
						continue
					}
					min, max := layout.PointCubeMinMax(c)
					min = min.AddScalar(epsilon)
					max = max.AddScalar(-epsilon)
					points.Store(c, c.Min(max).Max(min))
				}
			}
		}
	}
	return points
}

func (s *singularVertexGroup) Map(mapping *CoordMap[Coord3D]) {
	s.Vertex = mapping.Value(s.Vertex)
}

func (s *singularVertexGroup) Repair(m *Mesh, epsilon float64) {
	for _, group := range s.Groups {
		// Move point inwards according to the approximate
		// vertex normal, as computed using MWA:
		// "A Comparison of Algorithms for Vertex Normal Computations"
		// http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.99.2846&rep=rep1&type=pdf.
		var d Coord3D
		for _, t := range group {
			seg := t.otherSegment(s.Vertex)
			dot := seg[0].Sub(s.Vertex).Normalize().Dot(seg[1].Sub(s.Vertex).Normalize())
			theta := math.Acos(math.Max(-1.0, math.Min(1.0, dot)))
			d = d.Add(t.Normal().Scale(-theta))
		}
		norm := d.Norm()
		if norm == 0 {
			panic("vertex normal is unknown at singular vertex")
		}
		d = d.Scale(epsilon / norm)
		newPoint := s.Vertex.Add(d)
		for _, t := range group {
			m.Remove(t)
			for i, c := range t {
				if c == s.Vertex {
					t[i] = newPoint
				}
			}
			m.Add(t)
		}
	}
}

func singularVertexGroups(m *Mesh) []*singularVertexGroup {
	p := newPtrMeshMesh(m)
	var results []*singularVertexGroup
	p.IterateCoords(func(c *ptrCoord) {
		clusters := c.Clusters()
		if len(clusters) > 1 {
			group := &singularVertexGroup{
				Groups: make([][]*Triangle, len(clusters)),
				Vertex: c.Coord3D,
			}
			for i, cluster := range clusters {
				for _, t := range cluster {
					orig := m.Find(t.Coords[0].Coord3D, t.Coords[1].Coord3D, t.Coords[2].Coord3D)[0]
					group.Groups[i] = append(group.Groups[i], orig)
				}
			}
			results = append(results, group)
		}
	})
	return results
}

func mapInPlace(m *Mesh, mapping *CoordMap[Coord3D]) {
	m.Iterate(func(t *Triangle) {
		var changed bool
		for _, c := range t {
			if _, ok := mapping.Load(c); ok {
				changed = true
			}
		}
		if changed {
			m.Remove(t)
			for i, c := range t {
				if c1, ok := mapping.Load(c); ok {
					t[i] = c1
				}
			}
			m.Add(t)
		}
	})
}
