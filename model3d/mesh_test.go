package model3d

import (
	"math/rand"
	"testing"
)

func TestNewMeshRect(t *testing.T) {
	mesh := NewMeshRect(XYZ(-0.3, -0.4, -0.2), XYZ(0.4, 0.35, 0.19))
	MustValidateMesh(t, mesh, true)
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
