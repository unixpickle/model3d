package model3d

import (
	"math"

	"github.com/unixpickle/model3d/numerical"
)

const (
	QEFDecimatorDefaultMinDet = 1e-8
)

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
func QEFDecimate(m *Mesh, minTris int, options *QEFDecimatorOptions) {
	if options == nil {
		options = &QEFDecimatorOptions{}
	}
	dec := newQEFDecimator(m, *options)
	for m.NumTriangles() > minTris {
		if !dec.Step() {
			break
		}
	}
}

type qefDecimator struct {
	Mesh  *Mesh
	QEFs  *CoordMap[numerical.QEF4]
	Heap  *qefHeap
	Pairs *CoordMap[*CoordMap[struct{}]]

	// Configuration flags
	Options QEFDecimatorOptions
}

func newQEFDecimator(m *Mesh, options QEFDecimatorOptions) *qefDecimator {
	res := &qefDecimator{
		Mesh:    m,
		QEFs:    NewCoordMap[numerical.QEF4](),
		Heap:    newQEFHeap(),
		Pairs:   NewCoordMap[*CoordMap[struct{}]](),
		Options: options,
	}

	allPairs := map[Segment]struct{}{}
	for _, v := range m.VertexSlice() {
		tris := m.Find(v)
		qef := numerical.QEF4{}
		if options.TikhonovRegularization > 0 {
			qef = *numerical.NewQEF4Dist(v.Array()).Scale(options.TikhonovRegularization)
		}
		for _, tri := range tris {
			normal := tri.Normal()
			bias := -normal.Dot(tri[0])
			qef = *qef.Add(
				numerical.NewQEF4Outer(numerical.Vec4{normal.X, normal.Y, normal.Z, bias}),
			)
			for _, v1 := range tri {
				if v != v1 {
					allPairs[NewSegment(v, v1)] = struct{}{}
				}
			}
		}
		res.QEFs.Store(v, qef)
	}

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
		if !q.mergeInMesh(pair, solution.Point) {
			continue
		}
		q.mergeQEFAndHeap(pair, solution.Point)
		return true
	}
}

func (q *qefDecimator) mergeInMesh(pair Segment, newV Coord3D) bool {
	if newV != pair[0] && newV != pair[1] && len(q.Mesh.Find(newV)) > 0 {
		// No existing points can be used.
		return false
	}
	if !q.vertexLinkCondition(pair) || !q.noDuplicateTris(pair) {
		// Do not allow non-manifold results.
		return false
	}

	affected := map[*Triangle]struct{}{}
	for _, v := range pair {
		for _, t := range q.Mesh.Find(v) {
			affected[t] = struct{}{}
		}
	}

	var newTris []*Triangle
	for t := range affected {
		newT := &Triangle{t[0], t[1], t[2]}
		for i, c := range t {
			if c == pair[0] {
				newT[i] = newV
			} else if c == pair[1] {
				newT[i] = newV
			}
		}
		if newT[0] == newT[1] || newT[0] == newT[2] || newT[1] == newT[2] {
			continue
		}
		if newT.Normal().Dot(t.Normal()) < q.Options.MinNormalDotProduct {
			// Normal flip
			return false
		}
		newTris = append(newTris, newT)
	}

	for t := range affected {
		q.Mesh.Remove(t)
	}
	for _, t := range newTris {
		q.Mesh.Add(t)
	}

	return true
}

func (q *qefDecimator) vertexLinkCondition(pair Segment) bool {
	neighborCounts := NewCoordToNumber[int]()
	for _, v := range pair {
		adjacentSet := NewCoordMap[struct{}]()
		for _, n := range q.Mesh.Find(v) {
			for _, c := range n {
				if c != v {
					adjacentSet.Store(c, struct{}{})
				}
			}
		}
		adjacentSet.KeyRange(func(c Coord3D) bool {
			neighborCounts.Add(c, 1)
			return true
		})
	}
	var twiceCount int
	neighborCounts.ValueRange(func(n int) bool {
		if n == 2 {
			twiceCount += 1
		}
		return true
	})
	// Exactly two neighbors should be counted twice, since
	// there should be exactly two indicent triangles, and
	// no other common points.
	return twiceCount == 2
}

