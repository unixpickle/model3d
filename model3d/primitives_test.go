package model3d

import (
	"math"
	"math/rand"
	"os"
	"testing"
)

func TestTriangleAreaGradient(t *testing.T) {
	for i := 0; i < 1000; i++ {
		tri := &Triangle{
			Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
			Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()},
		}
		actual := tri.AreaGradient()
		expected := approxTriangleAreaGradient(tri)
		for i, a := range actual {
			e := expected[i]
			if a.Dist(e) > 1e-5 {
				t.Error("invalid gradient", a, "expected", e)
			}
		}
	}
}

func approxTriangleAreaGradient(t *Triangle) *Triangle {
	var grad Triangle
	for i, p := range t {
		var coordGrad [3]float64
		for j := 0; j < 3; j++ {
			arr := p.Array()
			arr[j] += 1e-8
			t1 := *t
			t1[i] = NewCoord3DArray(arr)
			area1 := t1.Area()
			arr[j] -= 2e-8
			t1[i] = NewCoord3DArray(arr)
			area2 := t1.Area()
			coordGrad[j] = (area1 - area2) / 2e-8
		}
		grad[i] = NewCoord3DArray(coordGrad)
	}
	return &grad
}

func TestTriangleDist(t *testing.T) {
	for i := 0; i < 100; i++ {
		tri := &Triangle{NewCoord3DRandNorm(), NewCoord3DRandNorm(), NewCoord3DRandNorm()}
		for j := 0; j < 10; j++ {
			c := NewCoord3DRandNorm()
			approx := approxTriangleDist(tri, c)
			actual := tri.Dist(c)
			if math.Abs(approx-actual) > 1e-5 {
				t.Fatalf("expected %f but got %f", approx, actual)
			}
		}
	}
}

func approxTriangleDist(t *Triangle, c Coord3D) float64 {
	min := 0.0
	max := t[0].Dist(c)
	for i := 0; i < 64; i++ {
		mid := (min + max) / 2
		collides := t.SphereCollision(c, mid)
		if collides {
			max = mid
		} else {
			min = mid
		}
	}
	return (min + max) / 2
}

func TestSegmentEntersSphere(t *testing.T) {
	center := XYZ(1, 2, 3)
	radius := 0.5

	segments := [][2]Coord3D{
		{
			{X: -1, Y: 2.4, Z: 3},
			{X: 2, Y: 2.4, Z: 3},
		},
		{
			{X: -1, Y: 2.6, Z: 3},
			{X: 2, Y: 2.6, Z: 3},
		},
	}
	insides := []bool{
		true,
		false,
	}

	for i, seg := range segments {
		actual := segmentEntersSphere(seg[0], seg[1], center, radius)
		expected := insides[i]
		if actual != expected {
			t.Errorf("test %d: expected %v but got %v", i, expected, actual)
		}
	}

	for i := 0; i < 100; i++ {
		c := XYZ(rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64())

		// Vector from center of sphere to line.
		v := XYZ(rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64())
		v = v.Scale(1 / v.Norm())

		// Direction of the line should be orthogonal to
		// the vector from the center of the sphere.
		v1 := XYZ(rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64())
		v1 = v1.Scale(1 / v1.Norm())
		v1 = v1.Add(v.Scale(-v1.Dot(v)))

		r := math.Abs(rand.NormFloat64()) + 1e-2

		v = v.Scale(rand.NormFloat64())

		p1 := c.Add(v).Add(v1.Scale(10 * r))
		p2 := c.Add(v).Add(v1.Scale(-10 * r))

		actual := segmentEntersSphere(p1, p2, c, r)
		expected := v.Norm() < r

		if actual != expected {
			t.Errorf("random case mismatch: got %v but expected %v", actual, expected)
		}
	}
}

func TestTriangleCollisions(t *testing.T) {
	t.Run("RandomPairs", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			t1 := randomTriangle()
			t2 := randomTriangle()
			intersection := t1.TriangleCollisions(t2)
			if len(intersection) == 0 {
				continue
			}
			seg := intersection[0]
			for _, frac := range []float64{-0.1, 0, 0.1, 0.5, 0.9, 1, 1.1} {
				shouldContain := frac >= 0 && frac <= 1
				c := seg[0].Scale(frac).Add(seg[1].Scale(1 - frac))
				contains1 := triangleContains(t1, c)
				contains2 := triangleContains(t2, c)
				if (contains1 && contains2) != shouldContain {
					t.Fatal("incorrect containment for frac", frac)
				}
			}
		}
	})

	t.Run("SelfIntersections", func(t *testing.T) {
		mesh := NewMeshPolar(func(g GeoCoord) float64 {
			return 1
		}, 50)
		collider := MeshToCollider(mesh)
		mesh.Iterate(func(tri *Triangle) {
			if len(collider.TriangleCollisions(tri)) != 0 {
				t.Fatal("self collision")
			}
		})
	})
}

