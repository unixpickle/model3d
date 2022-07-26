package model3d

import (
	"fmt"
	"math"
	"sort"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/numerical"
)

const (
	Floater97DefaultMSETol   = 1e-12
	Floater97DefaultMaxIters = 5000
)

// BuildAutomaticUVMap creates a MeshUVMap for an entire
// mesh which fits in the unit square (0, 0) to (1, 1) and
// should work best at the given resolution.
//
// The resolution specifies the side-length of the targeted
// texture image. It must be a power of two. This is used
// to determine spacing in the final layout.
//
// The mesh itself should be manifold, but needn't have any
// special kind of topology.
//
// This is meant for quick applications that don't need a
// lot of control over the resulting parameterization. The
// underlying algorithm and exact results are subject to
// change.
func BuildAutomaticUVMap(m *Mesh, resolution int) MeshUVMap {
	foundPower := false
	for i := 0; i < 32; i++ {
		if 1<<uint(i) == resolution {
			foundPower = true
			break
		}
	}
	if !foundPower {
		panic("resolution must be power of 2")
	}

	// Attempt to target a constant number of patches by
	// putting a limit on the triangles per patch.
	nTris := essentials.MaxInt(128, len(m.TriangleSlice())/50)
	discs := MeshToPlaneGraphsLimited(m, nTris)

	params := make([]MeshUVMap, len(discs))
	for i, disc := range discs {
		parameterization := Floater97(
			disc,
			PNormBoundary(disc, 4), // Almost square, but no colinear points.
			Floater97ShapePreservingWeights(disc),
			nil,
		)
		ExtendBoundaryUVs(disc, parameterization)
		params[i] = NewMeshUVMapForCoords(disc, parameterization)
	}
	return PackMeshUVMaps(
		model2d.XY(0, 0),
		model2d.XY(1, 1),
		0.5/float64(resolution),
		params,
	)
}

// CircleBoundary computes a mapping of the boundary of a
// mesh m to the unit circle based on segment length.
//
// The mesh must be properly oriented, and be manifold
// except along the boundary.
// The mesh's boundary must contain at least three segments
// which are all connected in a cycle. This means that the
// mesh must be mappable to a disc.
func CircleBoundary(m *Mesh) *CoordMap[model2d.Coord] {
	points := boundarySequence(m)
	totalLength := 0.0
	for i, p := range points {
		p1 := points[(i+1)%len(points)]
		totalLength += p1.Dist(p)
	}

	mapping := NewCoordMap[model2d.Coord]()
	mapping.Store(points[0], model2d.X(1.0))
	curLength := 0.0
	for i, p := range points {
		p1 := points[(i+1)%len(points)]
		curLength += p1.Dist(p)
		theta := 2 * math.Pi * curLength / totalLength
		mapping.Store(p1, model2d.XY(math.Cos(theta), math.Sin(theta)))
	}

	return mapping
}

// SquareBoundary computes a mapping of the boundary of a
// mesh m to the unit square. This may result in some
// triangles being mapped to three colinear points if the
// boundary contains two consecutive segments from one
// triangle.
//
// See CircleBoundary for restrictions on the mesh m.
func SquareBoundary(m *Mesh) *CoordMap[model2d.Coord] {
	res := NewCoordMap[model2d.Coord]()
	CircleBoundary(m).Range(func(k Coord3D, v model2d.Coord) bool {
		// Scale each coordinate so that it lands on the unit square.
		//
		// This introduces some extra stretch at the corners, but at
		// least it's easy to reason about the final boundary being
		// convex. This could be changed in the future to better
		// preserve arc-length.
		res.Store(k, v.Scale(1/v.Abs().MaxCoord()))
		return true
	})
	return res
}

// PNormBoundary is similar to CircleBoundary, except that
// the circle is defined under any p-norm, not just p=2.
func PNormBoundary(m *Mesh, p float64) *CoordMap[model2d.Coord] {
	res := NewCoordMap[model2d.Coord]()
	CircleBoundary(m).Range(func(k Coord3D, v model2d.Coord) bool {
		abs := v.Abs()
		pNorm := math.Pow(math.Pow(abs.X, p)+math.Pow(abs.Y, p), 1/p)
		res.Store(k, v.Scale(1/pNorm))
		return true
	})
	return res
}

