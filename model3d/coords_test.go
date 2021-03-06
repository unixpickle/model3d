package model3d

import (
	"math"
	"testing"
)

func TestCoord3DOrthoBasis(t *testing.T) {
	testBasis := func(c Coord3D) {
		b1, b2 := c.OrthoBasis()
		if math.Abs(b1.Norm()-1) > 1e-8 || math.Abs(b2.Norm()-1) > 1e-8 {
			t.Error("not unit vectors")
		} else if math.Abs(c.Dot(b1)) > 1e-8 || math.Abs(c.Dot(b2)) > 1e-8 {
			t.Error("not orthogonal to original")
		} else if math.Abs(b1.Dot(b2)) > 1e-8 {
			t.Error("not orthogonal to each other")
		}
	}
	for i := 0; i < 100; i++ {
		testBasis(NewCoord3DRandNorm())
	}
	testBasis(X(1e90))
	testBasis(Y(1e90))
	testBasis(Z(1e90))
}

func BenchmarkCoord3DOrthoBasis(b *testing.B) {
	c := NewCoord3DRandNorm()
	for i := 0; i < b.N; i++ {
		c.OrthoBasis()
	}
}

func BenchmarkCoord3DFastHash(b *testing.B) {
	c := NewCoord3DRandNorm()
	for i := 0; i < b.N; i++ {
		c.fastHash()
	}
}
