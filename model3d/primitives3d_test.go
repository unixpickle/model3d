package model3d

import (
	"math"
	"math/rand"
	"sort"
	"testing"
)

type solidSDF interface {
	Solid
	SDF(Coord3D) float64
}

type solidColliderSDF interface {
	Collider
	SDF(Coord3D) float64
	Contains(c Coord3D) bool
}

func TestSphereBounds(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func TestCylinderBounds(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &Cylinder{
			P1:     NewCoord3DRandNorm(),
			P2:     NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func TestConeBounds(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &Cone{
			Tip:    NewCoord3DRandNorm(),
			Base:   NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func TestTorusBounds(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidBounds(t, &Torus{
			Axis:        NewCoord3DRandNorm(),
			Center:      NewCoord3DRandNorm(),
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
			arr[faceAxis] = min.Array()[faceAxis] - epsilon
			face = NewCoord3DArray(arr)
		} else {
			var arr [3]float64
			arr[faceAxis] = max.Array()[faceAxis] + epsilon
			face = NewCoord3DArray(arr)
		}
		diff := max.Sub(min).Array()
		var axis1Arr, axis2Arr [3]float64
		axis1Arr[(faceAxis+1)%3] = diff[(faceAxis+1)%3]
		axis2Arr[(faceAxis+2)%3] = diff[(faceAxis+2)%3]
		axis1 = NewCoord3DArray(axis1Arr)
		axis2 = NewCoord3DArray(axis2Arr)

		coord := face.Add(axis1.Scale(rand.Float64())).Add(axis2.Scale(rand.Float64()))
		if solid.Contains(coord) {
			t.Error("solid contains point:", coord, "out of bounds:", min, max)
			break
		}
	}
}

func TestConeContainment(t *testing.T) {
	cone := &Cone{Tip: Z(2), Base: Z(0), Radius: 0.5}
	testPoints := map[Coord3D]bool{
		Z(1):             true,
		Z(1.999):         true,
		XZ(0.1, 1.999):   false,
		XZ(0.2, 1):       true,
		XZ(0.25, 1):      true,
		XZ(0.49, 0.0001): true,
	}
	for c, expected := range testPoints {
		actual := cone.Contains(c)
		if actual != expected {
			t.Errorf("coord %v: expected %v but got %v", c, expected, actual)
		}
	}
}

func TestRectSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		c1 := NewCoord3DRandNorm()
		c2 := NewCoord3DRandNorm()
		testSolidSDF(t, &Rect{
			MinVal: c1.Min(c2),
			MaxVal: c1.Max(c2).Add(XYZ(0.1, 0.1, 0.1)),
		})
	}
}

func TestSphereSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidSDF(t, &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()) + 0.1,
		})
	}
}

func TestCylinderSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		p1 := NewCoord3DRandUnit()
		p2 := NewCoord3DRandUnit()
		if p1.Dist(p2) < 0.1 {
			i--
			continue
		}
		testSolidSDF(t, &Cylinder{
			P1:     NewCoord3DRandNorm(),
			P2:     NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()) + 0.1,
		})
	}
}

func TestTorusSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		outer := math.Abs(rand.NormFloat64()) + 0.1
		inner := math.Abs(rand.NormFloat64()) + 0.1
		if inner > outer {
			outer, inner = inner, outer
		}
		testSolidBounds(t, &Torus{
			Axis:        NewCoord3DRandNorm(),
			Center:      NewCoord3DRandNorm(),
			OuterRadius: outer,
			InnerRadius: inner,
		})
	}
}

func testSolidSDF(t *testing.T, s solidSDF) {
	delta := s.Max().Sub(s.Min()).Norm() / 100.0
	mesh := MarchingCubesSearch(s, delta, 4)
	meshSDF := MeshToSDF(mesh)

	for i := 0; i < 1000; i++ {
		scale := s.Min().Sub(s.Max()).Scale(0.5)
		center := s.Min().Mid(s.Max())
		c := NewCoord3DRandNorm().Mul(scale).Add(center)

		sdf1 := meshSDF.SDF(c)
		sdf2 := s.SDF(c)
		if math.Abs(sdf1-sdf2) > delta*2 {
			t.Errorf("mismatched SDF: expected %f but got %f (solid %v)", sdf1, sdf2,
				s)
		}
	}
}

func TestRectCollider(t *testing.T) {
	for i := 0; i < 10; i++ {
		c1 := NewCoord3DRandNorm()
		c2 := NewCoord3DRandNorm()
		testSolidColliderSDF(t, &Rect{
			MinVal: c1.Min(c2),
			MaxVal: c1.Max(c2),
		})
	}
}

func TestSphereCollider(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidColliderSDF(t, &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func TestCylinderColliderSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidColliderSDF(t, &Cylinder{
			P1:     Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			P2:     Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Radius: math.Abs(rand.NormFloat64()),
		})
	}
}

func testSolidColliderSDF(t *testing.T, sc solidColliderSDF) {
	for i := 0; i < 1000; i++ {
		ray := &Ray{
			Origin:    NewCoord3DRandNorm(),
			Direction: NewCoord3DRandNorm(),
		}
		testSolidColliderSDFRay(t, sc, ray)
	}

	for i := 0; i < 10000; i++ {
		c := NewCoord3DRandNorm()
		r := math.Abs(rand.NormFloat64())
		sdf := sc.SDF(c)
		if math.Abs(sdf) < 1e-8 {
			continue
		}
		collides := sc.SphereCollision(c, r)
		if collides != (math.Abs(sdf) < r) {
			t.Errorf("collides(%f)=%v but sdf=%f", r, collides, sdf)
		}
	}
}

func testSolidColliderSDFRay(t *testing.T, sc solidColliderSDF, ray *Ray) {
	ground := &SolidCollider{
		Solid:               sc,
		Epsilon:             0.005,
		BisectCount:         64,
		NormalSamples:       16,
		NormalBisectEpsilon: 1e-5,
	}

	var actualCollisions []RayCollision
	sc.RayCollisions(ray, func(rc RayCollision) {
		actualCollisions = append(actualCollisions, rc)
		p := ray.Origin.Add(ray.Direction.Scale(rc.Scale))
		if math.Abs(sc.SDF(p)) > 1e-8 {
			t.Error("ray collision not on boundary")
		}
	})

	sort.Slice(actualCollisions, func(i, j int) bool {
		return actualCollisions[i].Scale < actualCollisions[j].Scale
	})
	first, ok := sc.FirstRayCollision(ray)
	if ok != (len(actualCollisions) > 0) {
		t.Error("ray collision count mismatches FirstRayCollision")
	}
	if ok {
		expFirst := actualCollisions[0]
		if math.Abs(first.Scale-expFirst.Scale) > 1e-8 ||
			first.Normal.Dot(expFirst.Normal) < 1-1e-8 {
			t.Errorf("unexpected first collision: expected %v but got %v", expFirst, first)
		}
	}

	// Make sure the collider isn't under-reporting
	// collisions.
	ground.RayCollisions(ray, func(rc RayCollision) {
		for _, ac := range actualCollisions {
			if math.Abs(ac.Scale-rc.Scale) < 1e-8 {
				return
			}
		}
		t.Error("unreported ray collision detected", rc, actualCollisions)
	})
}
