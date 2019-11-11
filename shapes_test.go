package model3d

import (
	"math/rand"
	"testing"
)

func TestTriangleAreaGradient(t *testing.T) {
	for i := 0; i < 1000; i++ {
		tri := &Triangle{
			Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
		}
		actual := tri.AreaGradient()
		expected := approxTriangleAreaGradient(tri)
		for i, a := range actual {
			e := expected[i]
			if a.Dist(e) > 1e-5 {
				t.Error("invalid gradient", a, "expected", e)
			}
		}
	}
}

func approxTriangleAreaGradient(t *Triangle) *Triangle {
	var grad Triangle
	for i, p := range t {
		var coordGrad [3]float64
		for j := 0; j < 3; j++ {
			arr := p.array()
			arr[j] += 1e-8
			t1 := *t
			t1[i] = newCoord3DArray(arr)
			area1 := t1.Area()
			arr[j] -= 2e-8
			t1[i] = newCoord3DArray(arr)
			area2 := t1.Area()
			coordGrad[j] = (area1 - area2) / 2e-8
		}
		grad[i] = newCoord3DArray(coordGrad)
	}
	return &grad
}
