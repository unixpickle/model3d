package model3d

import (
	"math"

	"github.com/unixpickle/model3d/numerical"
)

const (
	QEFDecimatorDefaultMinDet = 1e-8
)

// QEFDecimatorOptions stores configuration for QEFDecimate().
type QEFDecimatorOptions struct {
	// MinDet is the minimum determinant of a 3x3 matrix
	// before it is assumed singular.
	// If 0, defaults to QEFDecimatorDefaultMinDet.
	MinDet float64

	// MinNormalDotProduct is the minimum dot product between
	// the normals of an old and new triangle before a deletion
	// is rejected for "flipping" a face.
	//
	// A value of zero means that faces can't be flipped more
	// than 90 degrees.
	MinNormalDotProduct float64

	// If specified, add this amount of Tikhonov regularization
	// towards each vertex to its quadric, as derived for the
	// probabilistic plane quadric in
	// "Fast and Robust QEF Minimization using Probabilistic Quadrics".
	TikhonovRegularization float64
}

func (q *QEFDecimatorOptions) minDet() float64 {
	if q.MinDet == 0 {
		return QEFDecimatorDefaultMinDet
	}
	return q.MinDet
}

// QEFDecimate applies a mesh simplification technique based on
// quadric error functions.
//
// The mesh is decimated until the minimum number of triangles is
// reached, or no more simplification can be performed.
//
// The input mesh must be manifold. Otherwise, results are undefined.
func QEFDecimate(m *Mesh, minTris int, options *QEFDecimatorOptions) *Mesh {
	if options == nil {
		options = &QEFDecimatorOptions{}
	}
	dec := newQEFDecimator(m, *options)
	for dec.NumTris > minTris {
		if !dec.Step() {
			break
		}
	}
	return dec.Mesh.Mesh()
}

type qefDecimator struct {
	Mesh    *ptrMesh
	PtrMap  ptrCoordMap
	NumTris int

	QEFs  map[*ptrCoord]numerical.QEF4
	Heap  *qefHeap
	Pairs map[*ptrCoord]map[*ptrCoord]struct{}

	// Configuration flags
	Options QEFDecimatorOptions
}

func newQEFDecimator(mesh *Mesh, options QEFDecimatorOptions) *qefDecimator {
	m, ptrMap := ptrMeshAndMapping(mesh)
	res := &qefDecimator{
		Mesh:    m,
		PtrMap:  ptrMap,
		NumTris: mesh.NumTriangles(),
		QEFs:    map[*ptrCoord]numerical.QEF4{},
		Heap:    newQEFHeap(),
		Pairs:   map[*ptrCoord]map[*ptrCoord]struct{}{},
		Options: options,
	}

	allPairs := map[ptrSegment]struct{}{}
	ptrMap.Range(func(c Coord3D, pc *ptrCoord) bool {
		tris := pc.Triangles
		qef := numerical.QEF4{}
		if options.TikhonovRegularization > 0 {
			qef = *numerical.NewQEF4Dist(c.Array()).Scale(options.TikhonovRegularization)
		}
		for _, tri := range tris {
			rawTri := tri.Triangle()
			normal := rawTri.Normal()
			bias := -normal.Dot(rawTri[0])
			qef = *qef.Add(
				numerical.NewQEF4Outer(numerical.Vec4{normal.X, normal.Y, normal.Z, bias}),
			)
			for _, v1 := range tri.Coords {
				if pc != v1 {
					allPairs[newPtrSegment(pc, v1)] = struct{}{}
				}
			}
		}
		res.QEFs[pc] = qef
		return true
	})

	for pair := range allPairs {
		res.addPair(pair)
	}

	return res
}

func (q *qefDecimator) Step() bool {
	for {
		pair, solution, ok := q.Heap.Pop()
		if !ok {
			return false
		}
		newVPtr := q.mergeInMesh(pair, solution.Point)
		if newVPtr == nil {
			continue
		}
		q.mergeQEFAndHeap(pair, newVPtr)
		return true
	}
}

