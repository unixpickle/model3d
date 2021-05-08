package model3d

import (
	"math"
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d/model2d"
)

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

func TestNewMeshTorus(t *testing.T) {
	mesh := NewMeshTorus(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19), 0.3, 0.7, 20, 20)
	MustValidateMesh(t, mesh, true)
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
			mesh.Find(vertices[j], vertices[j+1])
		}
	})
}

func BenchmarkVertexToFace(b *testing.B) {
	mesh := NewMesh()
	for i := 0; i < 20000; i++ {
		mesh.Add(randomTriangle())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.getVertexToFace()
		var v2f map[Coord3D][]*Triangle
		mesh.vertexToFace.Store(v2f)
	}
}
