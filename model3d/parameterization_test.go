package model3d

import (
	"testing"

	"github.com/unixpickle/model3d/model2d"
)

func TestMeshToPlaneGraphs(t *testing.T) {
	t.Run("Square", func(t *testing.T) {
		m := NewMeshTriangles(TriangulateFace([]Coord3D{
			XYZ(0, 0, 0),
			XYZ(1, 0, 0),
			XYZ(1, 1, 0),
			XYZ(0, 1, 0),
		}))
		m1 := MeshToPlaneGraphs(m)
		if len(m1) != 1 {
			t.Fatalf("expected 1 mesh but got %d", len(m1))
		}
		if !meshesEqual(m, m1[0]) {
			t.Error("unequal meshes")
		}
	})

	testGenericShape := func(t *testing.T, m *Mesh) {
		splitUp := MeshToPlaneGraphs(m)
		joined := NewMesh()
		for _, subMesh := range splitUp {
			mustHaveSingleBoundary(t, subMesh)
			subMesh.Iterate(func(tri *Triangle) {
				if joined.Contains(tri) {
					t.Fatal("triangle duplicated across results")
				}
				joined.Add(tri)
			})
		}
		if !meshesEqual(joined, m) {
			t.Fatal("resulting meshes do not recombine correctly")
		}
	}

	t.Run("Sphere", func(t *testing.T) {
		m := NewMeshIcosphere(Origin, 1, 8)
		if n := len(MeshToPlaneGraphs(m)); n != 2 {
			t.Fatalf("expected exactly two meshes, but got %d", n)
		}
		testGenericShape(t, m)
	})
	t.Run("Torus", func(t *testing.T) {
		testGenericShape(t, NewMeshTorus(Origin, Z(1), 0.1, 0.5, 10, 30))
	})
	t.Run("Random", func(t *testing.T) {
		testGenericShape(t, MarchingCubesSearch(randomSolid{}, 0.05, 8))
	})
}

func mustHaveSingleBoundary(t *testing.T, m *Mesh) {
	vertexToNext := NewCoordMap[Coord3D]()

	var start Coord3D
	m.Iterate(func(t *Triangle) {
		for i := 0; i < 3; i++ {
			p1, p2 := t[i], t[(i+1)%3]
			if len(m.Find(p1, p2)) == 1 {
				vertexToNext.Store(p1, p2)
				start = p1
			}
		}
	})
	if vertexToNext.Len() == 0 {
		t.Error("the mesh did not contain any boundary edges")
		return
	}

	cur := vertexToNext.Value(start)
	vertexToNext.Delete(start)
	for cur != start {
		next, ok := vertexToNext.Load(cur)
		if !ok {
			t.Error("mesh has invalid boundary")
			return
		}
		vertexToNext.Delete(cur)
		cur = next
	}
	if vertexToNext.Len() > 0 {
		t.Error("mesh has multiple boundaries")
		return
	}
}

func TestBoundaryNonDegenerate(t *testing.T) {
	testFn := func(t *testing.T, fn func(*Mesh) *CoordMap[model2d.Coord]) {
		// Try a few times since MeshToPlaneGraphs is non-deterministic.
		for i := 0; i < 4; i++ {
			// Create a mesh which likely has whole triangles
			// on the boundary.
			m := NewMeshIcosphere(Origin, 1, 8)
			subMesh := MeshToPlaneGraphs(m)[0]

			boundary := fn(subMesh)
			subMesh.Iterate(func(t3d *Triangle) {
				var t2d [3]model2d.Coord
				all := true
				for i, c := range t3d {
					if v, ok := boundary.Load(c); ok {
						t2d[i] = v
					} else {
						all = false
						break
					}
				}
				if all {
					area := model2d.NewTriangle(t2d[0], t2d[1], t2d[2]).Area()
					if area == 0 {
						t.Fatalf("degenerate triangle: %v", t2d)
					}
				}
			})
		}
	}
	t.Run("Circle", func(t *testing.T) {
		testFn(t, CircleBoundary)
	})
	t.Run("PNormBoundary", func(t *testing.T) {
		testFn(t, func(m *Mesh) *CoordMap[model2d.Coord] {
			return PNormBoundary(m, 4)
		})
	})
}
