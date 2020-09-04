package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestMeshVolume(t *testing.T) {
	for i := 0; i < 5; i++ {
		solid := &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: rand.Float64()*2 + 0.1,
		}
		mesh := MarchingCubesSearch(solid, 0.02, 8)
		expected := 4.0 / 3.0 * math.Pi * math.Pow(solid.Radius, 3)
		actual := mesh.Volume()
		if math.Abs(expected-actual)/actual > 1e-2 {
			t.Errorf("expected volume %f but got %f", expected, actual)
		}
	}
}
