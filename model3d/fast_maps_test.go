// Generated from templates/fast_maps_test.template

package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestCoordMap(t *testing.T) {
	cm := NewCoordMap()

	checkBehavior := func() {
		baseline := map[Coord3D]int{}
		keys := []Coord3D{}
		for i := 0; i < 1000; i++ {
			if len(baseline) > 0 && rand.Intn(3) == 0 {
				// Remove a random element.
				idx := rand.Intn(len(keys))
				k := keys[idx]

				keys[idx] = keys[len(keys)-1]
				keys = keys[:len(keys)-1]
				delete(baseline, k)

				cm.Delete(k)
				if val, ok := cm.Load(NewCoord3DRandNorm()); ok || val != nil {
					t.Fatalf("deletion did not work, got (%v, %v)", val, ok)
				}
			} else {
				// Add a random element.
				k := NewCoord3DRandNorm()
				v := rand.Intn(100)
				if _, ok := baseline[k]; !ok {
					keys = append(keys, k)
				}
				baseline[k] = v
				cm.Store(k, v)
			}
			if len(baseline) != cm.Len() {
				t.Fatalf("should have length %d but got %d", len(baseline), cm.Len())
			}
			count := 0
			cm.Range(func(k Coord3D, v any) bool {
				count++
				if baseline[k] != v {
					t.Fatal("invalid entry")
				}
				return true
			})
			if count != len(baseline) {
				t.Fatalf("expected to enumerate %d values but got %d", len(baseline), count)
			}
			for k, v := range baseline {
				v1, ok := cm.Load(k)
				if !ok || v1 != v {
					t.Fatalf("expected to get (%d, true) but got (%d, %v)", v, v1, ok)
				}
			}
		}
		for k := range baseline {
			cm.Delete(k)
		}
		if cm.Len() != 0 {
			t.Fatal("failed to clear")
		}
	}

	// Testing with fast map.
	checkBehavior()

	coll1, coll2 := findHashCollision()
	cm.Store(coll1, 3)
	cm.Store(coll2, 1)
	if cm.Len() != 2 {
		t.Fatal("incorrect length after collision")
	}
	if v, ok := cm.Load(coll1); !ok || v != 3 {
		t.Error("bad first collision value")
	}
	if v, ok := cm.Load(coll2); !ok || v != 1 {
		t.Error("bad second collision value")
	}
	cm.Delete(coll1)
	if cm.Len() != 1 {
		t.Fatal("incorrect length after delete")
	}
	cm.Delete(coll2)

	// Testing with slow map.
	checkBehavior()
}

func BenchmarkCoordMap(b *testing.B) {
	b.Run("Fast", func(b *testing.B) {
		cm := NewCoordMap()
		c := NewCoord3DRandNorm()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cm.Store(c, 1337)
			cm.Delete(c)
		}
	})
	b.Run("Slow", func(b *testing.B) {
		cm := NewCoordMap()
		c1, c2 := findHashCollision()
		cm.Store(c1, 1337)
		cm.Store(c2, 1337)
		cm.Delete(c1)
		cm.Delete(c2)
		c := NewCoord3DRandNorm()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cm.Store(c, 1337)
			cm.Delete(c)
		}
	})
	b.Run("Baseline", func(b *testing.B) {
		cm := map[Coord3D]any{}
		c := NewCoord3DRandNorm()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cm[c] = 1337
			delete(cm, c)
		}
	})
}

func findHashCollision() (Coord3D, Coord3D) {
	delta := math.Nextafter(0, 1)
	c1 := XY(delta, delta)
	for i := -4.0; i < 5.0; i++ {
		for j := -4.0; j < 5.0; j++ {
			c2 := XY(delta+i*delta, delta+j*delta)
			if c1.fastHash() == c2.fastHash() {
				return c1, c2
			}
		}
	}
	panic("cannot find hash collision")
}