func boundarySequence(m *Mesh) []Coord3D {
	vertexToNext := NewCoordMap[Coord3D]()

	var start Coord3D
	m.Iterate(func(t *Triangle) {
		for i := 0; i < 3; i++ {
			p1, p2 := t[i], t[(i+1)%3]
			if len(m.Find(p1, p2)) == 1 {
				vertexToNext.Store(p1, p2)
				start = p1
			}
		}
	})
	if vertexToNext.Len() == 0 {
		panic("the mesh did not contain any boundary edges")
	}

	res := make([]Coord3D, 1, vertexToNext.Len())
	res[0] = start
	cur := vertexToNext.Value(start)
	for cur != start {
		res = append(res, cur)
		var ok bool
		cur, ok = vertexToNext.Load(cur)
		if !ok {
			panic("mesh is non-manifold, not oriented consistently, or has an invalid boundary")
		}
	}
	if len(res) < vertexToNext.Len() {
		panic("mesh has multiple, non-connected boundaries")
	}
	return res
}

// ExtendBoundaryUVs rescales vertices of triangles on the
// boundary of a plane graph triangulation to ensure that
// these triangles are not highly stretched or even fully
// degenerate.
//
// It is assumed that the boundary parameterization is
// centered around the origin, as done by CircleBoundary()
// and similar helpers.
func ExtendBoundaryUVs(m *Mesh, param *CoordMap[model2d.Coord]) {
	boundary := boundarySequence(m)
	for i, p1 := range boundary {
		p0 := boundary[(i+len(boundary)-1)%len(boundary)]
		p2 := boundary[(i+1)%len(boundary)]
		if tris := m.Find(p0, p1, p2); len(tris) == 1 {
			uv0, uv1, uv2 := param.Value(p0), param.Value(p1), param.Value(p2)

			seg3d := NewSegment(p0, p2)
			ratio3d := seg3d.Dist(p1) / seg3d.Length()

			seg2d := model2d.Segment{uv0, uv2}
			dist2d := seg2d.Dist(uv1)
			ratio2d := dist2d / seg2d.Length()

			if ratio2d >= ratio3d {
				// The UV triangle is already less degenerate.
				continue
			}

			extraDist := (ratio3d - ratio2d) * seg2d.Length()
			direction := uv1.ProjectOut(uv2.Sub(uv0)).Normalize()
			param.Store(p1, uv1.Add(direction.Scale(extraDist)))
		}
	}
}

// Floater97UniformWeights computes the uniform weighting
// scheme for the edgeWeights argument of Floater97().
// This is the simplest possible weighting scheme, but may
// result in distortion.
//
// As proved in Floater (1997), this weighting attempts to
// minimize the sum of squares of edge lengths in the
// resulting parameterization.
func Floater97UniformWeights(m *Mesh) *EdgeMap[float64] {
	res := NewEdgeMap[float64]()
	m.AllVertexNeighbors().Range(func(k Coord3D, v []Coord3D) bool {
		w := 1 / float64(len(v))
		for _, neighbor := range v {
			res.Store([2]Coord3D{k, neighbor}, w)
		}
		return true
	})
	return res
}

// Floater97InvChordLengthWeights computes the weighting
// scheme based on 1/(distance^r), where r is some exponent
// applied to the distance between points along each edge.
func Floater97InvChordLengthWeights(m *Mesh, r float64) *EdgeMap[float64] {
	res := NewEdgeMap[float64]()
	m.AllVertexNeighbors().Range(func(k Coord3D, v []Coord3D) bool {
		weights := make([]float64, len(v))
		total := 0.0
		for i, neighbor := range v {
			w := 1 / math.Pow(k.Dist(neighbor), r)
			weights[i] = w
			total += w
		}
		invTotal := 1 / total
		for i, neighbor := range v {
			res.Store([2]Coord3D{k, neighbor}, invTotal*weights[i])
		}
		return true
	})
	return res
}

