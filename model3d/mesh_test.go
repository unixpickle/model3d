package model3d

import (
	"math/rand"
	"testing"
)

func TestNewMeshRect(t *testing.T) {
	mesh := NewMeshRect(Coord3D{X: -0.3, Y: -0.4, Z: -0.2}, Coord3D{X: 0.4, Y: 0.35, Z: 0.19})
	if mesh.NeedsRepair() {
		t.Error("mesh needs repair")
	}
	if len(mesh.TriangleSlice()) != 12 {
		t.Errorf("expected exactly 12 triangles, but got %d", len(mesh.TriangleSlice()))
	}
	if _, n := mesh.RepairNormals(1e-8); n != 0 {
		t.Errorf("found %d bad normals", n)
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
	mesh.getVertexToTriangle()

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
