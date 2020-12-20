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
