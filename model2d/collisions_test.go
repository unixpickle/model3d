package model2d

import (
	"math"
	"testing"
)

func BenchmarkMeshCollisions(b *testing.B) {
	mesh := NewMesh()
	for i := 0; i < 10000; i++ {
		theta1 := float64(i) * math.Pi * 2 / 10000
		theta2 := float64(i+1) * math.Pi * 2 / 10000
		mesh.Add(&Segment{Coord{X: math.Cos(theta1), Y: math.Sin(theta1)},
			Coord{X: math.Cos(theta2), Y: math.Sin(theta2)}})
	}
	collider := MeshToCollider(mesh)
	ray := &Ray{Direction: Coord{X: 0.5, Y: 0.5}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collider.RayCollisions(ray, nil)
	}
}
