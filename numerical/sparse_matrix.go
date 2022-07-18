package numerical

import (
	"sort"

	"github.com/unixpickle/essentials"
)

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
// If it is, this will result in undefined behavior.
func (s *SparseMatrix) Set(row, col int, x float64) {
	s.rows[row] = append(s.rows[row], x)
	s.indices[row] = append(s.indices[row], col)
}

// Transpose computes the matrix transpose of s.
func (s *SparseMatrix) Transpose() *SparseMatrix {
	res := NewSparseMatrix(len(s.rows))
	for i := 0; i < len(s.rows); i++ {
		s.Iterate(i, func(j int, x float64) {
			res.Set(j, i, x)
		})
	}
	return res
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

// Apply computes A*x.
func (s *SparseMatrix) Apply(x Vec) Vec {
	res := make(Vec, len(x))
	for row, indices := range s.indices {
		for col, value := range s.rows[row] {
			res[row] += x[indices[col]] * value
		}
	}
	return res
}

// ApplyVec2 computes (A*x, A*y).
func (s *SparseMatrix) ApplyVec2(x []Vec2) []Vec2 {
	return sparseMatrixApply(s, x)
}

// ApplyVec3 computes (A*x, A*y, A*z).
func (s *SparseMatrix) ApplyVec3(x []Vec3) []Vec3 {
	return sparseMatrixApply(s, x)
}

func sparseMatrixApply[T Vector[T]](s *SparseMatrix, x []T) []T {
	zero := x[0].Zeros()
	res := make([]T, len(x))
	for i := range res {
		res[i] = zero
	}
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
