package model2d

import (
	"image/color"
	"testing"
)

func TestMeshHierarchy(t *testing.T) {
	// Create a testing mesh with a complex hierarchy.
	bitmap := MustReadBitmap("test_data/test_bitmap.png", func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		return r == 0 && g == 0 && b == 0
	})
	mesh := bitmap.Mesh().SmoothSq(30)
	MustValidateMesh(t, mesh)

	hierarchy := MeshToHierarchy(mesh)

	// Specific tests for this hierarchy.
	if len(hierarchy) != 3 {
		t.Errorf("expected 3 separate roots but found %d", len(hierarchy))
	}
	if depth := measureHierarchyDepth(hierarchy); depth != 5 {
		t.Errorf("expected 5 nested meshes but found %d", depth)
	}

	// Make sure all vertices are preserved.
	flatCount := len(mesh.VertexSlice())
	hierarchyCount := countHierarchyVertices(hierarchy)
	if flatCount != hierarchyCount {
		t.Errorf("expected %v vertices but got %v", flatCount, hierarchyCount)
	}

	// Make sure child containment is preserved.
	for _, h := range hierarchy {
		validateHierarchyContainment(t, h)
	}
	rawSolid := NewColliderSolid(MeshToCollider(mesh))
	for i := 0; i < 10000; i++ {
		c := NewCoordRandBounds(rawSolid.Min(), rawSolid.Max())
		contained := rawSolid.Contains(c)
		hContained := false
		for _, h := range hierarchy {
			if h.Contains(c) {
				hContained = true
			}
		}
		if contained != hContained {
			t.Errorf("point %v should have contained=%v but have %v", c, contained, hContained)
		}
	}
}

func countHierarchyVertices(hierarchies []*MeshHierarchy) int {
	var res int
	for _, child := range hierarchies {
		res += len(child.Mesh.VertexSlice())
		res += countHierarchyVertices(child.Children)
	}
	return res
}

func measureHierarchyDepth(hierarchies []*MeshHierarchy) int {
	var result int
	for _, h := range hierarchies {
		depth := measureHierarchyDepth(h.Children) + 1
		if depth > result {
			result = depth
		}
	}
	return result
}

func validateHierarchyContainment(t *testing.T, h *MeshHierarchy) {
	for _, c := range h.Children {
		for _, v := range c.Mesh.VertexSlice() {
			if !h.MeshSolid.Contains(v) {
				t.Fatal("child not contained in parent")
			}
		}
		validateHierarchyContainment(t, c)
	}
}

func BenchmarkMeshHierarchy(b *testing.B) {
	// Create a testing mesh with a complex hierarchy.
	bitmap := MustReadBitmap("test_data/test_bitmap.png", func(c color.Color) bool {
		r, g, b, _ := c.RGBA()
		return r == 0 && g == 0 && b == 0
	})
	mesh := bitmap.Mesh().SmoothSq(30)
	MustValidateMesh(b, mesh)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MeshToHierarchy(mesh)
	}
}