// Floater97ShapePreservingWeights computes the
// shape-preserving weights described in Floater (1997).
//
// The mesh must be properly connected, consistently
// oriented, and have exactly one boundary loop. Otherwise,
// this may panic() or return invalid results.
func Floater97ShapePreservingWeights(m *Mesh) *EdgeMap[float64] {
	boundaryMap := NewCoordMap[bool]()
	for _, c := range boundarySequence(m) {
		boundaryMap.Store(c, true)
	}

	res := NewEdgeMap[float64]()
	for _, center := range m.VertexSlice() {
		if boundaryMap.Value(center) {
			// The local parameterization weight strategy does not
			// make any sense for boundary vertices, and we don't
			// need these weights for the linear system.
			continue
		}
		neighbors, weights := localParameterizationWeights(m, center)
		for i, n := range neighbors {
			res.Store([2]Coord3D{center, n}, weights[i])
		}
	}

	return res
}

// Floater97DefaultSolver creates a reasonable numerical
// solver for most small-to-medium parameterization
// systems.
//
// Floater97DefaultMaxIters and Floater97DefaultMAETol
// are used as stopping criteria.
func Floater97DefaultSolver() *numerical.BiCGSTABSolver {
	return &numerical.BiCGSTABSolver{
		MaxIters:     Floater97DefaultMaxIters,
		MSETolerance: Floater97DefaultMSETol,
	}
}

// Floater97 computes the 2D parameterization of a mesh.
//
// The mesh m must be a simple-connected triangulated plane
// graph; in other words, it must be mappable to a disc.
// The boundary of this mesh must contain at least three
// points, as is the case for a single triangle.
//
// The boundary argument maps each boundary vertex in m to
// a coordinate on the 2D plane. The boundary must be a
// convex polygon for the resulting parameterization to be
// valid.
//
// The edgeWeights argument maps ordered pairs of connected
// vertices to a non-negative weight, where the first
// vertex in each pair is the "center" and the second
// vertex is a neighbor of that center.
// Boundary vertices are never used as centers, so these
// vertices never need to be the first vertex in an edge.
// For every center vertex, the weights of all of its
// connected edges must sum to 1 (sum_j of w[i][j] = 1).
//
// The solver argument should be able to solve the sparse
// linear system produced by the algorithm efficiently.
// If nil is provided, Floater97DefaultSolver() is used.
//
// The returned mapping assigns a 2D coordinate to every
// vertex in the original mesh, including the fixed
// boundary vertices.
//
// This is based on the paper:
// "Parametrization and smooth approximation of surface triangulations"
// (Floater, 1996). https://www.cs.jhu.edu/~misha/Fall09/Floater97.pdf
func Floater97(m *Mesh, boundary *CoordMap[model2d.Coord],
	edgeWeights *EdgeMap[float64], solver numerical.LargeLinearSolver) *CoordMap[model2d.Coord] {
	// Map coordinates to all their neighbors.
	neighbors := NewCoordToSlice[Coord3D]()
	m.Iterate(func(t *Triangle) {
		for i, c := range t {
			for j, c1 := range t {
				if i != j {
					cur := neighbors.Value(c)
					found := false
					for _, c2 := range cur {
						if c2 == c1 {
							found = true
							continue
						}
					}
					if !found {
						neighbors.Store(c, append(cur, c1))
					}
				}
			}
		}
	})

	// Map rows of system to non-boundary vertices.
	nonBoundaryToIndex := NewCoordMap[int]()
	nonBoundary := []Coord3D{}
	for _, v := range m.VertexSlice() {
		if _, ok := boundary.Load(v); !ok {
			nonBoundary = append(nonBoundary, v)
			nonBoundaryToIndex.Store(v, len(nonBoundary)-1)
		}
	}

	// We will solve matrix*x = bias.
	matrix := numerical.NewSparseMatrix(len(nonBoundary))
	bias := make([]numerical.Vec2, len(nonBoundary))
	for i, center := range nonBoundary {
		matrix.Set(i, i, -1.0)
		total := 0.0
		for _, neighbor := range neighbors.Value(center) {
			weight, ok := edgeWeights.Load([2]Coord3D{center, neighbor})
			if !ok {
				panic(fmt.Sprintf("missing edge weight between %v and %v", center, neighbor))
			}
			if weight < 0 {
				panic(fmt.Sprintf("weight %f should not be negative", weight))
			}
			j, ok := nonBoundaryToIndex.Load(neighbor)
			if !ok {
				// This is a boundary, so we don't actually have a
				// variable for it. Instead, we have a constant, and
				// it goes on the right-hand side of the equation.
				bias[i] = bias[i].Add(boundary.Value(neighbor).Scale(-weight).Array())
			} else {
				matrix.Set(i, j, weight)
			}
			total += weight
		}
		if math.Abs(total-1.0) > 1e-4 {
			panic(fmt.Sprintf("total edge weight must add up to 1.0 for every vertex, got %f",
				total))
		}
	}

	if solver == nil {
		solver = Floater97DefaultSolver()
	}
	solution := make([]numerical.Vec2, len(bias))
	for i := 0; i < 2; i++ {
		bias1d := make([]float64, len(bias))
		for j, v := range bias {
			bias1d[j] = v[i]
		}
		for j, x := range solver.SolveLinearSystem(matrix.Apply, bias1d) {
			solution[j][i] = x
		}
	}

	result := NewCoordMap[model2d.Coord]()
	boundary.Range(func(k Coord3D, v model2d.Coord) bool {
		result.Store(k, v)
		return true
	})
	for i, point := range solution {
		result.Store(nonBoundary[i], model2d.NewCoordArray(point))
	}
	return result
}

