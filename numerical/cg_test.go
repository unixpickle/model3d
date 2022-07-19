package numerical

import "testing"

func TestBiCGSTAB(t *testing.T) {
	matrix := NewSparseMatrix(5)
	matrix.Set(0, 0, 2.0)
	matrix.Set(0, 1, -2.0)
	matrix.Set(1, 4, 3.0)
	matrix.Set(1, 1, -0.1)
	matrix.Set(2, 0, 3.0)
	matrix.Set(2, 2, 1.0)
	matrix.Set(2, 4, -1.5)
	matrix.Set(3, 3, -1.0)
	matrix.Set(4, 3, 1.25)
	matrix.Set(4, 0, -0.25)

	groundTruth := Vec{3.14, -0.78, 2.9, 3.1, -2}
	b := matrix.Apply(groundTruth)

	// Since this is a small system, we find the exact
	// solution in N+1 iterations, where N is the size
	// of the matrix.
	solver := NewBiCGSTAB(matrix.Apply, b, nil)
	for i := 0; i < 5; i++ {
		solver.Iter()
	}
	solution := solver.Iter()
	if solution.Dist(groundTruth) > 1e-5 {
		t.Errorf("expected %v but got %v", groundTruth, solution)
	}
}
