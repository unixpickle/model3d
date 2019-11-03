package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestSphereSolid(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &SphereSolid{
			Center: Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func TestCylinderSolid(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &CylinderSolid{
			P1:     Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			P2:     Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func TestTorusSolid(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &TorusSolid{
			Axis:        Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Center:      Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			OuterRadius: math.Abs(rand.NormFloat64()),
			InnerRadius: math.Abs(rand.NormFloat64()),
		})
	}
}

func testSolidBounds(t *testing.T, solid Solid) {
	min := solid.Min()
	max := solid.Max()

	const epsilon = 1e-4
	for i := 0; i < 10000; i++ {
		var face, axis1, axis2 Coord3D
		faceAxis := rand.Intn(3)
		if rand.Intn(2) == 0 {
			var arr [3]float64
			arr[faceAxis] = min.array()[faceAxis] - epsilon
			face = newCoord3DArray(arr)
		} else {
			var arr [3]float64
			arr[faceAxis] = max.array()[faceAxis] + epsilon
			face = newCoord3DArray(arr)
		}
		diff := max.Sub(min).array()
		var axis1Arr, axis2Arr [3]float64
		axis1Arr[(faceAxis+1)%3] = diff[(faceAxis+1)%3]
		axis2Arr[(faceAxis+2)%3] = diff[(faceAxis+2)%3]
		axis1 = newCoord3DArray(axis1Arr)
		axis2 = newCoord3DArray(axis2Arr)

		coord := face.Add(axis1.Scale(rand.Float64())).Add(axis2.Scale(rand.Float64()))
		if solid.Contains(coord) {
			t.Error("solid contains point:", coord, "out of bounds:", min, max)
			break
		}
	}
}
