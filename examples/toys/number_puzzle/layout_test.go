package main

import (
	"reflect"
	"testing"
)

func TestTranslation(t *testing.T) {
	shape := NewDigitContinuous([]Location{
		{1, 2},
		{2, 2},
		{2, 3},
		{3, 3},
		{3, 2},
		{3, 1},
	})
	shape.Translate(Location{2, 1})
	shape1 := NewDigitContinuous([]Location{
		{3, 3},
		{4, 3},
		{4, 4},
		{5, 4},
		{5, 3},
		{5, 2},
	})

	if !reflect.DeepEqual(shape, shape1) {
		t.Errorf("should be equal: %v %v", shape, shape1)
	}
}

func TestRotation(t *testing.T) {
	shape := NewDigitContinuous([]Location{
		{1, 2},
		{2, 2},
		{2, 3},
		{3, 3},
		{3, 2},
		{3, 1},
	})
	r1 := shape.Copy()
	r2 := shape.Copy()

	r1.Translate(Location{10, 3})
	r1.Rotate()
	r2.Rotate()
	if !reflect.DeepEqual(r1, r2) {
		t.Errorf("should be equal: %v %v", r1, r2)
	}

	for i := 0; i < 4; i++ {
		if i > 0 {
			if reflect.DeepEqual(r1, r2) {
				t.Errorf("should not be equal after %d rotations", i)
			}
		}
		r2.Rotate()
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Errorf("should be equal: %v %v", r1, r2)
	}
}