func (q *qefDecimator) mergeInMesh(pair ptrSegment, newV Coord3D) *ptrCoord {
	if newV != pair[0].Coord3D && newV != pair[1].Coord3D {
		if existingC, ok := q.PtrMap.Load(newV); ok && len(existingC.Triangles) > 0 {
			// No existing points can be used.
			return nil
		}
	}
	if !q.vertexLinkCondition(pair) || !q.noDuplicateTris(pair) {
		// Do not allow non-manifold results.
		return nil
	}
	newVPtr := q.PtrMap.Coord(newV)

	affected := map[*ptrTriangle]struct{}{}
	for _, v := range pair {
		for _, t := range v.Triangles {
			affected[t] = struct{}{}
		}
	}

	addTris := []*ptrTriangle{}
	for t := range affected {
		newVs := [3]*ptrCoord{t.Coords[0], t.Coords[1], t.Coords[2]}
		for i, c := range t.Coords {
			if c == pair[0] || c == pair[1] {
				newVs[i] = newVPtr
			}
		}
		if newVs[0] == newVs[1] || newVs[0] == newVs[2] || newVs[1] == newVs[2] {
			continue
		}
		pt := &ptrTriangle{Coords: newVs}
		normDot := pt.Triangle().Normal().Dot(t.Triangle().Normal())
		if math.IsNaN(normDot) || math.IsInf(normDot, 0) {
			// Singular triangle.
			return nil
		}
		if normDot < q.Options.MinNormalDotProduct {
			// Normal flip
			return nil
		}
		addTris = append(addTris, pt)
	}

	for t := range affected {
		t.RemoveCoords()
		q.Mesh.Remove(t)
		q.NumTris -= 1
	}

	for _, t := range addTris {
		t.AddCoords()
		q.Mesh.Add(t)
		q.NumTris += 1
	}

	return newVPtr
}

func (q *qefDecimator) vertexLinkCondition(pair ptrSegment) bool {
	neighborCounts := map[*ptrCoord]int{}
	for _, v := range pair {
		adjacentSet := map[*ptrCoord]struct{}{}
		for _, n := range v.Triangles {
			for _, c := range n.Coords {
				if c != v {
					adjacentSet[c] = struct{}{}
				}
			}
		}
		for c := range adjacentSet {
			neighborCounts[c] += 1
		}
	}
	var twiceCount int
	for _, n := range neighborCounts {
		if n == 2 {
			twiceCount += 1
		}
	}
	// Exactly two neighbors should be counted twice, since
	// there should be exactly two indicent triangles, and
	// no other common points.
	return twiceCount == 2
}

func (q *qefDecimator) noDuplicateTris(pair ptrSegment) bool {
	newTriCounts := map[ptrSegment]int{}
	for i, v := range pair {
		otherV := pair[1-i]
		for _, tri := range v.Triangles {
			v1, v2 := tri.Coords[0], tri.Coords[1]
			if v1 == v {
				v1 = tri.Coords[2]
			} else if v2 == v {
				v2 = tri.Coords[2]
			}
			if v1 == otherV || v2 == otherV {
				// This triangle would be collapsed.
				continue
			}
			newTriCounts[newPtrSegment(v1, v2)] += 1
		}
	}
	for _, n := range newTriCounts {
		if n > 1 {
			return false
		}
	}
	return true
}

func (q *qefDecimator) mergeQEFAndHeap(pair ptrSegment, newV *ptrCoord) {
	newQEF := numerical.QEF4{}
	for _, v := range pair {
		qef, ok := q.QEFs[v]
		if !ok {
			panic("no QEF for vertex")
		}
		delete(q.QEFs, v)
		newQEF = *newQEF.Add(&qef)
	}
	q.QEFs[newV] = newQEF

	newPairs := map[ptrSegment]struct{}{}
	for _, v := range pair {
		pairs, ok := q.Pairs[v]
		if !ok {
			panic("pairs must exist if we are working on one of them")
		}
		delete(q.Pairs, v)
		for other := range pairs {
			if other == pair[0] || other == pair[1] {
				// This is the pair we are currently merging
				continue
			}
			pairs, _ := q.Pairs[other]
			delete(pairs, v)

			newPair := newPtrSegment(newV, other)
			q.Heap.Remove(newPtrSegment(v, other))
			newPairs[newPair] = struct{}{}
		}
	}

	for newPair := range newPairs {
		q.addPair(newPair)
	}
}

