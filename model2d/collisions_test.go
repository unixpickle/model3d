package model2d

import (
	"math"
	"testing"
)

func TestEmptyColliders(t *testing.T) {
	// Make sure we don't get a crash from this.
	GroupSegments(nil)

	if MeshToCollider(NewMesh()).CircleCollision(Coord{}, 1000) {
		t.Error("unexpected collision")
	}
}

func TestSegmentCollisions(t *testing.T) {
	segment := &Segment{Coord{X: 1, Y: 2}, Coord{X: 2, Y: 1}}
	ray := &Ray{
		Origin:    Coord{X: 0, Y: 1},
		Direction: Coord{X: 1, Y: 0.5},
	}
	n := segment.RayCollisions(ray, func(rc RayCollision) {
		if math.Abs(rc.Scale-1.3333333) > 0.0001 {
			t.Error("unexpected scale")
		}
		if math.Abs(rc.Normal.X-0.7071) > 0.001 || math.Abs(rc.Normal.Y-0.7071) > 0.001 {
			t.Error("unexpected normal")
		}
	})
	if n != 1 {
		t.Errorf("bad number of collisions")
	}
	ray1 := *ray
	ray1.Direction = Coord{X: 1, Y: 1.01}
	if segment.RayCollisions(&ray1, nil) != 0 {
		t.Error("spurrious ray collision")
	}
}

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
