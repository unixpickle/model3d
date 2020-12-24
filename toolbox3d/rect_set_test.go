package toolbox3d

import (
	"math/rand"
	"testing"

	"github.com/unixpickle/model3d/model3d"
)

func TestRectSetAddRemove(t *testing.T) {
	for i := 0; i < 10; i++ {
		var solid model3d.Solid

		solid = model3d.JoinedSolid{}
		rectSet := NewRectSet()

		bounder := model3d.JoinedSolid{}
		for t := 0; t < 10; t++ {
			nextRect := randomRect()
			bounder = append(bounder, nextRect)
			if rand.Intn(2) == 0 {
				rectSet.Remove(nextRect)
				solid = &model3d.SubtractedSolid{
					Positive: solid,
					Negative: nextRect,
				}
			} else {
				rectSet.Add(nextRect)
				solid = model3d.JoinedSolid{solid, nextRect}
			}
		}

		setSolid := naiveRectSetSolid(rectSet)
		testSolidsEquivalent(t, bounder, setSolid, solid)
	}
}

func TestAddRectSet(t *testing.T) {
	for i := 0; i < 10; i++ {
		rs1 := randomRectSet()
		rs2 := randomRectSet()

		expected := model3d.JoinedSolid{
			naiveRectSetSolid(rs1),
			naiveRectSetSolid(rs2),
		}
		rs1.AddRectSet(rs2)
		testSolidsEquivalent(t, expected, naiveRectSetSolid(rs1), expected)
	}
}

func TestRemoveRectSet(t *testing.T) {
	for i := 0; i < 10; i++ {
		rs1 := randomRectSet()
		rs2 := randomRectSet()

		expected := &model3d.SubtractedSolid{
			Positive: naiveRectSetSolid(rs1),
			Negative: naiveRectSetSolid(rs2),
		}
		rs1.RemoveRectSet(rs2)
		testSolidsEquivalent(t, expected, naiveRectSetSolid(rs1), expected)
	}
}

func TestRectSetSolid(t *testing.T) {
	for i := 0; i < 10; i++ {
		rs1 := randomRectSet()
		actual := rs1.Solid()
		expected := naiveRectSetSolid(rs1)
		testSolidsEquivalent(t, expected, actual, expected)
	}
}

func TestRectSetMesh(t *testing.T) {
	t.Run("Unaligned", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			rs := randomRectSet()
			mesh := rs.Mesh()
			if mesh.NeedsRepair() {
				t.Error("mesh needs repair")
			}
			if n := len(mesh.SingularVertices()); n != 0 {
				t.Errorf("mesh has %d singular vertices", n)
			}
			if _, n := mesh.RepairNormals(1e-5); n != 0 {
				t.Errorf("mesh has %d bad normals", n)
			}
			actual := model3d.NewColliderSolid(model3d.MeshToCollider(mesh))
			expected := rs.Solid()
			testSolidsEquivalent(t, expected, actual, expected)
		}
	})
	t.Run("Aligned", func(t *testing.T) {
		scales := []model3d.Coord3D{
			model3d.XYZ(1, 1, 1),
			model3d.XYZ(1, 10, 1),
			model3d.XYZ(1, 10, 10),
			model3d.XYZ(1, 10, 100),
			model3d.XYZ(1, 1, 100),
			model3d.XYZ(1, 1, 1000),
		}
		for _, scale := range scales {
			rs := alignedNonManifoldRectSet(scale)
			mesh := rs.Mesh()
			if mesh.NeedsRepair() {
				t.Error("mesh needs repair")
			}
			if n := len(mesh.SingularVertices()); n != 0 {
				t.Errorf("mesh has %d singular vertices", n)
			}
			if _, n := mesh.RepairNormals(1e-5); n != 0 {
				t.Errorf("mesh has %d bad normals", n)
			}
		}
	})
}

func randomRectSet() *RectSet {
	rectSet := NewRectSet()
	for len(rectSet.rectSlice()) == 0 {
		for t := 0; t < 10; t++ {
			nextRect := randomRect()
			if rand.Intn(2) == 0 {
				rectSet.Remove(nextRect)
			} else {
				rectSet.Add(nextRect)
			}
		}
	}
	return rectSet
}

func alignedNonManifoldRectSet(scale model3d.Coord3D) *RectSet {
	for {
		rs := NewRectSet()
		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				for z := 0; z < 10; z++ {
					if rand.Intn(2) == 0 {
						rs.Add(&model3d.Rect{
							MinVal: model3d.XYZ(float64(x), float64(y), float64(z)).Mul(scale),
							MaxVal: model3d.XYZ(float64(x+1), float64(y+1), float64(z+1)).Mul(scale),
						})
					}
				}
			}
		}
		m := rs.ExactMesh()
		if m.NeedsRepair() && len(m.SingularVertices()) != 0 {
			return rs
		}
	}
}

func naiveRectSetSolid(r *RectSet) model3d.Solid {
	var res model3d.JoinedSolid
	for _, rect := range r.rectSlice() {
		r := rect
		res = append(res, &r)
	}
	return res
}

func randomRect() *model3d.Rect {
	min := model3d.NewCoord3DRandNorm()
	return &model3d.Rect{
		MinVal: min,
		MaxVal: min.Add(model3d.NewCoord3DRandUniform().Scale(2)),
	}
}

func testSolidsEquivalent(t *testing.T, b model3d.Bounder, actual, expected model3d.Solid) {
	min, max := b.Min(), b.Max()
	for i := 0; i < 100; i++ {
		point := model3d.NewCoord3DRandBounds(min, max)
		a := actual.Contains(point)
		x := expected.Contains(point)
		if a != x {
			t.Fatalf("point %v: expected %v but got %v", point, x, a)
		}
	}
}
