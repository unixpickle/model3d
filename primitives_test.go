package model3d

import (
	"math"
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
			arr := p.Array()
			arr[j] += 1e-8
			t1 := *t
			t1[i] = NewCoord3DArray(arr)
			area1 := t1.Area()
			arr[j] -= 2e-8
			t1[i] = NewCoord3DArray(arr)
			area2 := t1.Area()
			coordGrad[j] = (area1 - area2) / 2e-8
		}
		grad[i] = NewCoord3DArray(coordGrad)
	}
	return &grad
}

func TestTriangleDist(t *testing.T) {
	for i := 0; i < 100; i++ {
		tri := &Triangle{NewCoord3DRandNorm(), NewCoord3DRandNorm(), NewCoord3DRandNorm()}
		for j := 0; j < 10; j++ {
			c := NewCoord3DRandNorm()
			approx := approxTriangleDist(tri, c)
			actual := tri.Dist(c)
			if math.Abs(approx-actual) > 1e-5 {
				t.Fatalf("expected %f but got %f", approx, actual)
			}
		}
	}
}

func approxTriangleDist(t *Triangle, c Coord3D) float64 {
	min := 0.0
	max := t[0].Dist(c)
	for i := 0; i < 64; i++ {
		mid := (min + max) / 2
		collides := t.SphereCollision(c, mid)
		if collides {
			max = mid
		} else {
			min = mid
		}
	}
	return (min + max) / 2
}
