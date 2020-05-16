package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestARAPSparsePermute(t *testing.T) {
	matrix := newARAPSparse(1000)
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

func TestARAPCholesky(t *testing.T) {
	factor := map[[2]int]float64{}
	for i := 0; i < 15; i++ {
		for j := 0; j <= i; j++ {
			d := rand.NormFloat64()
			factor[[2]int{i, j}] = d
			factor[[2]int{j, i}] = d
		}
	}

	matrix := newARAPSparse(15)
	for i := 0; i < 15; i++ {
		for j := 0; j < 15; j++ {
			var sum float64
			for k := 0; k < 15; k++ {
				sum += factor[[2]int{i, k}] * factor[[2]int{j, k}]
			}
			matrix.Set(i, j, sum)
		}
	}

	chol := newARAPCholesky(matrix)

	inVec := make([]Coord3D, 15)
	for i := range inVec {
		inVec[i] = NewCoord3DRandNorm()
	}

	t.Run("Apply", func(t *testing.T) {
		realOut := matrix.Apply(inVec)
		cholOut := chol.Apply(inVec)
		for i, x := range realOut {
			a := cholOut[i]
			if a.Dist(x) > 1e-5 || math.IsNaN(a.Sum()) {
				t.Errorf("expected %v but got %v", x, a)
				return
			}
		}
	})

	t.Run("ApplyInverse", func(t *testing.T) {
		inverted := matrix.Apply(chol.ApplyInverse(inVec))
		for i, x := range inVec {
			a := inverted[i]
			if a.Dist(x) > 1e-5 || math.IsNaN(a.Sum()) {
				t.Errorf("expected %v but got %v", x, a)
				return
			}
		}
	})
}

func BenchmarkARAPCholesky(b *testing.B) {
	// Some arbitrary mesh shape.
	mesh := MarchingCubesSearch(JoinedSolid{
		&Sphere{Radius: 1},
		&Rect{MaxVal: Coord3D{X: 1.2, Y: 0.2, Z: 0.2}},
	}, 0.05, 8)

	op := newARAPOperator(NewARAP(mesh), map[int]Coord3D{
		// Some arbitrary constraint.
		0: Coord3D{X: 2},
	})
	mat := op.squeezedMatrix()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newARAPCholesky(mat)
	}
}