func localParameterizationWeights(m *Mesh, center Coord3D) ([]Coord3D, []float64) {
	ps := orderedNeighbors(m, center)

	// Compute a local parameterization using Section 3.1 of
	// "Free-Form Shape Design Using Triangulated Surfaces"
	// (https://www.cs.cmu.edu/~aw/pdf/tri.pdf).
	angles := make([]float64, len(ps))
	totalAngle := 0.0
	for i := 0; i < len(ps); i++ {
		p1 := ps[i]
		p2 := ps[(i+1)%len(ps)]
		angles[i] = totalAngle

		v1 := p1.Sub(center).Normalize()
		v2 := p2.Sub(center).Normalize()
		totalAngle += math.Acos(math.Max(0, math.Min(1, v1.Dot(v2))))
	}
	for i := range angles {
		angles[i] *= 2 * math.Pi / totalAngle
	}
	ps2d := make([]Coord2D, len(ps))
	for i, theta := range angles {
		dist := ps[i].Dist(center)
		ps2d[i] = model2d.XY(math.Cos(theta), math.Sin(theta)).Scale(dist)
	}

	baryCoords := make([]float64, len(ps))
	for i, theta := range angles {
		oppositeTheta := theta + math.Pi
		if oppositeTheta > 2*math.Pi {
			oppositeTheta -= 2 * math.Pi
		}
		index := sort.SearchFloat64s(angles, oppositeTheta)
		i1 := (index + (len(ps) - 1)) % len(ps)
		i2 := index % len(ps)
		if i1 == i || i2 == i {
			panic("impossible opposite edge situation; mesh might contain degenerate triangles")
		}

		p1 := ps2d[i]
		p2 := ps2d[i1]
		p3 := ps2d[i2]

		// Compute barycentric coordinates of origin in triangle(p1, p2, p3).
		mat := model2d.NewMatrix2Columns(p2.Sub(p1), p3.Sub(p1))
		b23 := mat.MulColumnInv(model2d.Origin.Sub(p1), mat.Det())
		b2 := math.Max(0, math.Min(1, b23.X))
		b3 := math.Max(0, math.Min(1, b23.Y))
		b1 := math.Max(0, 1-(b2+b3))

		baryCoords[i] += b1 / float64(len(ps))
		baryCoords[i1] += b2 / float64(len(ps))
		baryCoords[i2] += b3 / float64(len(ps))
	}

	return ps, baryCoords
}

func orderedNeighbors(m *Mesh, center Coord3D) []Coord3D {
	vertexToNext := NewCoordMap[Coord3D]()
	var start Coord3D
	for _, t := range m.Find(center) {
		for i := 0; i < 3; i++ {
			p1 := t[i]
			p2 := t[(i+1)%3]
			if p1 == center || p2 == center {
				continue
			}
			vertexToNext.Store(p1, p2)
			start = p1
		}
	}

	cur := vertexToNext.Value(start)
	res := make([]Coord3D, 1, vertexToNext.Len())
	res[0] = start
	for cur != start {
		res = append(res, cur)
		var ok bool
		cur, ok = vertexToNext.Load(cur)
		if !ok {
			panic("inconsistent neighbor ring around vertex; mesh might be incorrectly oriented.")
		}
	}

	return res
}

