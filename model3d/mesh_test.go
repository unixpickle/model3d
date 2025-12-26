package model3d

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d/model2d"
)

func TestNewMeshIcosphere(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4} {
		mesh := NewMeshIcosphere(XYZ(-0.3, 0.4, -0.2), 0.315, n)
		if len(mesh.TriangleSlice()) != 20*n*n {
			t.Errorf("invalid triangle count: %d (expected %d)", len(mesh.TriangleSlice()), 20*n*n)
		}
		MustValidateMesh(t, mesh, true)
	}
}

func TestNewMeshRect(t *testing.T) {
	mesh := NewMeshRect(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19))
	MustValidateMesh(t, mesh, true)
}

func TestNewMeshCylinder(t *testing.T) {
	mesh := NewMeshCylinder(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19), 0.7, 20)
	MustValidateMesh(t, mesh, true)
}

func TestNewMeshCone(t *testing.T) {
	mesh := NewMeshCone(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19), 0.7, 20)
	MustValidateMesh(t, mesh, true)
}

func TestNewMeshConeSlice(t *testing.T) {
	mesh := NewMeshConeSlice(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19), 0.7, 0.3, 20)
	MustValidateMesh(t, mesh, true)
}

func TestNewMeshTorus(t *testing.T) {
	mesh := NewMeshTorus(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19), 0.3, 0.7, 20, 20)
	MustValidateMesh(t, mesh, true)
}

func TestNewMeshIcosahedron(t *testing.T) {
	mesh := NewMeshIcosahedron()
	MustValidateMesh(t, mesh, true)
}

func TestMeshDuplicateVertices(t *testing.T) {
	m := NewMesh()
	for i := 0; i < 2; i++ {
		tri := &Triangle{X(1), Y(1), X(1)}
		m.Add(tri)
		if n := len(m.Find(tri[0])); n != 1 {
			t.Errorf("iter %d: expected 1 but got: %d", i, n)
		}
		m.Remove(tri)
		if n := len(m.Find(tri[0])); n != 0 {
			t.Errorf("iter %d: expected 0 but got: %d", i, n)
		}
	}
}

func TestProfileMesh(t *testing.T) {
	mesh2d := model2d.NewMeshPolar(func(t float64) float64 {
		return 2 + math.Cos(t*10)
	}, 100)
	mesh3d := ProfileMesh(mesh2d, 0.1, 0.5)
	MustValidateMesh(t, mesh3d, true)
}

func TestVertexSlice(t *testing.T) {
	t1 := &Triangle{
		XY(0, 1),
		XY(1, 0),
		XY(1, 1),
	}
	t2 := &Triangle{
		XY(0, 1),
		XY(1, 0),
		XY(-1, 1),
	}
	mesh := NewMesh()
	mesh.Add(t1)
	if len(mesh.VertexSlice()) != 3 {
		t.Error("unexpected number of vertices")
	}
	mesh.Add(t2)
	if len(mesh.VertexSlice()) != 4 {
		t.Error("unexpected number of vertices")
	}
	mesh.Remove(t1)
	if len(mesh.VertexSlice()) != 3 {
		t.Error("unexpected number of vertices")
	}
	mesh.Remove(t2)
	if len(mesh.VertexSlice()) != 0 {
		t.Error("unexpected number of vertices")
	}
}

func BenchmarkMeshFind(b *testing.B) {
	mesh := NewMeshPolar(func(g GeoCoord) float64 {
		return 1
	}, 50)
	var vertices []Coord3D
	found := map[Coord3D]bool{}
	mesh.Iterate(func(t *Triangle) {
		for _, c := range t {
			if !found[c] {
				found[c] = true
				vertices = append(vertices, c)
			}
		}
	})

	// Make sure we don't do this lazily.
	mesh.getVertexToFace()

	// Pre-compute the random indices to prevent
	// random number generation from taking most
	// of the time.
	indices := make([]int, 0x1000)
	for i := range indices {
		indices[i] = rand.Intn(len(vertices))
	}

	b.Run("Single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mesh.Find(vertices[indices[i&0xfff]])
		}
	})
	b.Run("Double", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			j := indices[i&0xfff]
			if j+1 == len(vertices) {
				j -= 1
			}
			mesh.Find(vertices[j], vertices[j+1])
		}
	})
}

func BenchmarkMeshAddFindRemove(b *testing.B) {
	m1 := NewMeshPolar(func(g GeoCoord) float64 {
		return 1.0
	}, 30)
	m2 := NewMesh()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m1.Iterate(func(t *Triangle) {
			m2.Add(t)
			for _, c := range t {
				m2.Find(c)
			}
		})
		m1.Iterate(func(t *Triangle) {
			m2.Remove(t)
			for _, c := range t {
				m2.Find(c)
			}
		})
	}
}

func BenchmarkVertexToFace(b *testing.B) {
	mesh := NewMesh()
	for i := 0; i < 20000; i++ {
		mesh.Add(randomTriangle())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.getVertexToFace()
		var v2f *CoordToSlice[*Triangle]
		mesh.vertexToFace.Store(v2f)
	}
}
