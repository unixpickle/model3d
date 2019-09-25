package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestSegmentEntersSphere(t *testing.T) {
	center := Coord3D{X: 1, Y: 2, Z: 3}
	radius := 0.5

	segments := [][2]Coord3D{
		{
			{X: -1, Y: 2.4, Z: 3},
			{X: 2, Y: 2.4, Z: 3},
		},
		{
			{X: -1, Y: 2.6, Z: 3},
			{X: 2, Y: 2.6, Z: 3},
		},
	}
	insides := []bool{
		true,
		false,
	}

	for i, seg := range segments {
		actual := segmentEntersSphere(seg[0], seg[1], center, radius)
		expected := insides[i]
		if actual != expected {
			t.Errorf("test %d: expected %v but got %v", i, expected, actual)
		}
	}

	for i := 0; i < 100; i++ {
		c := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(), Z: rand.NormFloat64()}

		// Vector from center of sphere to line.
		v := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(), Z: rand.NormFloat64()}
		v = v.Scale(1 / v.Norm())

		// Direction of the line should be orthogonal to
		// the vector from the center of the sphere.
		v1 := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(), Z: rand.NormFloat64()}
		v1 = v1.Scale(1 / v1.Norm())
		v1 = v1.Add(v.Scale(-v1.Dot(v)))

		r := math.Abs(rand.NormFloat64()) + 1e-2

		v = v.Scale(rand.NormFloat64())

		p1 := c.Add(v).Add(v1.Scale(10 * r))
		p2 := c.Add(v).Add(v1.Scale(-10 * r))

		actual := segmentEntersSphere(p1, p2, c, r)
		expected := v.Norm() < r

		if actual != expected {
			t.Errorf("random case mismatch: got %v but expected %v", actual, expected)
		}
	}
}