func triangleStretchAndArea(t *Triangle, m *CoordMap[model2d.Coord]) (stretchSq, area float64) {
	p2d := [3]model2d.Coord{}
	for i, c := range t {
		p2d[i] = m.Value(c)
	}

	// The mapping from 2D to 3D is equivalent to doing:
	// 2d coordinate -> barycentric -> 3d coordinate
	// which is represented as: A*U^-1.
	// Then we compute the SVD of U*A^T*A*U^-1.

	u := model2d.NewMatrix2Columns(p2d[1].Sub(p2d[0]), p2d[2].Sub(p2d[0]))
	a1 := t[1].Sub(t[0])
	a2 := t[2].Sub(t[0])
	aDiag := a1.Dot(a2)
	aTa := &model2d.Matrix2{a1.Dot(a1), aDiag, aDiag, a2.Dot(a2)}

	mapSq := u.Mul(aTa).Mul(u.Inverse())
	lambdas := mapSq.Eigenvalues()

	return (real(lambdas[0]) + real(lambdas[1])) / 2, u.Det() / 2
}

// MeshToPlaneGraphs splits a mesh m into one or more
// sub-meshes which are simply-connected triangulated plane
// graphs. These sub-meshes are suitable for Floater97().
//
// The mesh m must either be manifold, or be a subset of a
// manifold mesh. For example, calling MeshToPlaneGraphs()
// on a result of MeshToPlaneGraphs() should be an identity
// operation.
func MeshToPlaneGraphs(m *Mesh) []*Mesh {
	return MeshToPlaneGraphsLimited(m, 0)
}

// MeshToPlaneGraphsLimited is like MeshToPlaneGraphs, but
// limits the number of triangles per sub-mesh.
func MeshToPlaneGraphsLimited(m *Mesh, maxSize int) []*Mesh {
	m = m.Copy()
	var res []*Mesh
	for {
		next := nextMeshDiscs(m, maxSize)
		if len(next) > 0 {
			res = append(res, next...)
		} else {
			break
		}
	}
	return res
}

func nextMeshDiscs(m *Mesh, maxSize int) []*Mesh {
	var t1 *Triangle
	for t := range m.faces {
		t1 = t
		break
	}
	if t1 == nil {
		return nil
	}
	m.Remove(t1)

	// As we add triangles, we will track the cumulative
	// area at each triangle, so that we can possibly split
	// the resulting mesh into two halves.
	tris := []*Triangle{t1}
	cumAreas := []float64{t1.Area()}

	// The algorithm tracks the current boundary in terms
	// of segments and vertices. Since vertices might be
	// present in multiple segments, we reference count
	// them.
	segments := NewEdgeMap[bool]()
	vertices := NewCoordToNumber[int]()
	for _, s := range t1.Segments() {
		segments.Store(s, true)
	}
	for _, c := range t1 {
		vertices.Store(c, 1)
	}

	// We now search over triangles using a queue. The
	// queue consists of triangles which currently touch
	// the boundary; not all triangles can actually be
	// added.
	neighborQueue := m.Neighbors(t1)
	inQueue := map[*Triangle]bool{}
	for _, t := range neighborQueue {
		inQueue[t] = true
	}
	for len(neighborQueue) > 0 && (maxSize == 0 || len(tris) < maxSize) {
		next := neighborQueue[0]
		delete(inQueue, next)
		neighborQueue = neighborQueue[1:]

		// If we add a new triangle from one part of the boundary
		// with a vertex touching a separate part of the boundary,
		// we will split the boundary into two disjoint sections.
		//
		// Visual explanation:
		//
		// ---old_boundary----
		// new \   / other
		// half \ /  half
		//       +
		// ---old_boundary----
		//
		// To avoid the above scenario, we skip triangles with a
		// vertex that is not part of an edge already on the
		// boundary. This can happen even if the triangle will
		// eventually be incorporated; it might just need a
		// different neighboring triangle to be added first.
		touchingSegment := [3]bool{}
		for i, c := range next {
			c1 := next[(i+1)%3]
			seg := NewSegment(c, c1)
			if segments.Value(seg) {
				touchingSegment[i] = true
				touchingSegment[(i+1)%3] = true
			}
		}
		wouldDivideBoundary := false
		for i, c := range next {
			if vertices.Value(c) > 0 && !touchingSegment[i] {
				wouldDivideBoundary = true
				break
			}
		}
		if wouldDivideBoundary {
			// The triangle may be re-discovered later when it can be
			// added without creating two boundaries.
			continue
		}

		m.Remove(next)
		tris = append(tris, next)
		cumAreas = append(cumAreas, cumAreas[len(cumAreas)-1]+next.Area())
		for _, seg := range next.Segments() {
			if segments.Value(seg) {
				segments.Delete(seg)
				for _, p := range seg {
					if vertices.Add(p, -1) == 0 {
						vertices.Delete(p)
					}
				}
			} else {
				segments.Store(seg, true)
				for _, p := range seg {
					vertices.Add(p, 1)
				}
			}
		}
		for _, neighbor := range m.Neighbors(next) {
			if !inQueue[neighbor] {
				neighborQueue = append(neighborQueue, neighbor)
				inQueue[neighbor] = true
			}
		}
	}

	if segments.Len() == 0 {
		// We completely covered a surface that was isomorphic
		// to a sphere, with no boundary left at the final step.
		// We must produce two discs, and we try to divide them
		// as evenly as possible.
		index := sort.SearchFloat64s(cumAreas, cumAreas[len(cumAreas)-1]/2)
		if index > len(tris)-1 {
			index = len(tris) - 1
		}
		return []*Mesh{NewMeshTriangles(tris[:index]), NewMeshTriangles(tris[index:])}
	}

	return []*Mesh{NewMeshTriangles(tris)}
}

