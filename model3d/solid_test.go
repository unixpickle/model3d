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

func TestSolidMux(t *testing.T) {
	solids := make([]Solid, 5)
	for i := 0; i < 5; i++ {
		solids[i] = &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: 0.7,
		}
	}
	mux := NewSolidMux(solids)

	groundTruth := func(c Coord3D) []bool {
		res := make([]bool, len(solids))
		for i, s := range solids {
			res[i] = s.Contains(c)
		}
		return res
	}

	for i := 0; i < 10000; i++ {
		c := NewCoord3DRandNorm()
		actual := mux.AllContains(c)
		expected := groundTruth(c)
		if len(actual) != len(expected) {
			t.Fatal("bad length")
		}
		for i, x := range expected {
			a := actual[i]
			if a != x {
				t.Fatalf("index %d: expected %v but got %v", i, x, a)
			}
		}

		actualContains := mux.Contains(c)
		expectedContains := JoinedSolid(solids).Contains(c)
		if actualContains != expectedContains {
			t.Fatalf("containment at %v: expected %v but got %v", c, expectedContains,
				actualContains)
		}
	}
}
