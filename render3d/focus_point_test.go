package render3d

import (
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d"
)

func TestSampleAroundUniform(t *testing.T) {
	direction := model3d.NewCoord3DRandUnit()
	minCos := 0.83
	f := func(c model3d.Coord3D) model3d.Coord3D {
		if c.Dot(direction) < minCos {
			return model3d.Coord3D{}
		}
		return model3d.Coord3D{
			X: c.X*c.Y - c.Z,
			Y: c.Z*c.Z - 0.7*c.Y*c.Y + 0.3*c.X*c.X,
			Z: 0.6*c.X + 0.5*c.Y + 0.3*c.Z,
		}
	}

	const iters = 5000000

	var expected model3d.Coord3D
	for i := 0; i < iters; i++ {
		expected = expected.Add(f(model3d.NewCoord3DRandUnit()))
	}

	gen := rand.New(rand.NewSource(1337))

	var actual model3d.Coord3D
	for i := 0; i < iters; i++ {
		realDir := sampleAroundUniform(gen, minCos, direction)
		weight := 1 / densityAroundUniform(minCos, direction, realDir)
		actual = actual.Add(f(realDir).Scale(weight))
	}

	diff := actual.Sub(expected).Scale(1.0 / iters)
	if diff.Norm() > 1e-3 {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}
