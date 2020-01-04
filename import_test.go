package model3d

import (
	"bytes"
	"math"
	"math/rand"
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
					t.Error("unexpected vector component")
				}
			}
		}
	}
}
