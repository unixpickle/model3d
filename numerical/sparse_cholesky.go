package numerical

import (
	"math"
	"sort"

	"github.com/unixpickle/essentials"
)

// SparseCholesky is a sparse LU decomposition of a
// symmetric matrix.
//
// Once instantiated, this object can be used to quickly
// apply the inverse of a matrix to vectors.
type SparseCholesky struct {
	lower *SparseMatrix
	upper *SparseMatrix
	perm  []int
}

func NewSparseCholesky(mat *SparseMatrix) *SparseCholesky {
	perm := mat.RCM()
	mat = mat.Permute(perm)
	size := len(mat.rows)

	lower := NewSparseMatrix(size)
	upper := NewSparseMatrix(size)

	diagonal := make([]float64, size)
	for row := 0; row < size; row++ {
		mat.Iterate(row, func(col int, x float64) {
			if col == row {
				diagonal[row] = x
			}
		})
	}

	belowCache := newSparseMatrixMap(size)

	for i := 0; i < size; i++ {
		diagonalEntry := diagonal[i]
		lower.Iterate(i, func(col int, x float64) {
			diagonalEntry -= x * x
		})
		// TODO: see if we need to make sure the diagonal entry
		// does not equal zero.
		diagonalEntry = math.Sqrt(diagonalEntry)
		lower.Set(i, i, diagonalEntry)
		upper.Set(i, i, diagonalEntry)

		belowCache.Clear()

		mat.Iterate(i, func(j int, x float64) {
			if j > i {
				belowCache.Add(j, x)
			}
		})

		lower.Iterate(i, func(k int, x float64) {
			if k >= i || x == 0 {
				return
			}
			// The entries in lower and upper are sorted because
			// we added each entry in order.
			for entryIdx := len(upper.rows[k]) - 1; entryIdx >= 0; entryIdx-- {
				j := upper.indices[k][entryIdx]
				if j <= i {
					break
				}
				y := upper.rows[k][entryIdx]
				if y != 0 {
					belowCache.Add(j, -x*y)
				}
			}
		})

		// Intentionally adding the entries in order.
		s := 1 / diagonalEntry
		belowCache.IterateSorted(func(j int, v float64) {
			x := v * s
			lower.Set(j, i, x)
			upper.Set(i, j, x)
		})
	}

	return &SparseCholesky{
		lower: lower,
		upper: upper,
		perm:  perm,
	}
}

// ApplyVec3 computes (A*x, A*y, A*z).
func (a *SparseCholesky) ApplyVec3(x []Vec3) []Vec3 {
	out := make([]Vec3, len(x))
	b := permuteVectors(x, a.perm)
	for i := range out {
		var sum Vec3
		a.upper.Iterate(i, func(col int, x float64) {
			sum = sum.Add(b[col].Scale(x))
		})
		out[i] = sum
	}
	for i := len(out) - 1; i >= 0; i-- {
		var sum Vec3
		a.lower.Iterate(i, func(col int, x float64) {
			sum = sum.Add(out[col].Scale(x))
		})
		out[i] = sum
	}
	return permuteVectorsInv(out, a.perm)
}

// ApplyInverseVec3 computes (A^-1*x, A^-1*y, A^-1*z).
func (a *SparseCholesky) ApplyInverseVec3(x []Vec3) []Vec3 {
	b := permuteVectors(x, a.perm)
	out := make([]Vec3, len(x))
	a.lower.backsubLowerVec3(out, b)
	a.upper.backsubUpperVec3(out, out)
	return permuteVectorsInv(out, a.perm)
}

// A SparseMatrix is a square matrix where entries can be
// set to non-zero values in any order, and not all
// entries must be set.
type SparseMatrix struct {
	rows    [][]float64
	indices [][]int
}

func NewSparseMatrix(size int) *SparseMatrix {
	return &SparseMatrix{
		rows:    make([][]float64, size),
		indices: make([][]int, size),
	}
}

// Set adds an entry to the matrix.
//
// The entry should not already be set.
func (a *SparseMatrix) Set(row, col int, x float64) {
	a.rows[row] = append(a.rows[row], x)
	a.indices[row] = append(a.indices[row], col)
}

// Iterate loops through the non-zero entries in a row.
func (a *SparseMatrix) Iterate(row int, f func(col int, x float64)) {
	for i, col := range a.indices[row] {
		f(col, a.rows[row][i])
	}
}

