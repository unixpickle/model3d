package numerical

import (
	"math/rand"
	"testing"
)

func TestSparseMatrixPermute(t *testing.T) {
	matrix := NewSparseMatrix(1000)
	entries := map[[2]int]float64{}

	for i := 0; i < 100; i++ {
		x := rand.Intn(1000)
		y := rand.Intn(1000)
		if _, ok := entries[[2]int{x, y}]; ok {
			i--
			continue
		}
		val := rand.NormFloat64()
		entries[[2]int{x, y}] = val
		entries[[2]int{y, x}] = val
		matrix.Set(x, y, val)
		matrix.Set(y, x, val)
	}

	perm := rand.Perm(1000)

	permuted := matrix.Permute(perm)
	permEntries := map[[2]int]bool{}
	for row := 0; row < 1000; row++ {
		permuted.Iterate(row, func(col int, x float64) {
			origRow := perm[row]
			origCol := perm[col]
			expected := entries[[2]int{origRow, origCol}]
			if x != expected {
				t.Errorf("unexpected entry: %f (expected %f)", x, expected)
			}
			permEntries[[2]int{row, col}] = true
		})
	}

	if len(permEntries) != len(entries) {
		t.Errorf("expected %d entries but got %d", len(entries), len(permEntries))
	}
}
