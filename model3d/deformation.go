package model3d

import (
	"math"
)

const (
	arapMaxCGIterations = 2000
	arapMaxIterations   = 20
)

// ARAP implements as-rigid-as-possible deformations for a
// pre-determined mesh.
type ARAP struct {
	coordToIdx map[Coord3D]int
	coords     []Coord3D
	neighbors  [][]int
	weights    [][]float64
	triangles  [][3]int
}

// NewARAP creates an ARAP instance for the given mesh
// topology.
//
// The ARAP instance will not hold a reference to m or its
// triangles. Rather, it copies the data as needed.
func NewARAP(m *Mesh) *ARAP {
	coords := m.VertexSlice()
	triangles := m.TriangleSlice()
	a := &ARAP{
		coordToIdx: map[Coord3D]int{},
		coords:     coords,
		neighbors:  make([][]int, len(coords)),
		weights:    make([][]float64, len(coords)),
		triangles:  make([][3]int, 0, len(triangles)),
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
		var weights []float64
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
			weights = append(weights, cotangentSum)
		}
		a.weights[c1] = weights
	}

	return a
}

// Deform creates a new mesh by enforcing constraints on
// some points of the mesh.
func (a *ARAP) Deform(constraints map[Coord3D]Coord3D) *Mesh {
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
// The constraints argument maps coordinates from the
// original mesh to their new, fixed locations.
//
// The result maps all old coordinates to new coordinates.
func (a *ARAP) Laplace(constraints map[Coord3D]Coord3D) map[Coord3D]Coord3D {
	fullL := newARAPOperator(a, nil)
	targets := fullL.Apply(a.coords)

	l := newARAPOperator(a, a.indexConstraints(constraints))
	outs := l.LinSolve(targets)
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
func (a *ARAP) DeformMap(constraints, initialGuess map[Coord3D]Coord3D) map[Coord3D]Coord3D {
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

	for iter := 0; iter < arapMaxIterations; iter++ {
		// Step 1: find nearest rigid deformations.
		rotations := make([]Matrix3, len(a.coords))
		for i, c := range a.coords {
			var covariance Matrix3
			for j, n := range a.neighbors[i] {
				weight := a.weights[i][j]
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
			rotations[i] = *v.Mul(u.Transpose())
		}

		// Step 2: solve for new points.
		targets := l.Targets(rotations)
		currentOutput = l.LinSolve(targets)
	}

	return currentOutput
}

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
	return &arapOperator{
		arap:           a,
		constraints:    constraints,
		squeezedToFull: squeezedToFull,
		fullToSqueezed: fullToSqueezed,
	}
}

// LinSolve performs a linear solve for x in Lx=b.
// It is assumed that b and x are unsqueezed (full rank),
// and the constrained rows of b are simply ignored.
func (a *arapOperator) LinSolve(b []Coord3D) []Coord3D {
	if len(a.squeezedToFull) == 0 {
		// All points are constrained.
		return b
	}
	b = a.Squeeze(b)
	for i, c := range a.SqueezeDelta() {
		b[i] = b[i].Add(c)
	}

	// Conjugate gradient algorithm with initial solution
	// at the original coordinates.
	x := a.Squeeze(a.arap.coords)
	r := arapSub(b, a.Apply(x))
	p := r

	for i := 0; i < arapMaxCGIterations; i++ {
		rMag := arapDot(r, r)
		// TODO: more clever convergence check based on
		// residual magnitude.
		if rMag.Sum() == 0 {
			break
		}
		rMag = rMag.Max(Coord3D{X: 1, Y: 1, Z: 1}.Scale(rMag.Sum() * 1e-18))

		ap := a.Apply(p)

		// Technically this should work, but it seems to
		// cause numerical instabilities.
		// Perhaps there is a bug in the matrix, i.e. it
		// is not positive definite?
		//
		//     alpha := rMag.Div(arapDot(p, ap))
		//
		alpha := arapDot(p, r).Div(arapDot(p, ap))
		nextX := arapAdd(x, arapScale(p, alpha))

		// Use explicit update for r to avoid compounding
		// error over many updates.
		nextR := arapSub(b, a.Apply(nextX))
		nextRMag := arapDot(nextR, nextR)

		beta := nextRMag.Div(rMag)
		nextP := arapAdd(nextR, arapScale(p, beta))
		x, r, p = nextX, nextR, nextP
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

type arapEdge [2]int

func newARAPEdge(i1, i2 int) arapEdge {
	if i1 < i2 {
		return arapEdge{i1, i2}
	} else {
		return arapEdge{i2, i1}
	}
}

func arapAdd(v1, v2 []Coord3D) []Coord3D {
	res := make([]Coord3D, len(v1))
	for i, x := range v1 {
		res[i] = x.Add(v2[i])
	}
	return res
}

func arapSub(v1, v2 []Coord3D) []Coord3D {
	res := make([]Coord3D, len(v1))
	for i, x := range v1 {
		res[i] = x.Sub(v2[i])
	}
	return res
}

func arapDot(v1, v2 []Coord3D) Coord3D {
	var res Coord3D
	for i, x := range v1 {
		res = res.Add(x.Mul(v2[i]))
	}
	return res
}

func arapScale(v []Coord3D, s Coord3D) []Coord3D {
	res := make([]Coord3D, len(v))
	for i, x := range v {
		res[i] = x.Mul(s)
	}
	return res
}
