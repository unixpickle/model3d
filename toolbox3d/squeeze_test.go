package toolbox3d

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d/model3d"
)

func TestAxisSqueeze(t *testing.T) {
	for _, axis := range []Axis{AxisX, AxisY, AxisZ} {
		t.Run(fmt.Sprintf("Axis%d", axis), func(t *testing.T) {
			t.Run("Smaller", func(t *testing.T) {
				testTransform(t, &AxisSqueeze{
					Min:   -0.1,
					Max:   0.9,
					Axis:  axis,
					Ratio: 0.8,
				})
			})
			t.Run("Larger", func(t *testing.T) {
				testTransform(t, &AxisSqueeze{
					Min:   -0.1,
					Max:   0.9,
					Axis:  axis,
					Ratio: 1.2,
				})
			})
		})
	}
}

func testTransform(t *testing.T, transform model3d.Transform) {
	solid := testingSolid()
	mesh := model3d.MarchingCubesSearch(solid, 0.02, 8)
	sdf := model3d.MeshToSDF(mesh)

	transformed := model3d.TransformSolid(transform, solid)

	solid1 := model3d.TransformSolid(transform.Inverse(), transformed)
	mesh1 := model3d.MarchingCubesSearch(solid1, 0.02, 8)
	sdf1 := model3d.MeshToSDF(mesh1)

	mesh2 := model3d.MarchingCubesSearch(transformed, 0.02, 8)
	mesh2 = mesh2.MapCoords(transform.Inverse().Apply)
	sdf2 := model3d.MeshToSDF(mesh2)
	solid2 := model3d.NewColliderSolid(model3d.MeshToCollider(mesh2))

	min, max := solid.Min(), solid.Max()
	min = min.Sub(model3d.XYZ(1, 1, 1).Scale(0.1))
	max = max.Add(model3d.XYZ(1, 1, 1).Scale(0.1))
	for i := 0; i < 1000; i++ {
		c := model3d.NewCoord3DRandUniform().Mul(max.Sub(min)).Add(min)
		dist := sdf.SDF(c)
		dist1 := sdf1.SDF(c)
		dist2 := sdf2.SDF(c)
		if math.Abs(dist-dist1) > 0.04 {
			t.Errorf("bad SDF at %v", c)
		}
		if math.Abs(dist-dist2) > 0.04 {
			t.Errorf("bad SDF at %v", c)
		}
		if math.Abs(dist) > 0.04 {
			contained := solid.Contains(c)
			contained1 := solid1.Contains(c)
			contained2 := solid2.Contains(c)
			if contained != contained1 {
				t.Errorf("disagreement on solid transform at %v", c)
			}
			if contained != contained2 {
				t.Errorf("disagreement on mesh transform at %v", c)
			}
		}
	}
}

