package numerical

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestARAPSparsePermute(t *testing.T) {
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

func TestSparseCholesky(t *testing.T) {
	factor := map[[2]int]float64{}
	for i := 0; i < 15; i++ {
		for j := 0; j <= i; j++ {
			d := rand.NormFloat64()
			factor[[2]int{i, j}] = d
			factor[[2]int{j, i}] = d
		}
	}

	matrix := NewSparseMatrix(15)
	for i := 0; i < 15; i++ {
		for j := 0; j < 15; j++ {
			var sum float64
			for k := 0; k < 15; k++ {
				sum += factor[[2]int{i, k}] * factor[[2]int{j, k}]
			}
			matrix.Set(i, j, sum)
		}
	}

	chol := NewSparseCholesky(matrix)

	inVec := make([]Vec3, 15)
	for i := range inVec {
		inVec[i] = NewVec3RandomNormal()
	}

	t.Run("ApplyVec3", func(t *testing.T) {
		realOut := matrix.ApplyVec3(inVec)
		cholOut := chol.ApplyVec3(inVec)
		for i, x := range realOut {
			a := cholOut[i]
			if a.Dist(x) > 1e-5 || math.IsNaN(a.Sum()) {
				t.Errorf("expected %v but got %v", x, a)
				return
			}
		}
	})

	t.Run("ApplyInverseVec3", func(t *testing.T) {
		inverted := matrix.ApplyVec3(chol.ApplyInverseVec3(inVec))
		for i, x := range inVec {
			a := inverted[i]
			if a.Dist(x) > 1e-5 || math.IsNaN(a.Sum()) {
				t.Errorf("expected %v but got %v", x, a)
				return
			}
		}
	})
}

func BenchmarkSparseCholesky(b *testing.B) {
	r, err := os.Open("test_data/sparse_mat.json.gz")
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()
	gr, err := gzip.NewReader(r)
	if err != nil {
		b.Fatal(err)
	}
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		b.Fatal(err)
	}
	unjson := map[string]string{}
	if err := json.Unmarshal(data, &unjson); err != nil {
		b.Fatal(err)
	}
	coordToValue := map[[2]int]float64{}
	maxCoord := 0
	for k, v := range unjson {
		parts := strings.Split(k, ",")
		x, _ := strconv.Atoi(parts[0])
		y, _ := strconv.Atoi(parts[1])
		coord := [2]int{x, y}
		value, _ := strconv.ParseFloat(v, 64)
		coordToValue[coord] = value
		if x > maxCoord {
			maxCoord = x
		}
	}

	mat := NewSparseMatrix(maxCoord + 1)
	for coord, value := range coordToValue {
		mat.Set(coord[0], coord[1], value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSparseCholesky(mat)
	}
}
