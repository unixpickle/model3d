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
