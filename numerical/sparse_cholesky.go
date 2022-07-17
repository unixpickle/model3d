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

// ApplyVec2 computes (A*x, A*y).
func (s *SparseCholesky) ApplyVec2(x []Vec2) []Vec2 {
	return sparseCholeskyApply(s, x)
}

// ApplyVec3 computes (A*x, A*y, A*z).
func (s *SparseCholesky) ApplyVec3(x []Vec3) []Vec3 {
	return sparseCholeskyApply(s, x)
}

func sparseCholeskyApply[T Vector[T]](s *SparseCholesky, x []T) []T {
	out := make([]T, len(x))
	b := permuteVectors(x, s.perm)
	for i := range out {
		sum := x[0].Zeros()
		s.upper.Iterate(i, func(col int, x float64) {
			sum = sum.Add(b[col].Scale(x))
		})
		out[i] = sum
	}
	for i := len(out) - 1; i >= 0; i-- {
		sum := x[0].Zeros()
		s.lower.Iterate(i, func(col int, x float64) {
			sum = sum.Add(out[col].Scale(x))
		})
		out[i] = sum
	}
	return permuteVectorsInv(out, s.perm)
}

// ApplyInverseVec2 computes (A^-1*x, A^-1*y).
func (s *SparseCholesky) ApplyInverseVec2(x []Vec2) []Vec2 {
	return sparseCholeskyApplyInverse(s, x)
}

// ApplyInverseVec3 computes (A^-1*x, A^-1*y, A^-1*z).
func (s *SparseCholesky) ApplyInverseVec3(x []Vec3) []Vec3 {
	return sparseCholeskyApplyInverse(s, x)
}

func sparseCholeskyApplyInverse[T Vector[T]](s *SparseCholesky, x []T) []T {
	b := permuteVectors(x, s.perm)
	out := make([]T, len(x))
	sparseMatrixBacksubLower(s.lower, out, b)
	sparseMatrixBacksubUpper(s.upper, out, out)
	return permuteVectorsInv(out, s.perm)
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
func (s *SparseMatrix) Set(row, col int, x float64) {
	s.rows[row] = append(s.rows[row], x)
	s.indices[row] = append(s.indices[row], col)
}

// Iterate loops through the non-zero entries in a row.
func (s *SparseMatrix) Iterate(row int, f func(col int, x float64)) {
	for i, col := range s.indices[row] {
		f(col, s.rows[row][i])
	}
}

// Permute permutes the rows and columns by perm, where
// perm is the result of applying the permutation to the
// list [0...n-1].
func (s *SparseMatrix) Permute(perm []int) *SparseMatrix {
	permInv := make([]int, len(perm))
	for i, j := range perm {
		permInv[j] = i
	}
	res := NewSparseMatrix(len(perm))
	for i, j := range perm {
		oldRow := s.indices[j]
		newRow := make([]int, 0, len(oldRow))
		for _, k := range oldRow {
			newRow = append(newRow, permInv[k])
		}
		res.indices[i] = newRow
		res.rows[i] = append([]float64{}, s.rows[j]...)
	}
	return res
}

// RCM computes the reverse Cuthill-McKee permutation for
// the matrix.
func (s *SparseMatrix) RCM() []int {
	remaining := map[int]bool{}
	for i := range s.indices {
		remaining[i] = true
	}

	remainingNeighbors := func(i int) int {
		var count int
		for _, neighbor := range s.indices[i] {
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

	permutation := make([]int, 0, len(s.indices))
	for i := range s.indices {
		var expand int
		if i >= len(permutation) {
			expand = drawBestStart()
			permutation = append(permutation, expand)
			delete(remaining, expand)
		} else {
			expand = permutation[i]
		}

		allAdj := s.indices[expand]
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
func (s *SparseMatrix) ApplyVec3(x []Vec3) []Vec3 {
	res := make([]Vec3, len(x))
	for row, indices := range s.indices {
		for col, value := range s.rows[row] {
			res[row] = res[row].Add(x[indices[col]].Scale(value))
		}
	}
	return res
}

// sparseMatrixBacksubUpper writes U^-1*b to out, assuming
// this is an upper-triangular matrix U.
func sparseMatrixBacksubUpper[T Vector[T]](s *SparseMatrix, out, b []T) {
	for row := len(b) - 1; row >= 0; row-- {
		bValue := b[row]
		var diagValue float64
		s.Iterate(row, func(col int, x float64) {
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

// sparseMatrixBacksubLower writes L^-1*b to out, assuming
// this is a lower-triangular matrix L.
func sparseMatrixBacksubLower[T Vector[T]](s *SparseMatrix, out, b []T) {
	for row, bValue := range b {
		var diagValue float64
		s.Iterate(row, func(col int, x float64) {
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

func (s *sparseMatrixMap) Add(idx int, x float64) {
	if !s.contained[idx] {
		s.data[idx] = x
		s.contained[idx] = true
		s.indices = append(s.indices, idx)
	} else {
		s.data[idx] += x
	}
}

func (s *sparseMatrixMap) IterateSorted(f func(idx int, x float64)) {
	sort.Ints(s.indices)
	for _, idx := range s.indices {
		f(idx, s.data[idx])
	}
}

func (s *sparseMatrixMap) Clear() {
	for _, idx := range s.indices {
		s.contained[idx] = false
	}
	s.indices = s.indices[:0]
}

func permuteVectors[T Vector[T]](v []T, p []int) []T {
	res := make([]T, len(v))
	for i, j := range p {
		res[i] = v[j]
	}
	return res
}

func permuteVectorsInv[T Vector[T]](v []T, p []int) []T {
	res := make([]T, len(v))
	for i, j := range p {
		res[j] = v[i]
	}
	return res
}
