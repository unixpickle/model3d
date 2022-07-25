package model2d

import (
	"math"
	"sort"
	"testing"
)

func TestTriangle(t *testing.T) {
	t.Run("Contains", func(t *testing.T) {
		tri := NewTriangle(XY(0.5, 0.2), XY(1.5, 0.2), XY(1.5, 1.2))
		tests := []struct {
			Point    Coord
			Expected bool
		}{
			{XY(-0.5, 0.2), false},
			{XY(2.5, 0.2), false},
			{XY(1.5, 1.7), false},
			{XY(0.5, 1.2), false},
			{XY(0.6, 0.25), true},
			{XY(1.0, 0.6), true},
			{XY(1.0, 0.8), false},
			{XY(1.4, 1.0), true},
			{XY(1.4, 1.15), false},
			{XY(1.55, 1.0), false},
		}
		for _, test := range tests {
			actual := tri.Contains(test.Point)
			if actual != test.Expected {
				t.Errorf("point %v should be %v but got %v", test.Point, test.Expected, actual)
			}
		}
	})

	t.Run("SDF", func(t *testing.T) {
		tri := NewTriangle(XY(0.1, 0.3), XY(1.5, -0.1), XY(1.2, 1.4))
		for i := 0; i < 1000; i++ {
			c := NewCoordRandNorm()
			point, sdf := tri.PointSDF(c)
			normal, sdf1 := tri.NormalSDF(c)
			bary, sdf2 := tri.BarycentricSDF(c)
			if math.Abs(sdf-sdf1) > 1e-8 || math.Abs(sdf-sdf2) > 1e-8 {
				t.Fatalf("inconsistent SDF results")
			}
			if math.Abs(sdf) > 1e-5 && (tri.Contains(c) != (sdf > 0)) {
				t.Fatalf("inconsistent containment")
			}
			if math.Abs(point.Dist(c)-math.Abs(sdf)) > 1e-8 {
				t.Fatalf("dist should be %f but got %f", point.Dist(c), sdf)
			}
			if math.Abs(normal.Norm()-1) > 1e-8 {
				t.Fatalf("normal is not unit length: %v", normal)
			}
			baryPoint := tri.AtBarycentric(bary)
			if math.Abs(baryPoint.Dist(c)-math.Abs(sdf)) > 1e-8 {
				t.Fatalf("dist should be %f but got %f for barycentric %v", sdf, baryPoint.Dist(c),
					bary)
			}

			// Normals should push us in/out of the
			// triangle, proving the point is on the
			// border. However, this won't necessarily
			// work near corners (but will work using
			// per-vertex normals *at* corners).
			var closeCorner bool
			for _, corner := range tri.Coords() {
				if point != corner && point.Dist(corner) < 1e-4 {
					closeCorner = true
				}
			}
			if !closeCorner {
				outside := point.Add(normal.Scale(1e-5))
				inside := point.Sub(normal.Scale(1e-5))
				if tri.Contains(outside) {
					t.Fatalf("outside point is contained: %v %v %v", point, normal, outside)
				}
				if !tri.Contains(inside) {
					t.Fatalf("inside point is not contained: %v %v %v", point, normal, inside)
				}
			}
		}
	})

	t.Run("RayCollisions", func(t *testing.T) {
		tri := NewTriangle(XY(0.5, 0.2), XY(1.5, 0.2), XY(1.5, 1.2))
		scales := []float64{}
		ray := &Ray{
			Origin:    X(1.0),
			Direction: Y(0.5),
		}
		count := tri.RayCollisions(ray, func(rc RayCollision) {
			scales = append(scales, rc.Scale)
		})
		if count != 2 || len(scales) != 2 {
			t.Fatalf("expected 2 but got %d, %d", count, len(scales))
		}
		sort.Float64s(scales)
		if scales[0] != 0.4 || scales[1] != 2*(0.2+0.5) {
			t.Fatalf("unexpected scales: %v", scales)
		}
	})

	t.Run("Degenerate", func(t *testing.T) {
		tri := NewTriangle(
			XY(0.1259765625397359, 0.062061342678565015),
			XY(0.12597656434913396, 0.0613543861243865),
			XY(0.1259765625, 0.06255842497713796),
		)
		point := XY(0.1259765625, 0.0615234375)
		bary := tri.Barycentric(point)
		if tri.AtBarycentric(bary).Dist(point) > 1e-5 {
			t.Errorf("bad bary: %v => %v", bary, tri.AtBarycentric(bary))
		}
	})
}
