package model2d

import (
	"math"
	"testing"
)

func TestMeshArea(t *testing.T) {
	t.Run("Circle", func(t *testing.T) {
		circle := NewMeshPolar(func(t float64) float64 {
			return 1.0
		}, 2000)
		expected := math.Pi
		actual := circle.Area()
		if math.Abs(actual-expected) > 1e-4 {
			t.Errorf("expected area %f but got %f", expected, actual)
		}
	})
	t.Run("Concentric", func(t *testing.T) {
		mesh := NewMeshPolar(func(t float64) float64 {
			return 1.0
		}, 2000)
		mesh.AddMesh(mesh.MapCoords(func(c Coord) Coord {
			return XY(c.X*0.5, -c.Y*0.5)
		}))
		expected := 0.75 * math.Pi
		actual := mesh.Area()
		if math.Abs(actual-expected) > 1e-4 {
			t.Errorf("expected area %f but got %f", expected, actual)
		}
	})
}
