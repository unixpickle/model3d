package model3d

import (
	"math"
	"math/rand"
	"reflect"
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
			if !reflect.DeepEqual(a, x) {
				t.Error("collision mismatch")
			}
		}

		// Check the barycentric coordinates.
		for _, a := range actual {
			tc := a.Extra.(*TriangleCollision)
			var baryCoord Coord3D
			for i, c := range tc.Triangle {
				baryCoord = baryCoord.Add(c.Scale(tc.Barycentric[i]))
			}
			actualCoord := ray.Origin.Add(ray.Direction.Scale(a.Scale))
			if actualCoord.Dist(baryCoord) > 1e-8 {
				t.Errorf("invalid barycentric coordinates: %v", tc.Barycentric)
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

func BenchmarkMeshRayCollisionsRect(b *testing.B) {
	mesh := NewMeshRect(Coord3D{X: -0.3, Y: -0.4, Z: -0.2}, Coord3D{X: 0.4, Y: 0.35, Z: 0.19})
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

func BenchmarkMeshRayCollisionsComplex(b *testing.B) {
	solid := JoinedSolid{
		&Sphere{Center: Coord3D{X: 1.0, Y: 0.7, Z: 0.1}, Radius: 0.2},
		&Sphere{Center: Coord3D{X: 1.3, Y: 0.75, Z: 0}, Radius: 0.22},
		&Sphere{Center: Coord3D{X: 0.9, Y: 0.2, Z: 0.1}, Radius: 0.3},
		&Cylinder{P2: Coord3D{X: 3, Y: 3, Z: 3}, Radius: 0.1},
	}
	mesh := MarchingCubes(solid, 0.04)

	// Make the mesh 4x more triangles without having to
	// scan the entire volume more densely.
	subdiv := NewSubdivider()
	subdiv.AddFiltered(mesh, func(p1, p2 Coord3D) bool {
		return true
	})
	subdiv.Subdivide(mesh, func(p1, p2 Coord3D) Coord3D {
		return p1.Mid(p2)
	})

	runCollider := func(b *testing.B, collider Collider) {
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

	b.Run("Balanced", func(b *testing.B) {
		collider := MeshToCollider(mesh)
		runCollider(b, collider)
	})

	b.Run("Unbalanced", func(b *testing.B) {
		collider := BVHToCollider(NewBVHAreaDensity(mesh.TriangleSlice()))
		runCollider(b, collider)
	})
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
