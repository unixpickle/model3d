package toolbox3d

import (
	"testing"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func TestChamferInsetFunc(t *testing.T) {
	t.Run("Inwards", func(t *testing.T) {
		chamfer := &ChamferInsetFunc{
			BottomRadius: 2,
			TopRadius:    1,
		}
		if minInset := chamfer.MinInset(); minInset != 0 {
			t.Fatalf("unexpected min inset: %f", minInset)
		}
		testCases := []struct {
			z        float64
			expected float64
		}{
			{z: 0, expected: 2},
			{z: 0.5, expected: 1.5},
			{z: 1, expected: 1},
			{z: 2, expected: 0},
			{z: 4, expected: 0},
			{z: 4.5, expected: 0.5},
			{z: 5, expected: 1},
		}
		for _, tc := range testCases {
			if actual := chamfer.Inset(0, 5, tc.z); actual != tc.expected {
				t.Errorf("unexpected inset at z=%f: got %f but expected %f", tc.z, actual, tc.expected)
			}
		}
	})

	t.Run("Outwards", func(t *testing.T) {
		chamfer := &ChamferInsetFunc{
			BottomRadius: 2,
			TopRadius:    1,
			Outwards:     true,
		}
		if minInset := chamfer.MinInset(); minInset != -2 {
			t.Fatalf("unexpected min inset: %f", minInset)
		}
		testCases := []struct {
			z        float64
			expected float64
		}{
			{z: 0, expected: -2},
			{z: 0.5, expected: -1.5},
			{z: 1, expected: -1},
			{z: 2, expected: 0},
			{z: 4, expected: 0},
			{z: 4.5, expected: -0.5},
			{z: 5, expected: -1},
		}
		for _, tc := range testCases {
			if actual := chamfer.Inset(0, 5, tc.z); actual != tc.expected {
				t.Errorf("unexpected inset at z=%f: got %f but expected %f", tc.z, actual, tc.expected)
			}
		}
	})
}

func TestExtrudeChamferInsetFunc(t *testing.T) {
	shape := model2d.NewRect(model2d.Origin, model2d.XY(10, 10))

	t.Run("Inwards", func(t *testing.T) {
		solid := Extrude(shape, 0, 3, &ChamferInsetFunc{BottomRadius: 2})
		testCases := []struct {
			point    model3d.Coord3D
			contains bool
		}{
			{point: model3d.XYZ(5, 5, 0), contains: true},
			{point: model3d.XYZ(1.5, 5, 0), contains: false},
			{point: model3d.XYZ(1.5, 5, 0.5), contains: false},
			{point: model3d.XYZ(1.5, 5, 1), contains: true},
			{point: model3d.XYZ(0.5, 5, 1.5), contains: false},
			{point: model3d.XYZ(0.5, 5, 2), contains: true},
		}
		for _, tc := range testCases {
			if actual := solid.Contains(tc.point); actual != tc.contains {
				t.Errorf("unexpected containment for %v: got %v but expected %v",
					tc.point, actual, tc.contains)
			}
		}
	})

	t.Run("OutwardsBounds", func(t *testing.T) {
		solid := Extrude(shape, 0, 3, &ChamferInsetFunc{BottomRadius: 2, Outwards: true})
		if min := solid.Min(); min != model3d.XYZ(-2, -2, 0) {
			t.Fatalf("unexpected min bounds: %v", min)
		}
		if max := solid.Max(); max != model3d.XYZ(12, 12, 3) {
			t.Fatalf("unexpected max bounds: %v", max)
		}
		testCases := []struct {
			point    model3d.Coord3D
			contains bool
		}{
			{point: model3d.XYZ(-1.5, 5, 0), contains: true},
			{point: model3d.XYZ(-1.5, 5, 0.5), contains: false},
			{point: model3d.XYZ(-0.5, 5, 0.5), contains: true},
			{point: model3d.XYZ(-0.5, 5, 2), contains: false},
		}
		for _, tc := range testCases {
			if actual := solid.Contains(tc.point); actual != tc.contains {
				t.Errorf("unexpected containment for %v: got %v but expected %v",
					tc.point, actual, tc.contains)
			}
		}
	})
}