// A MeshUVMap is a mapping between triangles in a 3D mesh
// and triangles on a 2D surface.
//
// The order of 3D triangles corresponds to the order of 2D
// triangles (e.g. tri3d[i] maps to tri2d[i], 0 <= i < 3).
type MeshUVMap map[*Triangle][3]model2d.Coord

// JoinMeshUVMaps adds all keys and values from
// all UV maps to a resulting mapping.
//
// This will not modify the coordinates in the mappings.
func JoinMeshUVMaps(ms ...MeshUVMap) MeshUVMap {
	res := MeshUVMap{}
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// PackMeshUVMaps rescales and combines all of the provided
// UV maps into a single rectangle given by the bounds
// min and max.
//
// The border argument is an amount of space to put around
// the edges of each separate UV map in the texture to
// avoid interpolation from mixing them.
func PackMeshUVMaps(min, max model2d.Coord, border float64,
	params []MeshUVMap) MeshUVMap {
	tree := newParamQuadTree(params)
	return tree.Joined(border, min, max)
}

// NewMeshUVMapForCoords maps triangles in the mesh to 2D
// triangles using direct per-point lookups.
//
// The mapping must have an entry for every vertex in the
// mesh.
func NewMeshUVMapForCoords(mesh *Mesh, mapping *CoordMap[model2d.Coord]) MeshUVMap {
	res := MeshUVMap{}
	mesh.Iterate(func(t *Triangle) {
		var mapped [3]model2d.Coord
		for i, c := range t {
			if value, ok := mapping.Load(c); ok {
				mapped[i] = value
			} else {
				panic("coordinate not present in mapping")
			}
		}
		res[t] = mapped
	})
	return res
}

// MapFn creates a function that maps 2D coordinates to 3D
// using the UV map.
//
// The resulting function also returns the 3D triangle
// corresponding to the mapped point.
//
// Resulting 3D points will always be produced, even if the
// 2D point lands outside the 2D triangulation. In this
// case, the nearest 2D point on the triangulation is used.
func (m MeshUVMap) MapFn() func(c model2d.Coord) (Coord3D, *Triangle) {
	tris := make([]*model2d.Triangle, 0, len(m))
	invMap := map[*model2d.Triangle]*Triangle{}
	for t3d, ps2d := range m {
		t2d := model2d.NewTriangle(ps2d[0], ps2d[1], ps2d[2])
		tris = append(tris, t2d)
		invMap[t2d] = t3d
	}

	model2d.GroupBounders(tris)
	lookup := newTri2dLookup(tris)
	if math.IsNaN(lookup.bounds.Max().Sub(lookup.bounds.Min()).Norm()) {
		panic("NaN detected in bounds; possibly degenerate mapping")
	}

	// The numerical precision of collision detection will
	// vary with the overall scale of 2D coordinates.
	epsilon := 1e-8 * lookup.bounds.Min().Abs().Max(lookup.bounds.Max().Abs()).MaxCoord()

	return func(c model2d.Coord) (Coord3D, *Triangle) {
		t2d, abc := lookup.Find(c, epsilon)
		t3d := invMap[t2d]
		return t3d.AtBarycentric(abc), t3d
	}
}

// Bounds2D gets the bounding box of the 2D triangles.
func (m MeshUVMap) Bounds2D() (min, max model2d.Coord) {
	first := true
	for _, t2d := range m {
		for _, c := range t2d {
			if first {
				first = false
				min = c
				max = c
			} else {
				min = min.Min(c)
				max = max.Max(c)
			}
		}
	}
	return
}

// ToBounds creates a new UV map where the 2D bounding box
// is rescaled and translated to a new min and max.
func (m MeshUVMap) ToBounds(min, max model2d.Coord) MeshUVMap {
	if !model2d.BoundsValid(model2d.NewRect(min, max)) {
		panic("bounds are invalid")
	}
	oldMin, oldMax := m.Bounds2D()
	scale := max.Sub(min).Div(oldMax.Sub(oldMin))

	res := MeshUVMap{}
	for k, v := range m {
		var newTri [3]model2d.Coord
		for i, c := range v {
			newTri[i] = c.Sub(oldMin).Mul(scale).Add(min)
		}
		res[k] = newTri
	}
	return res
}

// Area3D gets the total area of all the 3D triangles.
func (m MeshUVMap) Area3D() float64 {
	var sum float64
	for k := range m {
		sum += k.Area()
	}
	return sum
}

type tri2dLookup struct {
	bounds   model2d.Rect
	root     *model2d.Triangle
	children []*tri2dLookup
}

func newTri2dLookup(grouped []*model2d.Triangle) *tri2dLookup {
	if len(grouped) == 1 {
		return &tri2dLookup{
			bounds: *model2d.BoundsRect(grouped[0]),
			root:   grouped[0],
		}
	}
	i := len(grouped) / 2
	ch1 := newTri2dLookup(grouped[:i])
	ch2 := newTri2dLookup(grouped[i:])
	return &tri2dLookup{
		bounds: *model2d.NewRect(
			ch1.bounds.Min().Min(ch2.bounds.Min()),
			ch1.bounds.Max().Max(ch2.bounds.Max()),
		),
		children: []*tri2dLookup{ch1, ch2},
	}
}

func (t *tri2dLookup) Find(c model2d.Coord, epsilon float64) (*model2d.Triangle, [3]float64) {
	// Perfect containment lookup is faster than nearest
	// point lookup, and should often be sufficient if the
	// texture covers most of the plane.
	if tri, bary := t.findContains(c); tri != nil {
		return tri, bary
	}

	var resultTri *model2d.Triangle
	var resultBary [3]float64
	resultDist := math.Inf(1)
	t.findNearest(c, &resultTri, &resultBary, &resultDist)
	return resultTri, resultBary
}

func (t *tri2dLookup) findContains(c model2d.Coord) (*model2d.Triangle, [3]float64) {
	if !t.bounds.Contains(c) {
		return nil, [3]float64{}
	}
	if t.root != nil {
		if model2d.InBounds(t.root, c) {
			bary := t.root.Barycentric(c)
			if bary[0] >= 0 && bary[1] >= 0 && bary[2] >= 0 {
				return t.root, bary
			}
		}
		return nil, [3]float64{}
	}
	for _, ch := range t.children {
		if tri, bary := ch.findContains(c); tri != nil {
			return tri, bary
		}
	}
	return nil, [3]float64{}
}

func (t *tri2dLookup) findNearest(c model2d.Coord, tri **model2d.Triangle, coord *[3]float64,
	distBound *float64) {
	if t.root != nil {
		if bary, sdf := t.root.BarycentricSDF(c); sdf > -*distBound {
			*distBound = -sdf
			*tri = t.root
			*coord = bary
		}
		return
	}

	// Try the closer child first, and ignore children that
	// cannot possibly have a closer point.
	chs := [2]*tri2dLookup{t.children[0], t.children[1]}
	ds := [2]float64{
		t.children[0].bounds.SDF(c),
		t.children[1].bounds.SDF(c),
	}
	if ds[0] < ds[1] {
		chs[0], chs[1] = chs[1], chs[0]
		ds[0], ds[1] = ds[1], ds[0]
	}
	for i, ch := range chs {
		d := ds[i]
		if d < -*distBound {
			break
		}
		ch.findNearest(c, tri, coord, distBound)
	}
}

type paramQuadTree struct {
	Leaf MeshUVMap

	// Branches contains at most four elements.
	Branches []*paramQuadTree
}

func newParamQuadTree(params []MeshUVMap) *paramQuadTree {
	sortedParams := append([]MeshUVMap{}, params...)
	sortedAreas := make([]float64, len(params))
	for i, p := range params {
		sortedAreas[i] = p.Area3D()
	}
	essentials.VoodooSort(sortedAreas, func(i, j int) bool {
		return sortedAreas[i] > sortedAreas[j]
	}, sortedParams)
	return buildParamQuadTree(sortedParams, sortedAreas)
}

func buildParamQuadTree(params []MeshUVMap, areas []float64) *paramQuadTree {
	if len(params) == 1 {
		return &paramQuadTree{Leaf: params[0]}
	}
	if len(params) <= 4 {
		branches := make([]*paramQuadTree, len(params))
		for i, x := range params {
			branches[i] = &paramQuadTree{Leaf: x}
		}
		return &paramQuadTree{Branches: branches}
	}

	// Problem: assign parameterizations such that
	// area is distributed as evenly as possible
	// across all four quadrants.
	//
	// For now, we don't do anything particularly
	// intelligent to solve this knapsack problem.
	// Better search algorithms exist for this, but
	// the exact problem is NP-complete.
	var assignments [4][]MeshUVMap
	var assignmentsAreas [4][]float64
	var assignmentsTotals [4]float64

	for i, param := range params {
		area := areas[i]

		minArea := assignmentsTotals[0]
		dstIndex := 0
		for j := 1; j < 4; j++ {
			if assignmentsTotals[j] < minArea {
				minArea = assignmentsTotals[j]
				dstIndex = j
			}
		}

		assignments[dstIndex] = append(assignments[dstIndex], param)
		assignmentsAreas[dstIndex] = append(assignmentsAreas[dstIndex], area)
		assignmentsTotals[dstIndex] += area
	}

	branches := make([]*paramQuadTree, 4)
	for i, pile := range assignments {
		branches[i] = buildParamQuadTree(pile, assignmentsAreas[i])
	}
	return &paramQuadTree{Branches: branches}
}

func (p *paramQuadTree) Joined(border float64, min, max model2d.Coord) MeshUVMap {
	if p.Leaf != nil {
		return p.Leaf.ToBounds(min.AddScalar(border), max.AddScalar(-border))
	}

	if len(p.Branches) == 2 {
		// Split the grid in half along the longer dimension.
		diff := max.Sub(min)
		if diff.Y > diff.X {
			mp := (min.Y + max.Y) / 2
			return JoinMeshUVMaps(
				p.Branches[0].Joined(border, min, model2d.XY(max.X, mp)),
				p.Branches[1].Joined(border, model2d.XY(min.X, mp), max),
			)
		} else {
			mp := (min.X + max.X) / 2
			return JoinMeshUVMaps(
				p.Branches[0].Joined(border, min, model2d.XY(mp, max.Y)),
				p.Branches[1].Joined(border, model2d.XY(mp, min.Y), max),
			)
		}
	}

	// Split up into a grid of four.
	mp := min.Mid(max)
	xs := [3]float64{min.X, mp.X, max.X}
	ys := [3]float64{min.Y, mp.Y, max.Y}
	params := make([]MeshUVMap, len(p.Branches))
	for i, branch := range p.Branches {
		x := i % 2
		y := i / 2
		min := model2d.XY(xs[x], ys[y])
		max := model2d.XY(xs[x+1], ys[y+1])
		params[i] = branch.Joined(border, min, max)
	}

	return JoinMeshUVMaps(params...)
}