// Permute permutes the rows and columns by perm, where
// perm is the result of applying the permutation to the
// list [0...n-1].
func (a *SparseMatrix) Permute(perm []int) *SparseMatrix {
	permInv := make([]int, len(perm))
	for i, j := range perm {
		permInv[j] = i
	}
	res := NewSparseMatrix(len(perm))
	for i, j := range perm {
		oldRow := a.indices[j]
		newRow := make([]int, 0, len(oldRow))
		for _, k := range oldRow {
			newRow = append(newRow, permInv[k])
		}
		res.indices[i] = newRow
		res.rows[i] = append([]float64{}, a.rows[j]...)
	}
	return res
}

// RCM computes the reverse Cuthill-McKee permutation for
// the matrix.
func (a *SparseMatrix) RCM() []int {
	remaining := map[int]bool{}
	for i := range a.indices {
		remaining[i] = true
	}

	remainingNeighbors := func(i int) int {
		var count int
		for _, neighbor := range a.indices[i] {
			if remaining[neighbor] {
				count++
			}
		}
		return count
	}

	drawBestStart := func() int {
		result := -1
		var resultNeighbors int
		for i := range remaining {
			n := remainingNeighbors(i)
			if n < resultNeighbors || result == -1 {
				result = i
				resultNeighbors = n
			}
		}
		return result
	}

	permutation := make([]int, 0, len(a.indices))
	for i := range a.indices {
		var expand int
		if i >= len(permutation) {
			expand = drawBestStart()
			permutation = append(permutation, expand)
			delete(remaining, expand)
		} else {
			expand = permutation[i]
		}

		allAdj := a.indices[expand]
		neighbors := make([]int, 0, len(allAdj))
		neighborOrder := make([]int, 0, len(allAdj))
		for _, j := range allAdj {
			if !remaining[j] {
				continue
			}
			neighbors = append(neighbors, j)
			neighborOrder = append(neighborOrder, remainingNeighbors(j))
		}
		essentials.VoodooSort(neighborOrder, func(i, j int) bool {
			return neighborOrder[i] < neighborOrder[j]
		}, neighbors)
		for _, n := range neighbors {
			permutation = append(permutation, n)
			delete(remaining, n)
		}
	}

	for i := 0; i < len(permutation)/2; i++ {
		permutation[i], permutation[len(permutation)-1] = permutation[len(permutation)-1], permutation[i]
	}

	return permutation
}

// ApplyVec3 computes (A*x, A*y, A*z).
func (a *SparseMatrix) ApplyVec3(x []Vec3) []Vec3 {
	res := make([]Vec3, len(x))
	for row, indices := range a.indices {
		for col, value := range a.rows[row] {
			res[row] = res[row].Add(x[indices[col]].Scale(value))
		}
	}
	return res
}

// backsubUpperVec3 writes U^-1*b to out, assuming this is
// an upper-triangular matrix U.
func (a *SparseMatrix) backsubUpperVec3(out, b []Vec3) {
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

// backsubLowerVec3 writes L^-1*b to out, assuming this is
// a lower-triangular matrix L.
func (a *SparseMatrix) backsubLowerVec3(out, b []Vec3) {
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

type sparseMatrixMap struct {
	data      []float64
	contained []bool
	indices   []int
}

func newSparseMatrixMap(size int) *sparseMatrixMap {
	return &sparseMatrixMap{
		data:      make([]float64, size),
		contained: make([]bool, size),
		indices:   make([]int, 0, size),
	}
}

func (a *sparseMatrixMap) Add(idx int, x float64) {
	if !a.contained[idx] {
		a.data[idx] = x
		a.contained[idx] = true
		a.indices = append(a.indices, idx)
	} else {
		a.data[idx] += x
	}
}

func (a *sparseMatrixMap) IterateSorted(f func(idx int, x float64)) {
	sort.Ints(a.indices)
	for _, idx := range a.indices {
		f(idx, a.data[idx])
	}
}

func (a *sparseMatrixMap) Clear() {
	for _, idx := range a.indices {
		a.contained[idx] = false
	}
	a.indices = a.indices[:0]
}

func permuteVectors(v []Vec3, p []int) []Vec3 {
	res := make([]Vec3, len(v))
	for i, j := range p {
		res[i] = v[j]
	}
	return res
}

func permuteVectorsInv(v []Vec3, p []int) []Vec3 {
	res := make([]Vec3, len(v))
	for i, j := range p {
		res[j] = v[i]
	}
	return res
}