func TestSmartSqueeze(t *testing.T) {
	bounds := &model3d.Rect{
		MinVal: model3d.XYZ(-1, -2, -3),
		MaxVal: model3d.XYZ(2, 1, -1),
	}

	t.Run("Empty", func(t *testing.T) {
		ss := &SmartSqueeze{
			Axis:         AxisZ,
			SqueezeRatio: 0.3,
		}
		testSmartSqueezePermuted(t, bounds, ss, &AxisSqueeze{
			Axis:  AxisZ,
			Min:   -3,
			Max:   -1,
			Ratio: 0.3,
		})
	})

	t.Run("SingleMiddle", func(t *testing.T) {
		ss := &SmartSqueeze{
			Axis:         AxisZ,
			SqueezeRatio: 0.3,
		}
		ss.AddUnsqueezable(-2.5, -1.5)
		testSmartSqueezePermuted(t, bounds, ss, model3d.JoinedTransform{
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -1.5,
				Max:   -1,
				Ratio: 0.3,
			},
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -3,
				Max:   -2.5,
				Ratio: 0.3,
			},
		})

		// Sanity check that points below the squeezed region are unchanged.
		xform := ss.Transform(bounds)
		c := model3d.XYZ(1, 2, -4)
		if xform.Apply(c).Dist(c) > 1e-5 {
			t.Error("unexpected change to lower point")
		}
	})

	t.Run("SingleBottom", func(t *testing.T) {
		ss := &SmartSqueeze{
			Axis:         AxisZ,
			SqueezeRatio: 0.3,
		}
		ss.AddUnsqueezable(-4, -2.5)
		testSmartSqueezePermuted(t, bounds, ss, &AxisSqueeze{
			Axis:  AxisZ,
			Min:   -2.5,
			Max:   -1,
			Ratio: 0.3,
		})
	})

	t.Run("SingleTop", func(t *testing.T) {
		ss := &SmartSqueeze{
			Axis:         AxisZ,
			SqueezeRatio: 0.3,
		}
		ss.AddUnsqueezable(-1.5, 0)
		testSmartSqueezePermuted(t, bounds, ss, &AxisSqueeze{
			Axis:  AxisZ,
			Min:   -3,
			Max:   -1.5,
			Ratio: 0.3,
		})
	})

	t.Run("PinchAndSqueeze", func(t *testing.T) {
		ss := &SmartSqueeze{
			Axis:         AxisZ,
			SqueezeRatio: 0.3,
			PinchRange:   0.05,
			PinchPower:   2,
		}
		ss.AddUnsqueezable(-2.5, -2)
		ss.AddPinch(-2.9)
		testSmartSqueezePermuted(t, bounds, ss, model3d.JoinedTransform{
			&AxisPinch{
				Axis:  AxisZ,
				Min:   -2.95,
				Max:   -2.85,
				Power: 2,
			},
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -2,
				Max:   -1,
				Ratio: 0.3,
			},
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -2.85,
				Max:   -2.5,
				Ratio: 0.3,
			},
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -3,
				Max:   -2.95,
				Ratio: 0.3,
			},
		})
	})

	t.Run("SqueezeOverlap", func(t *testing.T) {
		ss := &SmartSqueeze{
			Axis:         AxisZ,
			SqueezeRatio: 0.3,
			PinchRange:   0.05,
			PinchPower:   2,
		}
		ss.AddUnsqueezable(-2.5, -2)
		ss.AddUnsqueezable(-2.3, -2)
		ss.AddUnsqueezable(-2.6, -1.9)
		ss.AddUnsqueezable(-2.2, -1.9)
		testSmartSqueezePermuted(t, bounds, ss, model3d.JoinedTransform{
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -1.9,
				Max:   -1,
				Ratio: 0.3,
			},
			&AxisSqueeze{
				Axis:  AxisZ,
				Min:   -3,
				Max:   -2.6,
				Ratio: 0.3,
			},
		})
	})
}

func testSmartSqueezePermuted(t *testing.T, b model3d.Bounder, s *SmartSqueeze,
	expected model3d.Transform) {
	for i := 0; i < 10; i++ {
		s1 := &SmartSqueeze{
			Axis:         s.Axis,
			PinchRange:   s.PinchRange,
			SqueezeRatio: s.SqueezeRatio,
			PinchPower:   s.PinchPower,
		}
		for _, j := range rand.Perm(len(s.Pinches)) {
			s1.AddPinch(s.Pinches[j])
		}
		for _, j := range rand.Perm(len(s.Unsqueezable)) {
			r := s.Unsqueezable[j]
			s1.AddUnsqueezable(r[0], r[1])
		}
		actual := s1.Transform(b)
		for i := 0; i < 100; i++ {
			c := model3d.NewCoord3DRandNorm()
			actualPoint := actual.Apply(c)
			expectedPoint := expected.Apply(c)
			if actualPoint.Dist(expectedPoint) > 1e-5 {
				t.Errorf("expected %v but got %v for input %v", expectedPoint, actualPoint, c)
				break
			}
		}
	}
}

func testingSolid() model3d.Solid {
	return model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.XY(0.2, 0.3),
			P2:     model3d.XZ(0.3, 0.5),
			Radius: 0.1,
		},
		&model3d.Cylinder{
			P1:     model3d.X(0.2),
			P2:     model3d.XZ(0.3, 0.5),
			Radius: 0.1,
		},
		&model3d.Sphere{Center: model3d.XZ(0.25, 0.25), Radius: 0.2},
	}
}