func TestTriangleCollisionMismatch(t *testing.T) {
	m := readNonIntersectingHook()

	flat := m.FlattenBase(0)
	flat1 := NewMesh()
	flat1.AddMesh(flat)

	i1 := flat.SelfIntersections()
	i2 := flat1.SelfIntersections()
	if i1 != i2 {
		t.Fatal("bad intersection count", i1, i2)
	}
}

func TestTriangleRectCollision(t *testing.T) {
	for i := 0; i < 1000; i++ {
		tri := randomTriangle()
		samplePoint := func() Coord3D {
			a, b := rand.Float64(), rand.Float64()
			b *= 1 - a
			if rand.Intn(2) == 0 {
				a, b = b, a
			}
			return tri[0].Scale(1 - (a + b)).Add(tri[1].Scale(a)).Add(tri[2].Scale(b))
		}

		rect := &Rect{
			MinVal: NewCoord3DRandNorm(),
		}
		rect.MaxVal = rect.MinVal.Add(NewCoord3DRandUniform())

		boundingRadius := rect.MinVal.Dist(rect.MaxVal.Mid(rect.MinVal))
		if !tri.SphereCollision(rect.MaxVal.Mid(rect.MinVal), boundingRadius) {
			if tri.RectCollision(rect) {
				t.Error("got rect collision outside of bounding sphere")
			}
		}

		for j := 0; j < 10; j++ {
			if rect.Contains(samplePoint()) {
				if !tri.RectCollision(rect) {
					t.Error("sampled point inside rect, but got no collision")
				}
			}
		}
	}
}

func BenchmarkTriangleRayCollision(b *testing.B) {
	t := &Triangle{
		XYZ(0.1, 0.1, 0),
		XYZ(1, 0.1, 0.1),
		XYZ(0.1, 1.0, 0.2),
	}
	ray := &Ray{
		Origin:    XYZ(0.2, 0.2, 0.3),
		Direction: Coord3D{Z: -1},
	}
	for i := 0; i < b.N; i++ {
		t.FirstRayCollision(ray)
	}
}

func BenchmarkTriangleRectCollision(b *testing.B) {
	t := &Triangle{
		XYZ(0.1, 0.1, 0),
		XYZ(1, 0.1, 0.1),
		XYZ(0.1, 1.0, 0.2),
	}
	rects := []*Rect{
		&Rect{
			MinVal: XYZ(0.09, 0.09, -1),
			MaxVal: XYZ(1.1, 1.1, 0.3),
		},
		&Rect{
			MinVal: XYZ(2.0, 0.09, -1),
			MaxVal: XYZ(2.1, 1.1, 0.3),
		},
	}
	for i := 0; i < b.N/2; i++ {
		for _, r := range rects {
			t.RectCollision(r)
		}
	}
}

// Load a 3D model that caused various bugs in the past.
func readNonIntersectingHook() *Mesh {
	r, err := os.Open("test_data/non_intersecting_hook.stl")
	if err != nil {
		panic(err)
	}
	defer r.Close()
	tris, err := ReadSTL(r)
	if err != nil {
		panic(err)
	}
	return NewMeshTriangles(tris)
}

func randomTriangle() *Triangle {
	t := &Triangle{}
	for i := range t {
		t[i] = NewCoord3DRandNorm()
	}
	return t
}

func triangleContains(t *Triangle, c Coord3D) bool {
	v1 := t[1].Sub(t[0])
	v2 := t[2].Sub(t[0])
	combo := (NewMatrix3Columns(v1, v2, t.Normal())).Inverse().MulColumn(c.Sub(t[0]))
	return math.Abs(combo.Z) < 1e-8 && combo.X > -1e-8 && combo.Y > -1e-8 &&
		combo.X+combo.Y <= 1+1e-8
}
