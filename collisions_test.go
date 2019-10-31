package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestEmptyColliders(t *testing.T) {
	// Make sure we don't get a crash from this.
	GroupTriangles(nil)

	if MeshToCollider(NewMesh()).SphereCollision(Coord3D{}, 1000) {
		t.Error("unexpected collision")
	}
}

func TestSegmentEntersSphere(t *testing.T) {
	center := Coord3D{X: 1, Y: 2, Z: 3}
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
		c := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(), Z: rand.NormFloat64()}

		// Vector from center of sphere to line.
		v := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(), Z: rand.NormFloat64()}
		v = v.Scale(1 / v.Norm())

		// Direction of the line should be orthogonal to
		// the vector from the center of the sphere.
		v1 := Coord3D{X: rand.NormFloat64(), Y: rand.NormFloat64(), Z: rand.NormFloat64()}
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

func randomTriangle() *Triangle {
	t := &Triangle{}
	for i := range t {
		t[i] = Coord3D{rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()}
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

func BenchmarkMeshToCollider(b *testing.B) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1
	}, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MeshToCollider(mesh)
	}
}

func BenchmarkMeshSphereCollisions(b *testing.B) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1
	}, 50)
	collider := MeshToCollider(mesh)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collider.SphereCollision(Coord3D{
			X: rand.NormFloat64(),
			Y: rand.NormFloat64(),
			Z: rand.NormFloat64(),
		}, math.Abs(rand.NormFloat64()))
	}
}
