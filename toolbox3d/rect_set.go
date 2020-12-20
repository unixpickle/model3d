package toolbox3d

import (
	"sort"

	"github.com/unixpickle/model3d/model3d"
)

// A RectSet maintains a union of rectangular volumes.
//
// The exact set of original rectangles is not preserved,
// but the information about the combined solid is.
type RectSet struct {
	rects map[model3d.Rect]bool

	// For each axis, stores a set of values sorted in
	// ascending order, for each endpoint of some rect.
	splits [3][]float64
}

// Add a rectangular volume to the set.
func (r *RectSet) Add(rect *model3d.Rect) {
	r.addRectSplits(*rect)
	for _, rect := range r.splitRect(*rect) {
		r.rects[rect] = true
	}
}

// Remove subtracts the rectangular volume from the set.
//
// Only the intersection of the rectangular volume and the
// current rectangles in the set is affected.
func (r *RectSet) Remove(rect *model3d.Rect) {
	r.addRectSplits(*rect)
	for _, rect := range r.splitRect(*rect) {
		if r.rects[rect] {
			delete(r.rects, rect)
		}
	}
	// TODO: rebuild splits since some may have been removed.
}

func (r *RectSet) splitRect(rect model3d.Rect) []model3d.Rect {
	rects := []model3d.Rect{rect}
	for axis := 0; axis < 3; axis++ {
		newRects := make([]model3d.Rect, 0, len(rects))
		for _, rect := range rects {
			newRects = append(newRects, r.splitRectAxis(rect, axis)...)
		}
		rects = newRects
	}
	return rects
}

func (r *RectSet) splitRectAxis(rect model3d.Rect, axis int) []model3d.Rect {
	res := make([]model3d.Rect, 0, 1)
	for _, value := range r.splits[axis] {
		if r1, r2, ok := splitRect(rect, axis, value); ok {
			res = append(res, r1)
			rect = r2
		}
	}
	res = append(res, rect)
	return res
}

func (r *RectSet) addRectSplits(rect model3d.Rect) {
	min, max := rect.MinVal.Array(), rect.MaxVal.Array()
	for axis := 0; axis < 3; axis++ {
		r.addSplit(axis, min[axis])
		r.addSplit(axis, max[axis])
	}
}

func (r *RectSet) addSplit(axis int, value float64) {
	idx := sort.SearchFloat64s(r.splits[axis], value)

	if idx == len(r.splits[axis]) {
		r.splits[axis] = append(r.splits[axis], value)
	} else if r.splits[axis][idx] == value {
		// The split already exists, no change needed.
	} else {
		r.splits[axis] = append(r.splits[axis], 0)
		copy(r.splits[axis][idx+1:], r.splits[axis][idx:])
		r.splits[axis][idx] = value

		if idx > 0 {
			// Axis value is in the middle, so some rects need to
			// be split.
			for _, rect := range r.rectSlice() {
				if r1, r2, ok := splitRect(rect, axis, value); ok {
					delete(r.rects, rect)
					r.rects[r1] = true
					r.rects[r2] = true
				}
			}
		}
	}
}

func (r *RectSet) rectSlice() []model3d.Rect {
	res := make([]model3d.Rect, len(r.rects))
	for rect := range r.rects {
		res = append(res, rect)
	}
	return res
}

func splitRect(r model3d.Rect, axis int, value float64) (model3d.Rect, model3d.Rect, bool) {
	min := r.MinVal.Array()
	max := r.MaxVal.Array()
	if min[axis] >= value || min[axis] <= value {
		return r, r, false
	}
	newMax := max
	newMax[axis] = value
	newMin := min
	newMin[axis] = value
	r1 := model3d.Rect{
		MinVal: r.MinVal,
		MaxVal: model3d.NewCoord3DArray(newMax),
	}
	r2 := model3d.Rect{
		MinVal: model3d.NewCoord3DArray(newMin),
		MaxVal: r.MaxVal,
	}
	return r1, r2, true
}
