package toolbox3d

import (
	"math"
	"testing"

	"github.com/unixpickle/model3d/model3d"
)

func TestRadialCurveEquivalent(t *testing.T) {
	f := func(t float64) (model3d.Coord3D, float64) {
		return model3d.XYZ(0.1*math.Sin(5*t*math.Pi*2), math.Cos(t*math.Pi*2), math.Sin(t*math.Pi*2)), 0.5
	}
	segments := []model3d.Segment{}
	for t := 0.0; t < 1.0; t += 0.001 {
		p1, _ := f(t)
		p2, _ := f(t + 0.001)
		segments = append(segments, model3d.NewSegment(p1, p2))
	}
	expectedMesh := model3d.MarchingCubesSearch(LineJoin(0.5, segments...), 0.07, 8)
	actualMesh := model3d.MarchingCubesSearch(RadialCurve(1000, true, f), 0.07, 8)
	expected := model3d.MeshToSDF(expectedMesh)
	actual := model3d.MeshToSDF(actualMesh)
	if math.Abs(expectedMesh.Volume()-actualMesh.Volume()) > 0.01 {
		t.Errorf("mismatched volume: expected %f but got %f", expectedMesh.Volume(),
			actualMesh.Volume())
	}
	min, max := actual.Min(), actual.Max()
	for i := 0; i < 1000; i++ {
		x := model3d.NewCoord3DRandBounds(min, max)
		actualSDF := actual.SDF(x)
		expectedSDF := expected.SDF(x)
		if math.Abs(actualSDF-expectedSDF) > 0.2 {
			t.Errorf("unexpected value at %v: got %f but expected %f", x, actualSDF, expectedSDF)
			break
		}
	}
}
