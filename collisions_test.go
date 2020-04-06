package model3d

import (
	"math"
	"math/rand"
	"os"
	"sort"
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

func TestMeshRayCollisions(t *testing.T) {
	// Small mesh for fast brute force.
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 0.5 + 0.1*math.Cos(g.Lon)
	}, 10)

	collider := MeshToCollider(mesh)
	for i := 0; i < 1000; i++ {
		ray := &Ray{
			Origin:    NewCoord3DRandNorm(),
			Direction: NewCoord3DRandUnit(),
		}
		var actual []RayCollision
		collider.RayCollisions(ray, func(c RayCollision) {
			actual = append(actual, c)
		})
		var expected []RayCollision
		mesh.Iterate(func(t *Triangle) {
			coll, ok := t.FirstRayCollision(ray)
			if ok {
				expected = append(expected, coll)
			}
		})

		if len(actual) != len(expected) {
			t.Fatal("incorrect number of collisions")
		}

		for _, s := range [][]RayCollision{actual, expected} {
			sort.Slice(s, func(i, j int) bool {
				return s[i].Scale < s[j].Scale
			})
		}

		for i, a := range actual {
			x := expected[i]
			if a != x {
				t.Error("collision mismatch")
			}
		}
	}
}

func TestMeshRayCollisionsConsistency(t *testing.T) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 0.5 + 0.1*math.Cos(g.Lon)
	}, 100)

	collider := MeshToCollider(mesh)
	for i := 0; i < 1000; i++ {
		ray := &Ray{
			Origin:    NewCoord3DRandNorm(),
			Direction: NewCoord3DRandUnit(),
		}
		var collisions []RayCollision
		count := collider.RayCollisions(ray, func(c RayCollision) {
			collisions = append(collisions, c)
		})
		if count != len(collisions) {
			t.Fatal("callback not called for every collision")
		}
		sort.Slice(collisions, func(i, j int) bool {
			return collisions[i].Scale < collisions[j].Scale
		})
		firstCollision, collides := collider.FirstRayCollision(ray)
		if collides != (len(collisions) > 0) {
			t.Error("mismatched collision reports")
		} else if collides && math.Abs(firstCollision.Scale-collisions[0].Scale) > 1e-8 {
			t.Error("mismatched collision scales for closest collision")
		}
	}
}

func TestSolidCollider(t *testing.T) {
	// Create a non-trivial, non-convex solid.
	solid := JoinedSolid{
		&CylinderSolid{
			P1:     Coord3D{X: 0.3, Y: 0.3, Z: -1},
			P2:     Coord3D{X: -0.3, Y: -0.3, Z: 1},
			Radius: 0.3,
		},
		&SphereSolid{
			Center: Coord3D{X: 0.1},
			Radius: 0.3,
		},
	}

	// Use a mesh as our ground-truth collider.
	mesh := MarchingCubesSearch(solid, 0.005, 8)
	ground := MeshToCollider(mesh)

	collider := &SolidCollider{
		Solid:               solid,
		Epsilon:             0.005,
		BisectCount:         32,
		NormalSamples:       16,
		NormalBisectEpsilon: 1e-5,
	}

	verifiedCollisions := func(c Collider, r *Ray) ([]RayCollision, bool) {
		var result []RayCollision
		c.RayCollisions(r, func(rc RayCollision) {
			result = append(result, rc)
		})
		sort.Slice(result, func(i, j int) bool {
			return result[i].Scale < result[j].Scale
		})
		scaleDelta := 0.02 / r.Direction.Norm()
		lastScale := 0.0
		for _, x := range result {
			if x.Scale-lastScale < scaleDelta {
				// Collisions are too close together, so
				// neither the mesh nor the SolidCollider
				// are expected to be accurate.
				return nil, false
			}
			lastScale = x.Scale
		}
		return result, true
	}

	for i := 0; i < 10000; i++ {
		ray := &Ray{
			Origin: NewCoord3DRandNorm().Scale(0.5),
			// Explicitly test non-unit directions.
			Direction: NewCoord3DRandNorm(),
		}

		if i == 0 {
			// Special ray that broke the code in the past colliding
			// very close to the bounding box.
			ray.Origin = Coord3D{-0.20424398336871702, -0.14223091122768425, 0.6248593138047999}
			ray.Direction = Coord3D{-1.592256851550497, -0.6710011000343341, 1.2010483574169686}
		}

		actual, ok := verifiedCollisions(collider, ray)
		if !ok {
			continue
		}
		expected, ok := verifiedCollisions(ground, ray)
		if !ok {
			continue
		}
		if len(actual) != len(expected) {
			t.Error("intersection count mismatch: expected", len(expected), "but got", len(actual),
				"=> expected:", expected, "actual:", actual)
		} else {
			for i, x := range expected {
				a := actual[i]
				if math.Abs(x.Scale-a.Scale) > 0.01 || x.Normal.Dot(a.Normal) < 0 {
					t.Error("expected", expected, "but got", actual)
				}
			}
		}
	}
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

func BenchmarkMeshFirstRayCollisions(b *testing.B) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1
	}, 50)
	collider := MeshToCollider(mesh)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collider.FirstRayCollision(&Ray{
			Origin:    NewCoord3DRandNorm(),
			Direction: NewCoord3DRandUnit(),
		})
	}
}

func BenchmarkMeshRayCollisions(b *testing.B) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1
	}, 50)
	collider := MeshToCollider(mesh)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collider.RayCollisions(&Ray{
			Direction: Coord3D{
				X: rand.NormFloat64(),
				Y: rand.NormFloat64(),
				Z: rand.NormFloat64(),
			},
		}, nil)
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

func BenchmarkMeshTriangleCollisions(b *testing.B) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1
	}, 50)
	collider := MeshToCollider(mesh)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collider.TriangleCollisions(randomTriangle())
	}
}