func (q *qefDecimator) noDuplicateTris(pair Segment) bool {
	newTriCounts := map[Segment]int{}
	for i, v := range pair {
		otherV := pair[1-i]
		for _, tri := range q.Mesh.Find(v) {
			seg := tri.otherSegment(v)
			if seg[0] == otherV || seg[1] == otherV {
				// This triangle would be collapsed.
				continue
			}
			newTriCounts[seg] += 1
		}
	}
	for _, n := range newTriCounts {
		if n > 1 {
			return false
		}
	}
	return true
}

func (q *qefDecimator) mergeQEFAndHeap(pair Segment, newV Coord3D) {
	if _, ok := q.QEFs.Load(newV); ok {
		panic("cannot merge into a vertex that already exists")
	}

	newQEF := numerical.QEF4{}
	for _, v := range pair {
		qef, ok := q.QEFs.Load(v)
		if !ok {
			panic("no QEF for vertex")
		}
		q.QEFs.Delete(v)
		newQEF = *newQEF.Add(&qef)
	}
	q.QEFs.Store(newV, newQEF)

	newPairs := map[Segment]struct{}{}
	for _, v := range pair {
		pairs, ok := q.Pairs.Load(v)
		if !ok {
			panic("pairs must exist if we are working on one of them")
		}
		q.Pairs.Delete(v)
		pairs.KeyRange(func(other Coord3D) bool {
			if other == pair[0] || other == pair[1] {
				// This is the pair we are currently merging
				return true
			}
			pairs, _ := q.Pairs.Load(other)
			pairs.Delete(v)

			newPair := NewSegment(newV, other)
			q.Heap.Remove(NewSegment(v, other))
			newPairs[newPair] = struct{}{}
			return true
		})
	}

	for newPair := range newPairs {
		q.addPair(newPair)
	}
}

func (q *qefDecimator) addPair(pair Segment) {
	qef1, _ := q.QEFs.Load(pair[0])
	qef2, _ := q.QEFs.Load(pair[1])
	qef := qef1.Add(&qef2)
	solution := q.solve(qef, pair[0], pair[1])
	q.Heap.Add(pair, solution)
	for i, v := range pair {
		m, ok := q.Pairs.Load(v)
		if !ok {
			m = NewCoordMap[struct{}]()
			q.Pairs.Store(v, m)
		}
		m.Store(pair[1-i], struct{}{})
	}
}

func (q *qefDecimator) solve(qef *numerical.QEF4, v1, v2 Coord3D) qefSolution {
	solutionV3, det := qef.Minimize()
	if det < q.Options.minDet() {
		bestCost := math.Inf(1)
		bestSolution := Origin
		for i, s := range []Coord3D{v1, v2, v1.Mid(v2)} {
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
	Pair Segment
}

type qefHeap struct {
	entries []qefHeapEntry
	idxs    map[Segment]int
}

func newQEFHeap() *qefHeap {
	return &qefHeap{
		idxs: map[Segment]int{},
	}
}

func (q *qefHeap) Add(pair Segment, solution qefSolution) {
	if _, ok := q.idxs[pair]; ok {
		panic("cannot re-add an existing pair")
	}
	q.idxs[pair] = len(q.entries)
	q.entries = append(q.entries, qefHeapEntry{qefSolution: solution, Pair: pair})
	q.percolateUp(len(q.entries) - 1)
}

func (q *qefHeap) Remove(pair Segment) bool {
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

func (q *qefHeap) Pop() (Segment, qefSolution, bool) {
	if len(q.entries) == 0 {
		return Segment{}, qefSolution{}, false
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
