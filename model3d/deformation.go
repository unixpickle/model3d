package model3d

import (
	"math"
)

// ARAPDefaultTolerance is the default convergence
// tolerance for ARAP.
const ARAPDefaultTolerance = 1e-2

const (
	arapMaxCGIterations = 2000
	arapMaxIterations   = 200
)

type ARAPWeightingScheme int

const (
	// ARAPWeightingCotangent is the default weighting scheme
	// for ARAP from the paper. Unfortunately, it creates a
	// loss function that can potentially become negative.
	ARAPWeightingCotangent ARAPWeightingScheme = iota

	ARAPWeightingAbsCotangent
	ARAPWeightingUniform
)

func (a ARAPWeightingScheme) weight(cot float64) float64 {
	switch a {
	case ARAPWeightingCotangent:
		return cot
	case ARAPWeightingAbsCotangent:
		return math.Abs(cot)
	case ARAPWeightingUniform:
		return 1
	default:
		panic("unknown weighting scheme")
	}
}

// ARAPConstraints maps coordinates from an original mesh
// to destination coordinates on a deformed mesh.
type ARAPConstraints map[Coord3D]Coord3D

// AddAround adds all of the points within r distance of c
// to the constraints, moving them such that c would move
// to target.
//
// Returns the number of points added.
func (a ARAPConstraints) AddAround(arap *ARAP, c Coord3D, r float64, target Coord3D) {
	offset := target.Sub(c)
	for _, c1 := range arap.coords {
		if c.Dist(c1) <= r {
			a[c1] = c1.Add(offset)
		}
	}
}

// ARAP implements as-rigid-as-possible deformations for a
// pre-determined mesh.
type ARAP struct {
	coordToIdx map[Coord3D]int
	coords     []Coord3D
	neighbors  [][]int
	weights    [][]float64
	rotWeights [][]float64
	triangles  [][3]int

	tolerance float64
}

// NewARAP creates an ARAP instance for the given mesh
// topology.
//
// The ARAP instance will not hold a reference to m or its
// triangles. Rather, it copies the data as needed.
//
// The instance uses cotangent weights, which are only
// guaranteed to work on meshes with smaller-than-right
// angles.
// For other weighting options, see NewARAPWeighted().
func NewARAP(m *Mesh) *ARAP {
	return NewARAPWeighted(m, ARAPWeightingCotangent, ARAPWeightingCotangent)
}

// NewARAPWeighted creates an ARAP with a specified
// weighting scheme.
//
// The linear weighting scheme is used for linear solves,
// whereas the rotation weighting scheme is used for
// finding rigid transformations.
//
// The ARAP instance will not hold a reference to m or its
// triangles. Rather, it copies the data as needed.
func NewARAPWeighted(m *Mesh, linear, rotation ARAPWeightingScheme) *ARAP {
	coords := m.VertexSlice()
	triangles := m.TriangleSlice()
	a := &ARAP{
		coordToIdx: map[Coord3D]int{},
		coords:     coords,
		neighbors:  make([][]int, len(coords)),
		weights:    make([][]float64, len(coords)),
		rotWeights: make([][]float64, len(coords)),
		triangles:  make([][3]int, 0, len(triangles)),
		tolerance:  ARAPDefaultTolerance,
	}

	for i, c := range coords {
		a.coordToIdx[c] = i
	}

	edgeToTri := map[arapEdge][]int{}
	m.Iterate(func(t *Triangle) {
		var tIdxs [3]int
		for i, c := range t {
			tIdxs[i] = a.coordToIdx[c]
		}
		triIdx := len(a.triangles)
		a.triangles = append(a.triangles, tIdxs)

		for i1, c1 := range tIdxs {
			for i2, c2 := range tIdxs {
				if i1 == i2 {
					continue
				}
				if i2 > i1 {
					e := newARAPEdge(c1, c2)
					edgeToTri[e] = append(edgeToTri[e], triIdx)
				}
				var found bool
				for _, n := range a.neighbors[c1] {
					if n == c2 {
						found = true
						break
					}
				}
				if !found {
					a.neighbors[c1] = append(a.neighbors[c1], c2)
				}
			}
		}
	})

	for c1, neighbors := range a.neighbors {
		var weights, rotWeights []float64
		for _, c2 := range neighbors {
			var cotangentSum float64
			for _, t := range edgeToTri[newARAPEdge(c1, c2)] {
				var otherCoord int
				for _, c3 := range a.triangles[t] {
					if c3 != c1 && c3 != c2 {
						otherCoord = c3
						break
					}
				}
				c3Point := a.coords[otherCoord]
				v1 := a.coords[c1].Sub(c3Point)
				v2 := a.coords[c2].Sub(c3Point)
				cosTheta := v1.Normalize().Dot(v2.Normalize())
				cotangentSum += cosTheta / math.Sqrt(math.Max(0, 1-cosTheta*cosTheta))
			}
			weights = append(weights, linear.weight(cotangentSum/2))
			rotWeights = append(rotWeights, rotation.weight(cotangentSum/2))
		}
		a.weights[c1] = weights
		a.rotWeights[c1] = rotWeights
	}

	return a
}

