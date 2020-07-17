package model2d

import (
	"math/rand"
	"testing"
)

func TestSegmentIntersections(t *testing.T) {
	seg1 := &Segment{XY(0, 1), XY(1, 0)}
	seg2 := &Segment{XY(0, 0), XY(0.549, 0.549)}
	seg3 := &Segment{XY(0.1, 1), XY(1.1, 0)}
	if seg1.SegmentCollision(seg3) {
		t.Error("unexpected collision")
	}
	if !seg1.SegmentCollision(seg2) {
		t.Error("expected collision")
	}
	if seg2.SegmentCollision(seg3) {
		t.Error("unexpected collision")
	}
}

func TestSegmentRectCollision(t *testing.T) {
	for i := 0; i < 10000; i++ {
		min := NewCoordRandNorm()
		rect := &Rect{
			MinVal: min,
			MaxVal: min.Add(NewCoordRandUniform().AddScalar(0.01)),
		}
		seg := &Segment{NewCoordRandNorm(), NewCoordRandNorm()}

		actual := seg.RectCollision(rect)
		if !actual {
			// Make sure there's no point that is inside the rect.
			for i := 0; i < 100; i++ {
				c := seg[0].Add(seg[1].Sub(seg[0]).Scale(rand.Float64()))
				if rect.Contains(c) {
					t.Fatal("found point inside the rect")
				}
			}
		} else {
			center := rect.MinVal.Mid(rect.MaxVal)
			radius := rect.MaxVal.Dist(center) + 1e-5
			if !seg.CircleCollision(center, radius) {
				t.Fatal("cannot collide with rect inside of circle we don't collide with")
			}
		}
	}
}
