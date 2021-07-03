package fileformats

import (
	"bytes"
	"math"
	"math/rand"
	"testing"
)

func TestSTL(t *testing.T) {
	tris := make([][4][3]float32, 50)
	for i := range tris {
		for j := 0; j < 4; j++ {
			for k := 0; k < 3; k++ {
				tris[i][j][k] = float32(rand.NormFloat64())
			}
		}
	}

	buf := bytes.NewBuffer(nil)
	writer, err := NewSTLWriter(buf, uint32(len(tris)))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range tris {
		err := writer.WriteTriangle(x[0], [3][3]float32{x[1], x[2], x[3]})
		if err != nil {
			t.Fatal(err)
		}
	}

	reader, err := NewSTLReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if int(reader.NumTriangles()) != len(tris) {
		t.Fatal("bad length")
	}

	for i, x := range tris {
		normal, vertices, err := reader.ReadTriangle()
		if err != nil {
			t.Fatal(err)
		}
		y := [4][3]float32{normal, vertices[0], vertices[1], vertices[2]}
		for j := 0; j < 4; j++ {
			for k := 0; k < 3; k++ {
				if x[j][k] != y[j][k] || math.IsNaN(float64(y[j][k])) ||
					math.IsInf(float64(y[j][k]), 0) {
					t.Errorf("mismatch at %d,%d,%d: expected %f got %f", i, j, k, x[j][k], y[j][k])
				}
			}
		}
	}
}
