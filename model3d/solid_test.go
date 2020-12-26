package model3d

import "testing"

func TestJoinedSolidOptimize(t *testing.T) {
	js := JoinedSolid{}
	for i := 0; i < 10; i++ {
		js = append(js, &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: 0.1,
		})
	}
	opt := js.Optimize()

	if opt.Min() != js.Min() {
		t.Error("incorrect min")
	}
	if opt.Max() != js.Max() {
		t.Error("incorrect max")
	}

	for i := 0; i < 10000; i++ {
		c := NewCoord3DRandNorm()
		actual := opt.Contains(c)
		expected := js.Contains(c)
		if actual != expected {
			t.Errorf("expected contains %v but got %v", expected, actual)
		}
	}
}