// Tolerance gets the current convergence tolerance.
// Will be ARAPDefaultTolerance by default.
func (a *ARAP) Tolerance() float64 {
	return a.tolerance
}

// SetTolerance changes the convergence tolerance.
// Lower values make the algorithm run longer but arrive
// at more accurate values.
//
// See ARAPDefaultTolerance.
func (a *ARAP) SetTolerance(t float64) {
	a.tolerance = t
}

// Deform creates a new mesh by enforcing constraints on
// some points of the mesh.
func (a *ARAP) Deform(constraints ARAPConstraints) *Mesh {
	mapping := a.deformMap(constraints, nil)
	res := NewMesh()
	for _, t := range a.triangles {
		res.Add(&Triangle{mapping[t[0]], mapping[t[1]], mapping[t[2]]})
	}
	return res
}

// Laplace deforms the mesh using a simple Laplacian
// heuristic.
//
// This can be used to generate an initial guess for the
// more general Deform() method.
//
// The result maps all old coordinates to new coordinates.
func (a *ARAP) Laplace(constraints ARAPConstraints) map[Coord3D]Coord3D {
	fullL := newARAPOperator(a, nil)
	targets := fullL.Apply(a.coords)

	l := newARAPOperator(a, a.indexConstraints(constraints))
	outs := l.LinSolve(targets, nil)
	res := make(map[Coord3D]Coord3D, len(a.coords))
	for i, c := range a.coords {
		res[c] = outs[i]
	}
	return res
}

// DeformMap performs constrained mesh deformation.
//
// The constraints argument maps coordinates from the
// original mesh to their new, fixed locations.
//
// If the initialGuess is specified, it is used for the
// first iteration of the algorithm as a starting point
// for the deformation.
//
// The result maps all old coordinates to new coordinates.
func (a *ARAP) DeformMap(constraints ARAPConstraints,
	initialGuess map[Coord3D]Coord3D) map[Coord3D]Coord3D {
	currentOutput := a.deformMap(constraints, initialGuess)
	res := make(map[Coord3D]Coord3D, len(a.coords))
	for i, c := range a.coords {
		res[c] = currentOutput[i]
	}
	return res
}

func (a *ARAP) deformMap(constraints, initialGuess map[Coord3D]Coord3D) []Coord3D {
	if initialGuess == nil {
		initialGuess = a.Laplace(constraints)
	}

	l := newARAPOperator(a, a.indexConstraints(constraints))

	currentOutput := make([]Coord3D, len(a.coords))
	for i, c := range a.coords {
		currentOutput[i] = initialGuess[c]
	}

	lastEnergy := a.energy(currentOutput)
	for iter := 0; iter < arapMaxIterations; iter++ {
		rotations := a.rotations(currentOutput)
		targets := l.Targets(rotations)
		currentOutput = l.LinSolve(targets, currentOutput)
		energy := a.energy(currentOutput)
		if 1-energy/lastEnergy < a.tolerance {
			break
		}
		lastEnergy = energy
	}

	return currentOutput
}

// rotations computes the rotations-of-best-fit for the
// current coordinate positions.
func (a *ARAP) rotations(currentOutput []Coord3D) []Matrix3 {
	rotations := make([]Matrix3, len(a.coords))
	for i, c := range a.coords {
		var covariance Matrix3
		for j, n := range a.neighbors[i] {
			weight := a.rotWeights[i][j]
			origDiff := a.coords[n].Sub(c)
			newDiff := currentOutput[n].Sub(currentOutput[i])
			piece := NewMatrix3Columns(
				origDiff.Scale(newDiff.X),
				origDiff.Scale(newDiff.Y),
				origDiff.Scale(newDiff.Z),
			)
			for i, x := range piece {
				covariance[i] += x * weight
			}
		}
		var u, s, v Matrix3
		covariance.SVD(&u, &s, &v)
		rot := *v.Mul(u.Transpose())
		if rot.Det() < 0 {
			var smallestIndex int
			smallestValue := s[0]
			for i, s1 := range []float64{s[4], s[8]} {
				if s1 < smallestValue {
					smallestIndex = i + 1
					smallestValue = s1
				}
			}
			v[smallestIndex] *= -1
			v[smallestIndex+3] *= -1
			v[smallestIndex+6] *= -1
			rot = *v.Mul(u.Transpose())
		}
		rotations[i] = rot
	}
	return rotations
}

