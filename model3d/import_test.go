package model3d

import (
	"bytes"
	"math"
	"math/rand"
	"os"
	"testing"
)

func TestImportSTL(t *testing.T) {
	original := make([]*Triangle, 10)
	for i := range original {
		t := Triangle{}
		for j := range t {
			t[j] = Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()}
		}
		original[i] = &t
	}
	decoded, err := ReadSTL(bytes.NewReader(EncodeSTL(original)))
	if err != nil {
		t.Fatal(err)
	}
	for i, actual := range decoded {
		expected := original[i]
		for j, p1 := range actual {
			p2 := expected[j]
			for k, c1 := range p1.Array() {
				c2 := p2.Array()[k]
				if math.Abs(c1-c2) > 1e-4 {
					t.Error("unexpected vector component", c1, c2)
				}
			}
		}
	}
}

func TestImportOFF(t *testing.T) {
	f, err := os.Open("test_data/cube.off")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	triangles, err := ReadOFF(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(triangles) != 12 {
		t.Errorf("expected %d triangles but got %d", 12, len(triangles))
	}
	mesh := NewMeshTriangles(triangles)
	volume := mesh.Volume()
	if math.Abs(volume-1) > 1e-5 || math.IsNaN(volume) || math.IsInf(volume, 0) {
		t.Errorf("incorrect volume: %f", volume)
	}
	area := mesh.Area()
	if math.Abs(area-6) > 1e-5 || math.IsNaN(area) || math.IsInf(area, 0) {
		t.Errorf("incorrect area: %f", area)
	}
}
