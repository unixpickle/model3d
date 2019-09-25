package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestMatrix3Inverse(t *testing.T) {
	for i := 0; i < 10; i++ {
		m := Matrix3{}
		for j := range m {
			m[j] = rand.NormFloat64()
		}
		m1 := m.Inverse()
		product := m.Mul(m1)
		product[0] -= 1
		product[4] -= 1
		product[8] -= 1
		for j, x := range product {
			if math.Abs(x) > 1e-8 {
				t.Errorf("entry %d should be 0 but got %f", j, x)
			}
		}
	}
}
