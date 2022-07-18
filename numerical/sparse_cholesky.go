package numerical

import (
	"math"
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
