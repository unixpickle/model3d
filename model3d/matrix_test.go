package model3d

import (
	"math"
	"math/cmplx"
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

func TestMatrix3Rotation(t *testing.T) {
	t.Run("AroundY", func(t *testing.T) {
		theta := 0.1
		mat := NewMatrix3Rotation(Coord3D{Y: 1}, theta)
		c1 := Coord3D{X: 2, Y: 1, Z: 3}
		newTheta := math.Atan2(c1.Z, c1.X) - theta
		norm := Coord2D{X: c1.X, Y: c1.Z}.Norm()
		expected := Coord3D{
			X: norm * math.Cos(newTheta),
			Z: norm * math.Sin(newTheta),
			Y: c1.Y,
		}
		actual := mat.MulColumn(c1)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	})

	t.Run("AroundZ", func(t *testing.T) {
		theta := 0.1
		mat := NewMatrix3Rotation(Coord3D{Z: 1}, theta)
		c1 := Coord3D{X: 2, Y: 1, Z: 3}
		newTheta := math.Atan2(c1.Y, c1.X) + theta
		norm := c1.Coord2D().Norm()
		expected := Coord3D{
			X: norm * math.Cos(newTheta),
			Y: norm * math.Sin(newTheta),
			Z: 3,
		}
		actual := mat.MulColumn(c1)
		if actual.Dist(expected) > 1e-5 {
			t.Errorf("expected %v but got %v", expected, actual)
		}
	})

	t.Run("Flip", func(t *testing.T) {
		axis := NewCoord3DRandUnit()
		c := NewCoord3DRandNorm()
		c1 := NewMatrix3Rotation(axis, 0.1).MulColumn(c)
		c2 := NewMatrix3Rotation(axis.Scale(-1), -0.1).MulColumn(c)
		if c1.Dist(c2) > 1e-5 {
			t.Error("negating axis negates rotation direction")
		}
	})
}

func TestMatrix3Eigenvalues(t *testing.T) {
	mats := []*Matrix3{
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{0, 1, 0, 0, 0, 1, 1, 0, 0},
	}
	eigs := [][3]complex128{
		{0, complex(0.5*(3*math.Sqrt(33)+15), 0), complex(0.5*(-3*math.Sqrt(33)+15), 0)},
		{1, complex(-1, math.Sqrt(3)) / 2, complex(-1, -math.Sqrt(3)) / 2},
	}
	for i, mat := range mats {
		expected := eigs[i]
		actual := map[complex128]int{}
		for _, x := range mat.Eigenvalues() {
			actual[x]++
		}
		for j, x := range expected {
			var notFirst bool
			var a complex128
			for anA, count := range actual {
				if count == 0 {
					continue
				}
				if !notFirst || cmplx.Abs(anA-x) < cmplx.Abs(a-x) {
					a = anA
					notFirst = true
				}
			}
			actual[a]--
			if cmplx.Abs(x-a) > 1e-8 {
				t.Errorf("case %d eig %d: should be %f but got %f", i, j, x, a)
			}
		}
	}
}
