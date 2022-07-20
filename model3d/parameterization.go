package model3d

import (
	"fmt"
	"math"
	"sort"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/numerical"
)

const (
	Floater97DefaultMAETol   = 1e-4
	Floater97DefaultMaxIters = 1000
)

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

// SquareBoundary copmutes a mapping of the boundary of a
// mesh m to the unit square.
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
		MAETolerance: Floater97DefaultMAETol,
	}
}

// Floater97 computes the 2D parameterization of a mesh
// which is disc-like (having at least three boundary
// points).
//
// The m argument is the mesh to parameterize, which must
// be mappable to a disc.
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
