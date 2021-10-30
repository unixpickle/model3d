package model3d

import (
	"math/rand"
	"testing"
)

func TestMarchingCubesDeterminism(t *testing.T) {
	table1 := mcLookupTable()
	for i := 0; i < 10; i++ {
		table2 := mcLookupTable()
		for key, value1 := range table1 {
			value2 := table2[key]
			if len(value1) != len(value2) {
				t.Fatal("mismatched triangle count")
			}
			for j, tri := range value1 {
				if value2[j] != tri {
					t.Fatal("mismatched triangle:", value1, value2)
				}
			}
		}
	}
}

func TestMarchingCubesRandom(t *testing.T) {
	for _, name := range []string{"Normal", "Search"} {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				var mesh *Mesh
				if name == "Normal" {
					mesh = MarchingCubes(randomSolid{}, 0.1)
				} else {
					mesh = MarchingCubesSearch(randomSolid{}, 0.1, 2)
				}
				MustValidateMesh(t, mesh, true)
			}
		})
	}
}

func TestMarchingCubesFilter(t *testing.T) {
	t.Run("Sphere", func(t *testing.T) {
		mesh := NewMeshIcosphere(XYZ(0.1, 0.3, -0.2), 1.0, 20)
		collider := MeshToCollider(mesh)
		solid := NewColliderSolid(collider)
		base := MarchingCubes(solid, 0.1)
		rc := MarchingCubesFilter(solid, collider.RectCollision, 0.1)
		if !meshesEqual(base, rc) {
			t.Fatal("meshes should be equal")
		}
	})
	t.Run("Boxes", func(t *testing.T) {
		mesh := NewMesh()
		mesh.AddMesh(NewMeshRect(XYZ(-1, -1, -1), XYZ(0, 0, 0)))
		mesh.AddMesh(NewMeshRect(XYZ(0.1, 0, 0), XYZ(1, 1, 1)))
		collider := MeshToCollider(mesh)
		solid := NewColliderSolid(collider)
		base := MarchingCubes(solid, 0.1)
		rc := MarchingCubesFilter(solid, collider.RectCollision, 0.1)
		if !meshesEqual(base, rc) {
			t.Fatal("meshes should be equal")
		}
	})
}

func TestMarchingCubesC2F(t *testing.T) {
	t.Run("Sphere", func(t *testing.T) {
		solid := &Sphere{Center: XYZ(0.1, 0.3, -0.2), Radius: 1.0}
		mesh1 := MarchingCubesSearch(solid, 0.05, 8)
		mesh2 := MarchingCubesC2F(solid, 0.1, 0.05, 0, 8)
		if !meshesEqual(mesh1, mesh2) {
			t.Fatal("meshes should be equal")
		}
	})

	t.Run("Boxes", func(t *testing.T) {
		mesh := NewMesh()
		mesh.AddMesh(NewMeshRect(XYZ(-1, -1, -1), XYZ(0, 0, 0)))
		mesh.AddMesh(NewMeshRect(XYZ(0.3, 0, 0), XYZ(1, 1, 1)))
		collider := MeshToCollider(mesh)
		solid := NewColliderSolid(collider)

		mesh1 := MarchingCubesSearch(solid, 0.1, 8)
		mesh2 := MarchingCubesC2F(solid, 0.2, 0.1, 0, 8)
		MustValidateMesh(t, mesh1, true)
		MustValidateMesh(t, mesh2, true)
		if !meshesEqual(mesh1, mesh2) {
			t.Fatal("meshes should be equal")
		}
	})
}

func BenchmarkMarchingCubes(b *testing.B) {
	runBench := func(b *testing.B, iters int) {
		b.Run("Cylinder", func(b *testing.B) {
			solid := &CylinderSolid{
				P1:     XYZ(1, 2, 3),
				P2:     XYZ(3, 1, 4),
				Radius: 0.5,
			}
			for i := 0; i < b.N; i++ {
				MarchingCubesSearch(solid, 0.025, iters)
			}
		})
		b.Run("Boxes", func(b *testing.B) {
			mesh := NewMesh()
			mesh.AddMesh(NewMeshRect(XYZ(-1, -1, -1), XYZ(0, 0, 0)))
			mesh.AddMesh(NewMeshRect(XYZ(0.1, 0, 0), XYZ(1, 1, 1)))
			solid := NewColliderSolid(MeshToCollider(mesh))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				MarchingCubesSearch(solid, 0.025, iters)
			}
		})
	}
	b.Run("Search", func(b *testing.B) {
		runBench(b, 8)
	})
	runBench(b, 0)
}

func BenchmarkMarchingCubesFilter(b *testing.B) {
	runBench := func(b *testing.B, iters int) {
		b.Run("Boxes", func(b *testing.B) {
			mesh := NewMesh()
			mesh.AddMesh(NewMeshRect(XYZ(-1, -1, -1), XYZ(0, 0, 0)))
			mesh.AddMesh(NewMeshRect(XYZ(0.1, 0, 0), XYZ(1, 1, 1)))
			collider := MeshToCollider(mesh)
			solid := NewColliderSolid(collider)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				MarchingCubesSearchFilter(solid, collider.RectCollision, 0.025, iters)
			}
		})
	}
	b.Run("Search", func(b *testing.B) {
		runBench(b, 8)
	})
	runBench(b, 0)
}

type randomSolid struct{}

func (r randomSolid) Min() Coord3D {
	return Coord3D{}
}

func (r randomSolid) Max() Coord3D {
	return XYZ(1, 1, 1)
}

func (r randomSolid) Contains(c Coord3D) bool {
	return InBounds(r, c) && rand.Intn(4) == 0
}