func (q *qefDecimator) addPair(pair ptrSegment) {
	qef1, _ := q.QEFs[pair[0]]
	qef2, _ := q.QEFs[pair[1]]
	qef := qef1.Add(&qef2)
	solution := q.solve(qef, pair[0], pair[1])
	q.Heap.Add(pair, solution)
	for i, v := range pair {
		m, ok := q.Pairs[v]
		if !ok {
			m = map[*ptrCoord]struct{}{}
			q.Pairs[v] = m
		}
		m[pair[1-i]] = struct{}{}
	}
}

func (q *qefDecimator) solve(qef *numerical.QEF4, v1, v2 *ptrCoord) qefSolution {
	solutionV3, det := qef.Minimize()
	if det < q.Options.minDet() {
		bestCost := math.Inf(1)
		bestSolution := Origin
		for i, s := range []Coord3D{v1.Coord3D, v2.Coord3D, v1.Coord3D.Mid(v2.Coord3D)} {
			cost := qef.Eval(s.Array())
			if i == 0 || cost < bestCost {
				bestCost = cost
				bestSolution = s
			}
		}
		solutionV3 = bestSolution.Array()
	}
	solution := NewCoord3DArray(solutionV3)
	return qefSolution{
		Cost:  qef.Eval(solutionV3),
		Point: solution,
	}
}

type qefSolution struct {
	Cost  float64
	Point Coord3D
}

type qefHeapEntry struct {
	qefSolution
	Pair ptrSegment
}

type qefHeap struct {
	entries []qefHeapEntry
	idxs    map[ptrSegment]int
}

func newQEFHeap() *qefHeap {
	return &qefHeap{
		idxs: map[ptrSegment]int{},
	}
}

func (q *qefHeap) Add(pair ptrSegment, solution qefSolution) {
	if _, ok := q.idxs[pair]; ok {
		panic("cannot re-add an existing pair")
	}
	q.idxs[pair] = len(q.entries)
	q.entries = append(q.entries, qefHeapEntry{qefSolution: solution, Pair: pair})
	q.percolateUp(len(q.entries) - 1)
}

func (q *qefHeap) Remove(pair ptrSegment) bool {
	idx, ok := q.idxs[pair]
	if !ok {
		return false
	}
	q.removeIdx(idx)
	return true
}

func (q *qefHeap) removeIdx(idx int) {
	delete(q.idxs, q.entries[idx].Pair)
	lastEntry := q.entries[len(q.entries)-1]
	q.entries = q.entries[:len(q.entries)-1]

	if idx == len(q.entries) {
		// All we had to do was delete the last entry.
		return
	}

	q.entries[idx] = lastEntry
	q.idxs[lastEntry.Pair] = idx

	if idx != 0 {
		// It's possible we moved something into the slot that wasn't a
		// descendant of the slot, so we have to check both directions.
		q.percolateUp(idx)
	}
	q.percolateDown(idx)
}

func (q *qefHeap) Pop() (ptrSegment, qefSolution, bool) {
	if len(q.entries) == 0 {
		return ptrSegment{}, qefSolution{}, false
	}
	res := q.entries[0]
	q.removeIdx(0)
	return res.Pair, res.qefSolution, true
}

func (q *qefHeap) percolateUp(idx int) {
	if idx == 0 {
		return
	}
	parentIdx := q.parent(idx)
	if q.entries[parentIdx].Cost > q.entries[idx].Cost {
		q.swap(parentIdx, idx)
		q.percolateUp(parentIdx)
	}
}

func (q *qefHeap) percolateDown(idx int) {
	i1, i2 := q.children(idx)
	if i1 >= len(q.entries) {
		return
	} else if i2 >= len(q.entries) {
		if q.entries[i1].Cost < q.entries[idx].Cost {
			q.swap(i1, idx)
			q.percolateDown(i1)
		}
	} else {
		lowerChild := i1
		if q.entries[i2].Cost < q.entries[i1].Cost {
			lowerChild = i2
		}
		if q.entries[lowerChild].Cost < q.entries[idx].Cost {
			q.swap(lowerChild, idx)
			q.percolateDown(lowerChild)
		}
	}
}

func (q *qefHeap) parent(idx int) int {
	return (idx - 1) / 2
}

func (q *qefHeap) children(idx int) (int, int) {
	return idx*2 + 1, idx*2 + 2
}

func (q *qefHeap) swap(i, j int) {
	e1, e2 := q.entries[i], q.entries[j]
	q.idxs[e1.Pair] = j
	q.idxs[e2.Pair] = i
	q.entries[i], q.entries[j] = e2, e1
}
