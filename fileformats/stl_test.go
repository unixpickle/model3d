package fileformats

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"testing"
)

func TestSTLBinary(t *testing.T) {
	testBinaryReader := func(t *testing.T, header []byte) {
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

		data := buf.Bytes()
		copy(data, header)

		reader, err := NewSTLReader(bytes.NewReader(data))
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
		for i := 0; i < 3; i++ {
			_, _, err = reader.ReadTriangle()
			if err != io.EOF {
				t.Fatalf("unexpected extra read error: %v", err)
			}
		}
	}

	t.Run("Basic", func(t *testing.T) {
		testBinaryReader(t, []byte{})
	})
	t.Run("DummySolid", func(t *testing.T) {
		testBinaryReader(t, []byte("solid"))
	})
	t.Run("SpaceSolid", func(t *testing.T) {
		header := make([]byte, 80)
		copy(header, []byte("solid"))
		for i := 5; i < len(header); i++ {
			header[i] = ' '
		}
		testBinaryReader(t, header)
	})
}

func TestSTLASCII(t *testing.T) {
	data := `solid someModel
	facet normal -0.998944 0.0459364 0

      outer loop
		vertex -87.9551 -78.7197 5
		vertex -87.9551 -78.7197 0

		vertex -88.1253 -82.4205 5
	  endloop
	endfacet
	facet normal 0 0 -1
      outer loop
        vertex 41.2344 53.881 0
        vertex 37.3304 73.9724 0
        vertex 36.99 80.7974 0
      endloop
	endfacet
	endsolid`

	testWithFinalNewline := func(t *testing.T, finalNewline bool) {
		lineData := data
		if finalNewline {
			lineData += "\n\n"
		}
		reader, err := NewSTLReader(bytes.NewReader([]byte(lineData)))
		if err != nil {
			t.Fatal(err)
		}
		var normals [][3]float32
		var tris [][3][3]float32
		for {
			normal, tri, err := reader.ReadTriangle()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			normals = append(normals, normal)
			tris = append(tris, tri)
		}
		for i := 0; i < 3; i++ {
			_, _, err = reader.ReadTriangle()
			if err != io.EOF {
				t.Fatalf("unexpected extra read error: %v", err)
			}
		}
		if len(tris) != 2 {
			t.Fatalf("expected 2 triangles but got %d", len(tris))
		}
		if normals[0] != [3]float32{-0.998944, 0.0459364, 0} || normals[1] != [3]float32{0, 0, -1} {
			t.Fatalf("unexpected normals: %v", normals)
		}
		expectedTris := [][3][3]float32{
			{
				{-87.9551, -78.7197, 5},
				{-87.9551, -78.7197, 0},
				{-88.1253, -82.4205, 5},
			},
			{
				{41.2344, 53.881, 0},
				{37.3304, 73.9724, 0},
				{36.99, 80.7974, 0},
			},
		}
		for i, x := range expectedTris {
			a := tris[i]
			if a != x {
				t.Errorf("expected tri %d to be %v but got %v", i, x, a)
			}
		}
	}

	t.Run("NoNewline", func(t *testing.T) {
		testWithFinalNewline(t, false)
	})
	t.Run("FinalNewline", func(t *testing.T) {
		testWithFinalNewline(t, true)
	})
}
