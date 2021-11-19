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

type pointNormalSDF interface {
	PointSDF
	NormalSDF(Coord3D) (Coord3D, float64)
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
		r := &Rect{
			MinVal: c1.Min(c2),
			MaxVal: c1.Max(c2).Add(XYZ(0.1, 0.1, 0.1)),
		}
		testMeshSDF(t, r, NewMeshRect(r.MinVal, r.MaxVal), 1e-5)
		testPointSDFConsistency(t, r)
		testNormalSDFConsistency(
			t,
			r,
			false,
			r.MinVal.Mid(r.MaxVal),
			r.MinVal.Mid(r.MaxVal).Add(X(0.01)),
			r.MinVal.Mid(r.MaxVal).Add(X(-0.01)),
			r.MinVal.Mid(r.MaxVal).Add(Y(0.01)),
			r.MinVal.Mid(r.MaxVal).Add(Y(-0.01)),
			r.MinVal.Mid(r.MaxVal).Add(Z(0.01)),
			r.MinVal.Mid(r.MaxVal).Add(Z(-0.01)),
			r.MinVal.Mid(r.MaxVal).Add(X(0.2)),
			r.MinVal.Mid(r.MaxVal).Add(X(-0.2)),
			r.MinVal.Mid(r.MaxVal).Add(Y(0.2)),
			r.MinVal.Mid(r.MaxVal).Add(Y(-0.2)),
			r.MinVal.Mid(r.MaxVal).Add(Z(0.2)),
			r.MinVal.Mid(r.MaxVal).Add(Z(-0.2)),
		)
	}
}

func TestSphereSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		sphere := &Sphere{
			Center: NewCoord3DRandNorm(),
			Radius: math.Abs(rand.NormFloat64()) + 0.1,
		}
		testSolidSDF(t, sphere)
		testPointSDFConsistency(t, sphere)
		testNormalSDFConsistency(t, sphere, true)
	}
}

func TestCylinderSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		p1 := NewCoord3DRandUnit()
		p2 := NewCoord3DRandUnit()
		cyl := &Cylinder{
			P1:     p1,
			P2:     p2,
			Radius: math.Abs(rand.NormFloat64()) + 0.1,
		}
		epsilon := 0.04
		numStops := int(math.Ceil(2 * math.Pi * cyl.Radius / epsilon))
		mesh := NewMeshCylinder(cyl.P1, cyl.P2, cyl.Radius, numStops)
		testMeshSDF(t, cyl, mesh, epsilon)
		testPointSDFConsistency(t, cyl)

		b1, b2 := cyl.P2.Sub(cyl.P1).OrthoBasis()
		testNormalSDFConsistency(
			t,
			cyl,
			false,
			cyl.P1.Mid(cyl.P2),
			cyl.P1.Mid(cyl.P2).Add(b1),
			cyl.P1.Mid(cyl.P2).Add(b2),
			cyl.P1.Add(cyl.P2.Sub(cyl.P1).Scale(-1)),
			cyl.P1.Add(cyl.P2.Sub(cyl.P1).Scale(-1)).Add(b1.Scale(cyl.Radius/2)),
			cyl.P2.Add(cyl.P1.Sub(cyl.P2).Scale(-1)),
			cyl.P2.Add(cyl.P1.Sub(cyl.P2).Scale(-1)).Add(b1.Scale(cyl.Radius/2)),
		)
	}
}

func TestConeSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		p1 := NewCoord3DRandUnit()
		p2 := NewCoord3DRandUnit()
		if p1.Dist(p2) < 0.1 {
			i--
			continue
		}
		cone := &Cone{
			Base:   p1,
			Tip:    p2,
			Radius: math.Abs(rand.NormFloat64()) + 0.1,
		}
		epsilon := 0.04
		numStops := int(math.Ceil(2 * math.Pi * cone.Radius / epsilon))
		mesh := NewMeshCone(cone.Tip, cone.Base, cone.Radius, numStops)
		testMeshSDF(t, cone, mesh, epsilon)
	}
}

func TestTorusSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		torus := randomTorus()
		epsilon := 0.04
		innerStops := int(math.Ceil(2 * math.Pi * torus.InnerRadius / epsilon))
		outerStops := int(math.Ceil(2 * math.Pi * (torus.OuterRadius + torus.InnerRadius) / epsilon))
		mesh := NewMeshTorus(torus.Center, torus.Axis, torus.InnerRadius, torus.OuterRadius,
			innerStops, outerStops)
		testMeshSDF(t, torus, mesh, epsilon)
		x, y := torus.Axis.OrthoBasis()
		outerPoint := torus.Center.Add(
			x.Scale(math.Cos(2)).Add(y.Scale(math.Sin(2))).Scale(torus.OuterRadius),
		)
		checkPoints := []Coord3D{torus.Center, outerPoint, torus.Center.Add(XYZ(1e-16, -1e-16, 0))}
		testPointSDFConsistency(t, torus, checkPoints...)
		testNormalSDFConsistency(t, torus, true, checkPoints...)
	}
}

func testMeshSDF(t *testing.T, s SDF, m *Mesh, epsilon float64) {
	meshSDF := MeshToSDF(m)
	for i := 0; i < 1000; i++ {
		scale := s.Min().Sub(s.Max()).Scale(0.5)
		center := s.Min().Mid(s.Max())
		c := NewCoord3DRandNorm().Mul(scale).Add(center)

		sdf1 := meshSDF.SDF(c)
		sdf2 := s.SDF(c)
		if math.Abs(sdf1-sdf2) > epsilon {
			t.Errorf("mismatched SDF: expected %f but got %f (solid %v)", sdf1, sdf2,
				s)
		}
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
		if math.Abs(sdf1-sdf2) > delta*3 {
			t.Errorf("mismatched SDF: expected %f but got %f (solid %v)", sdf1, sdf2,
				s)
		}
	}
}

func testPointSDFConsistency(t *testing.T, p PointSDF, checkPoints ...Coord3D) {
	rad := p.Min().Dist(p.Max())
	min := p.Min().AddScalar(-rad)
	max := p.Min().AddScalar(rad)
	checkPoint := func(c Coord3D) {
		point, sdf := p.PointSDF(c)
		if math.Abs(math.Abs(sdf)-point.Dist(c)) > 1e-5 {
			t.Errorf("mismatched SDF and point distance: %v (sdf=%f) (dist=%f)",
				c, sdf, point.Dist(c))
		}
		if math.Abs(p.SDF(point)) > 1e-5 {
			t.Errorf("nearest point %v should have 0 SDF, but got %f", point, p.SDF(point))
		}
	}
	for _, c := range checkPoints {
		checkPoint(c)
	}
	for i := 0; i < 100; i++ {
		checkPoint(NewCoord3DRandBounds(min, max))
	}
}

func testNormalSDFConsistency(t *testing.T, p pointNormalSDF, checkRandom bool, checkPoints ...Coord3D) {
	rad := p.Min().Dist(p.Max())
	min := p.Min().AddScalar(-rad)
	max := p.Min().AddScalar(rad)
	checkPoint := func(c Coord3D) {
		point, sdf1 := p.PointSDF(c)
		normal, sdf2 := p.NormalSDF(c)
		if math.Abs(sdf1-sdf2) > 1e-5 {
			t.Errorf("inconsistent SDF values: %f and %f", sdf1, sdf2)
		}
		delta := normal.Scale(rad * 1e-5)
		outside := p.SDF(point.Add(delta))
		inside := p.SDF(point.Sub(delta))
		if outside > 0 || inside < 0 {
			t.Errorf("unexpected SDFs when moving by normal/-normal: %f, %f", outside, inside)
		}
	}
	for _, c := range checkPoints {
		checkPoint(c)
	}
	if checkRandom {
		for i := 0; i < 1000; i++ {
			checkPoint(NewCoord3DRandBounds(min, max))
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

func TestTorusColliderSDF(t *testing.T) {
	for i := 0; i < 10; i++ {
		testSolidColliderSDF(t, randomTorus())
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

func randomTorus() *Torus {
	outer := math.Abs(rand.NormFloat64()) + 0.1
	inner := math.Abs(rand.NormFloat64()) + 0.1
	if inner > outer {
		outer, inner = inner, outer
	}
	return &Torus{
		Axis:        NewCoord3DRandNorm(),
		Center:      NewCoord3DRandNorm(),
		OuterRadius: outer,
		InnerRadius: inner,
	}
}