// energy computes the ARAP loss energy.
// This can be used for debugging, and perhaps in the
// future for convergence analysis.
func (a *ARAP) energy(currentOutput []Coord3D) float64 {
	rotations := a.rotations(currentOutput)
	var energy float64
	for i, neighbors := range a.neighbors {
		rotation := rotations[i]
		for j, n := range neighbors {
			w := a.weights[i][j]
			rotated := rotation.MulColumn(a.coords[i].Sub(a.coords[n]))
			diff := currentOutput[i].Sub(currentOutput[n]).Sub(rotated)
			energy += w * diff.Dot(diff)
		}
	}
	return energy
}

// indexConstraints converts the keys to indices.
func (a *ARAP) indexConstraints(constraints map[Coord3D]Coord3D) map[int]Coord3D {
	res := map[int]Coord3D{}
	for in, out := range constraints {
		if idx, ok := a.coordToIdx[in]; !ok {
			panic("constraint source was not in the original mesh")
		} else {
			res[idx] = out
		}
	}
	return res
}

// arapOperator implements the Laplace-Beltrami matrix.
//
// By default, it applies the entire matrix.
// However, it also allows for constrained vertices to be
// substituted for their exact values.
type arapOperator struct {
	arap        *ARAP
	constraints map[int]Coord3D

	// Mapping from constrained (reduced) coordinates to
	// full coordinate indices.
	squeezedToFull []int

	// Inverse of squeezedToFull with -1 at constraints.
	fullToSqueezed []int

	precond *arapPrecond
}

func newARAPOperator(a *ARAP, constraints map[int]Coord3D) *arapOperator {
	if constraints == nil {
		constraints = map[int]Coord3D{}
	}
	squeezedToFull := make([]int, 0, len(a.coords)-len(constraints))
	fullToSqueezed := make([]int, len(a.coords))
	for i := 0; i < len(a.coords); i++ {
		if _, ok := constraints[i]; !ok {
			fullToSqueezed[i] = len(squeezedToFull)
			squeezedToFull = append(squeezedToFull, i)
		} else {
			fullToSqueezed[i] = -1
		}
	}

	res := &arapOperator{
		arap:           a,
		constraints:    constraints,
		squeezedToFull: squeezedToFull,
		fullToSqueezed: fullToSqueezed,
	}
	res.precond = newARAPPrecond(res)
	return res
}

// LinSolve performs a linear solve for x in Lx=b.
// It is assumed that b and x are unsqueezed (full rank),
// and the constrained rows of b are simply ignored.
func (a *arapOperator) LinSolve(b, start []Coord3D) []Coord3D {
	if len(a.squeezedToFull) == 0 {
		// All points are constrained.
		return b
	}

	if start == nil {
		start = a.arap.coords
	}

	b = a.Squeeze(b)
	for i, c := range a.SqueezeDelta() {
		b[i] = b[i].Add(c)
	}

	preventZeros := func(c Coord3D) Coord3D {
		arr := c.Array()
		for i, x := range arr {
			if x == 0 {
				arr[i] = math.Nextafter(0, 1)
			}
		}
		return NewCoord3DArray(arr)
	}

	x := a.Squeeze(start)
	r := arapCopy(b)
	arapSub(r, a.Apply(x))
	z := make([]Coord3D, len(r))
	a.precond.ApplyInverse(z, r)
	p := arapCopy(z)
	eps := arapDot(b, b).Scale(1e-8)
	rMag := arapDot(r, r)
	rzDot := arapDot(r, z)

	for i := 0; i < arapMaxCGIterations && rMag.Max(eps) != eps && rMag.Sum() != 0; i++ {
		ap := a.Apply(p)

		alpha := rzDot.Div(preventZeros(arapDot(p, ap)))
		arapAddScaled(x, p, alpha)

		if (i+1)%50 == 0 {
			// Use explicit update for r to avoid compounding
			// error over many updates.
			copy(r, b)
			arapSub(r, a.Apply(x))
		} else {
			arapAddScaled(r, ap, alpha.Scale(-1))
		}
		a.precond.ApplyInverse(z, r)
		rMag = arapDot(r, r)
		nextRZDot := arapDot(r, z)

		beta := nextRZDot.Div(preventZeros(rzDot))
		for i, c := range z {
			p[i] = c.Add(p[i].Mul(beta))
		}
		rzDot = nextRZDot
	}

	return a.Unsqueeze(x)
}

