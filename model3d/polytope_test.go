// Generated from templates/polytope_test.template

package model3d

import (
	"math"
	"math/rand"
	"testing"
)

func TestPolytopeMesh(t *testing.T) {
	t.Run("Rect", func(t *testing.T) {
		testPolytopeMesh(t, ConvexPolytope{
			&LinearConstraint{
				Normal: X(1),
				Max:    0.3,
			},
			&LinearConstraint{
				Normal: X(-1),
				Max:    0.29,
			},
			&LinearConstraint{
				Normal: Y(1),
				Max:    0.1,
			},
			&LinearConstraint{
				Normal: Y(-1),
				Max:    0.12,
			},
			&LinearConstraint{
				Normal: Z(1),
				Max:    0.5,
			},
			&LinearConstraint{
				Normal: Z(-1),
				Max:    0.5,
			},
		})
	})

	t.Run("RectUnnormalized", func(t *testing.T) {
		testPolytopeMesh(t, ConvexPolytope{
			&LinearConstraint{
				Normal: X(1e90),
				Max:    0.3 * 1e90,
			},
			&LinearConstraint{
				Normal: X(-1),
				Max:    0.29,
			},
			&LinearConstraint{
				Normal: Y(1),
				Max:    0.1,
			},
			&LinearConstraint{
				Normal: Y(-1),
				Max:    0.12,
			},
			&LinearConstraint{
				Normal: Z(1),
				Max:    0.5,
			},
			&LinearConstraint{
				Normal: Z(-1e50),
				Max:    0.5 * 1e50,
			},
		})
	})
}

func testPolytopeMesh(t *testing.T, c ConvexPolytope) {
	mesh := c.Mesh()

	MustValidateMesh(t, mesh, true)

	solid := NewColliderSolid(MeshToCollider(mesh))
	sdf := MeshToSDF(mesh)

	min, max := mesh.Min(), mesh.Max()
	sampleMin := min.Sub(max.Sub(min).Scale(0.1))
	sampleMax := max.Add(max.Sub(min).Scale(0.1))
	for i := 0; i < 1000; i++ {
		coord := NewCoord3DRandBounds(sampleMin, sampleMax)
		if math.Abs(sdf.SDF(coord)) < 1e-5 {
			// Avoid checks close to the boundary,
			// where rounding errors might cause a
			// discrepancy.
			i--
			continue
		}
		if c.Contains(coord) != solid.Contains(coord) {
			t.Errorf("mismatch containment for %v (%v vs %v)", coord,
				c.Contains(coord), solid.Contains(coord))
		}
	}
}

func BenchmarkPolytope(b *testing.B) {
	rand.Seed(1337)

	polytope := ConvexPolytope{}
	for i := 0; i < 100; i++ {
		normal := NewCoord3DRandUnit()
		polytope = append(polytope, &LinearConstraint{
			Normal: normal,
			Max:    rand.Float64() + 0.1,
		})
	}

	b.Run("Mesh", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			polytope.Mesh()
		}
	})
	b.Run("Solid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			polytope.Solid()
		}
	})
}
