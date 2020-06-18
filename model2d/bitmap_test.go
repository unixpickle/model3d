package model2d

import (
	"math/rand"
	"testing"
)

func TestBitmapMesh(t *testing.T) {
	bmp := testingBitmap()
	mesh := bmp.Mesh()
	collider := MeshToCollider(mesh)

	t.Run("Contains", func(t *testing.T) {
		for y := 0; y < bmp.Height; y++ {
			for x := 0; x < bmp.Width; x++ {
				p := Coord{float64(x) + 0.5, float64(y) + 0.5}
				if ColliderContains(collider, p, 0) != bmp.Get(x, y) {
					t.Errorf("mismatching collider and bitmap values at %d, %d", x, y)
					break
				}
			}
		}
	})

	t.Run("Normals", func(t *testing.T) {
		if _, n := mesh.RepairNormals(1e-5); n != 0 {
			t.Error("normals need repair")
		}
		for y := 0; y < bmp.Height; y++ {
			for x := 0; x < bmp.Width; x++ {
				ray := &Ray{
					Origin:    Coord{float64(x) + 0.5, float64(y) + 0.5},
					Direction: Coord{rand.NormFloat64(), rand.NormFloat64()},
				}
				collision, collides := collider.FirstRayCollision(ray)
				if !collides {
					if bmp.Get(x, y) {
						t.Errorf("bad collision result at %d, %d", x, y)
						return
					}
					continue
				}
				facingOut := collision.Normal.Dot(ray.Direction) > 0
				if facingOut != bmp.Get(x, y) {
					t.Errorf("incorrect normal direction at %d, %d (contained %v)", x, y,
						bmp.Get(x, y))
					return
				}
			}
		}
	})

	t.Run("Ordered", func(t *testing.T) {
		remaining := map[*Segment]bool{}
		for _, seg := range mesh.SegmentsSlice() {
			remaining[seg] = true
		}
		for len(remaining) > 0 {
			var next *Segment
			for s := range remaining {
				next = s
				break
			}
			delete(remaining, next)
			lastVertex := next[0]
			for next != nil {
				if next[0] != lastVertex {
					t.Fatal("invalid loop")
				}
				lastVertex = next[1]
				for _, s1 := range mesh.Find(lastVertex) {
					if s1 == next {
						continue
					}
					next = s1
				}
				if !remaining[next] {
					break
				}
				delete(remaining, next)
			}
		}
	})

	t.Run("Singular", func(t *testing.T) {
		mesh.Iterate(func(s *Segment) {
			for _, p := range s {
				if len(mesh.Find(p)) != 2 {
					t.Fatal("incorrect neighbor count for vertex")
				}
			}
		})
	})
}

func testingBitmap() *Bitmap {
	bmp := NewBitmap(200, 300)
	for i := range bmp.Data {
		if rand.Intn(3) > 0 {
			bmp.Data[i] = true
		}
	}
	return bmp
}