// Squeeze gets a vector that can be put through the
// operator (i.e. that has constraints removed).
func (a *arapOperator) Squeeze(full []Coord3D) []Coord3D {
	result := make([]Coord3D, len(a.squeezedToFull))
	for i, j := range a.squeezedToFull {
		result[i] = full[j]
	}
	return result
}

// Unsqueeze performs the inverse of squeeze, filling in
// the constrained values as needed.
func (a *arapOperator) Unsqueeze(squeezed []Coord3D) []Coord3D {
	res := make([]Coord3D, len(a.arap.coords))
	for i, s := range a.fullToSqueezed {
		if s != -1 {
			res[i] = squeezed[s]
		} else {
			res[i] = a.constraints[i]
		}
	}
	return res
}

// SqueezeDelta gets the change in the un-constrained
// variables caused by squeezing out the constraints.
//
// This should be added to the other side of linear
// systems to find the correct values.
func (a *arapOperator) SqueezeDelta() []Coord3D {
	res := make([]Coord3D, len(a.squeezedToFull))
	for i, fullIdx := range a.squeezedToFull {
		neighbors := a.arap.neighbors[fullIdx]
		weights := a.arap.weights[fullIdx]
		var result Coord3D
		for j, n := range neighbors {
			w := weights[j]
			if nSqueezed := a.fullToSqueezed[n]; nSqueezed == -1 {
				result = result.Add(a.constraints[n].Scale(w))
			}
		}
		res[i] = result
	}
	return res
}

// Apply applies the Laplace-Beltrami operator to the
// squeezed vector to get another squeezed vector.
func (a *arapOperator) Apply(v []Coord3D) []Coord3D {
	res := make([]Coord3D, len(v))
	for i, fullIdx := range a.squeezedToFull {
		p := v[i]
		neighbors := a.arap.neighbors[fullIdx]
		weights := a.arap.weights[fullIdx]
		var result Coord3D
		for j, n := range neighbors {
			w := weights[j]
			result = result.Add(p.Scale(w))
			if nSqueezed := a.fullToSqueezed[n]; nSqueezed != -1 {
				result = result.Sub(v[nSqueezed].Scale(w))
			}
		}
		res[i] = result
	}
	return res
}

// Targets computes the right-hand side of the Poisson
// equation using rotation matrices.
func (a *arapOperator) Targets(rotations []Matrix3) []Coord3D {
	res := make([]Coord3D, len(a.arap.coords))
	for i, p := range a.arap.coords {
		neighbors := a.arap.neighbors[i]
		weights := a.arap.weights[i]
		var result Coord3D
		for j, n := range neighbors {
			var rotation Matrix3
			m1 := rotations[i]
			m2 := rotations[n]
			for i, x := range m1 {
				rotation[i] = x + m2[i]
			}
			w := weights[j] / 2
			diff := p.Sub(a.arap.coords[n]).Scale(w)
			result = result.Add(rotation.MulColumn(diff))
		}
		res[i] = result
	}
	return res
}

// arapPrecond is a preconditioner for an arapOperator.
type arapPrecond struct {
	lower *arapSparse
	upper *arapSparse
}

// newARAPPrecond creates a new preconditioning operator.
func newARAPPrecond(a *arapOperator) *arapPrecond {
	size := len(a.squeezedToFull)
	lower := newARAPSparse(size)
	upper := newARAPSparse(size)

	// https://en.wikipedia.org/wiki/Incomplete_Cholesky_factorization

	diagonal := make([]float64, size)
	for i, full := range a.squeezedToFull {
		var sum float64
		for j, w := range a.arap.weights[full] {
			if a.fullToSqueezed[a.arap.neighbors[full][j]] != -1 {
				sum += w
			}
		}
		diagonal[i] = sum
	}

	for i, full := range a.squeezedToFull {
		diagonalEntry := diagonal[i]
		lower.Iterate(i, func(col int, x float64) {
			diagonalEntry -= x * x
		})
		// TODO: see if we need to make sure the diagonal entry
		// does not equal zero.
		diagonalEntry = math.Sqrt(diagonalEntry)
		lower.Set(i, i, diagonalEntry)
		upper.Set(i, i, diagonalEntry)

		for neighborIdx, neighborFull := range a.arap.neighbors[full] {
			j := a.fullToSqueezed[neighborFull]
			if j <= i {
				continue
			}
			entry := -a.arap.weights[full][neighborIdx]

			lower.Iterate(i, func(k int, x float64) {
				if k >= i {
					return
				}
				lower.Iterate(j, func(k1 int, y float64) {
					if k == k1 {
						entry -= x * y
					}
				})
			})

			entry = entry / diagonalEntry
			lower.Set(j, i, entry)
			upper.Set(i, j, entry)
		}
	}

	return &arapPrecond{
		lower: lower,
		upper: upper,
	}
}

