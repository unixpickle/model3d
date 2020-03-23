package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestSolidToMeshSphere(t *testing.T) {
	solid := &SphereSolid{
		Radius: 0.5,
		Center: Coord3D{X: 1},
	}
	mesh := SolidToMesh(solid, 0.025, 0, 0, 0)

	// Margin of error for mesh.
	epsilon := 0.1

	t.Run("Singular", func(t *testing.T) {
		if mesh.NeedsRepair() {
			t.Fatal("mesh needs repair")
		}
		if len(mesh.SingularVertices()) > 0 {
			t.Fatal("singular vertices detected")
		}
	})

	t.Run("Boundary", func(t *testing.T) {
		mesh.Iterate(func(tri *Triangle) {
			for _, c := range tri {
				if math.Abs(c.Dist(solid.Center)-solid.Radius) > epsilon {
					t.Fatalf("vertex %v too far from solid bounds", c)
				}
			}
		})
	})

	t.Run("Collisions", func(t *testing.T) {
		collider := MeshToCollider(mesh)
		solid1 := NewColliderSolid(collider)
		for i := 0; i < 100; i++ {
			c := NewCoord3DRandNorm().Add(solid.Center)
			if math.Abs(c.Dist(solid.Center)-solid.Radius) < epsilon {
				// The coordinate is on the boundary.
				continue
			}
			if solid.Contains(c) != solid1.Contains(c) {
				t.Fatalf("containment mismatch for %v", c)
			}
		}
	})
}

func TestSolidToMeshSingularEdgesSimple(t *testing.T) {
	mesh := SolidToMesh(simpleSingular{}, 0.5, 0, 0, 0)
	if mesh.NeedsRepair() {
		t.Error("mesh needs repair")
	}
	if len(mesh.SingularVertices()) > 0 {
		t.Error("singular vertices detected")
	}
}

type simpleSingular struct{}

func (s simpleSingular) Min() Coord3D {
	return Coord3D{}
}

func (s simpleSingular) Max() Coord3D {
	return Coord3D{X: 1, Y: 1, Z: 1}
}

func (s simpleSingular) Contains(c Coord3D) bool {
	if c.Min(s.Min()) != s.Min() || c.Max(s.Max()) != s.Max() {
		return false
	}
	return (c.X < 0.1 && c.Y < 0.1 && c.Z < 0.1) || (c.X > 0.9 && c.Y > 0.9 && c.Z < 0.1)
}

func TestSolidToMeshSingularitiesRandom(t *testing.T) {
	for i := 0; i < 1000; i++ {
		mesh := SolidToMesh(randomSolid{}, 0.19, 0, 0, 0)
		if mesh.NeedsRepair() {
			t.Fatal("singular edge detected")
		}
		if len(mesh.SingularVertices()) > 0 {
			t.Fatal("singular vertices detected")
		}
	}
}

type randomSolid struct{}

func (r randomSolid) Min() Coord3D {
	return Coord3D{}
}

func (r randomSolid) Max() Coord3D {
	return Coord3D{X: 1, Y: 1, Z: 1}
}

func (r randomSolid) Contains(c Coord3D) bool {
	return InBounds(r, c) && rand.Intn(4) == 0
}

func TestSolidToMeshSingularities(t *testing.T) {
	// Create an adversarial pumpkin mesh.
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat)
	}, 30)
	mesh.Iterate(func(t *Triangle) {
		t1 := *t
		for i, c := range t1 {
			t1[i] = c.Scale(0.9)
		}
		t1[0], t1[1] = t1[1], t1[0]
		mesh.Add(&t1)
	})
	collider := MeshToCollider(mesh)
	solid := NewColliderSolid(collider)
	mesh = SolidToMesh(solid, 0.1, 0, 0, 0)
	if mesh.NeedsRepair() {
		t.Error("mesh needs repair")
	}
	if len(mesh.SingularVertices()) > 0 {
		t.Error("singular vertices detected")
	}
}

func TestSolidToMeshSingularVerticesSimple(t *testing.T) {
	mesh := SolidToMesh(simpleSingularVertex{}, 0.5, 0, 0, 0)
	if len(mesh.SingularVertices()) > 0 {
		t.Fatal("singular vertices detected")
	}
}

type simpleSingularVertex struct{}

func (s simpleSingularVertex) Min() Coord3D {
	return Coord3D{}
}

func (s simpleSingularVertex) Max() Coord3D {
	return Coord3D{X: 1, Y: 1, Z: 1}
}

func (s simpleSingularVertex) Contains(c Coord3D) bool {
	if c.Min(s.Min()) != s.Min() || c.Max(s.Max()) != s.Max() {
		return false
	}
	return (c.X < 0.1 && c.Y < 0.1 && c.Z < 0.1) || (c.X > 0.9 && c.Y > 0.9 && c.Z < 0.9)
}

func BenchmarkSolidToMesh(b *testing.B) {
	for _, direct := range []bool{false, true} {
		name := "Subdivision"
		divisions := 2
		resolution := 0.1
		if direct {
			name = "Direct"
			divisions = 0
			resolution /= 4
		}
		b.Run(name, func(b *testing.B) {
			solid := &CylinderSolid{
				P1:     Coord3D{X: 1, Y: 2, Z: 3},
				P2:     Coord3D{X: 3, Y: 1, Z: 4},
				Radius: 0.5,
			}
			for i := 0; i < b.N; i++ {
				SolidToMesh(solid, resolution, divisions, 0, 0)
			}
		})
	}
}