// ApplyInverse computes P^-1 * b, where P is the matrix
// for the preconditioner.
//
// The output is written to out.
func (a *arapPrecond) ApplyInverse(out, b []Coord3D) {
	a.lower.BacksubLower(out, b)
	a.upper.BacksubUpper(out, out)
}

// Apply computes P*b, where P is the matrix for the
// preconditioner.
//
// The output is written to out.
func (a *arapPrecond) Apply(out, b []Coord3D) {
	for i := range out {
		var sum Coord3D
		a.upper.Iterate(i, func(col int, x float64) {
			sum = sum.Add(b[col].Scale(x))
		})
		out[i] = sum
	}
	for i := len(out) - 1; i >= 0; i-- {
		var sum Coord3D
		a.lower.Iterate(i, func(col int, x float64) {
			sum = sum.Add(out[col].Scale(x))
		})
		out[i] = sum
	}
}

type arapSparse struct {
	rows    [][]float64
	indices [][]int
}

func newARAPSparse(size int) *arapSparse {
	return &arapSparse{
		rows:    make([][]float64, size),
		indices: make([][]int, size),
	}
}

func (a *arapSparse) Set(row, col int, x float64) {
	a.rows[row] = append(a.rows[row], x)
	a.indices[row] = append(a.indices[row], col)
}

func (a *arapSparse) Iterate(row int, f func(col int, x float64)) {
	for i, col := range a.indices[row] {
		f(col, a.rows[row][i])
	}
}

func (a *arapSparse) BacksubUpper(out, b []Coord3D) {
	for row := len(b) - 1; row >= 0; row-- {
		bValue := b[row]
		var diagValue float64
		a.Iterate(row, func(col int, x float64) {
			if col < row {
				panic("not upper-diagonal")
			} else if col == row {
				diagValue = x
			} else {
				bValue = bValue.Add(out[col].Scale(-x))
			}
		})
		out[row] = bValue.Scale(1 / diagValue)
	}
}

func (a *arapSparse) BacksubLower(out, b []Coord3D) {
	for row, bValue := range b {
		var diagValue float64
		a.Iterate(row, func(col int, x float64) {
			if col > row {
				panic("not lower-diagonal")
			} else if col == row {
				diagValue = x
			} else {
				bValue = bValue.Add(out[col].Scale(-x))
			}
		})
		out[row] = bValue.Scale(1 / diagValue)
	}
}

type arapEdge [2]int

func newARAPEdge(i1, i2 int) arapEdge {
	if i1 < i2 {
		return arapEdge{i1, i2}
	} else {
		return arapEdge{i2, i1}
	}
}

func arapCopy(v []Coord3D) []Coord3D {
	return append([]Coord3D{}, v...)
}

func arapAdd(v1, v2 []Coord3D) {
	if len(v1) != len(v2) {
		panic("length mismatch")
	}
	for i, x := range v1 {
		v1[i] = x.Add(v2[i])
	}
}

func arapAddScaled(v1, v2 []Coord3D, s Coord3D) {
	if len(v1) != len(v2) {
		panic("length mismatch")
	}
	for i, x := range v1 {
		v1[i] = x.Add(v2[i].Mul(s))
	}
}

func arapSub(v1, v2 []Coord3D) {
	if len(v1) != len(v2) {
		panic("length mismatch")
	}
	for i, x := range v1 {
		v1[i] = x.Sub(v2[i])
	}
}

func arapDot(v1, v2 []Coord3D) Coord3D {
	if len(v1) != len(v2) {
		panic("length mismatch")
	}
	var res Coord3D
	for i, x := range v1 {
		res = res.Add(x.Mul(v2[i]))
	}
	return res
}
