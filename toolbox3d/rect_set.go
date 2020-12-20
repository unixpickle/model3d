package toolbox3d

import (
	"sort"

	"github.com/unixpickle/model3d/model3d"
)

// A RectSet maintains the set of all points contained in
// a union of rectangular volumes.
//
// The exact list of original rectangles is not preserved,
// but the information about the combined solid is.
//
// A RectSet may use up to O(N^3) memory, where N is the
// number of contained rectangular volumes.
// In particular, usage is proportional to X*Y*Z, where X,
// Y and Z are the number of unique x, y, and z
// coordinates.
type RectSet struct {
	rects map[model3d.Rect]bool

	// For each axis, stores a set of values sorted in
	// ascending order, for each endpoint of some rect.
	splits [3][]float64
}

// NewRectSet creates an empty RectSet.
func NewRectSet() *RectSet {
	return &RectSet{
		rects: map[model3d.Rect]bool{},
	}
}

// Add adds a rectangular volume to the set.
func (r *RectSet) Add(rect *model3d.Rect) {
	r.addRectSplits(*rect)
	for _, rect := range r.splitRect(*rect) {
		r.rects[rect] = true
	}
}

// AddRectSet adds another RectSet's volume to the set.
func (r *RectSet) AddRectSet(r1 *RectSet) {
	for axis, otherSplits := range r1.splits {
		for _, s := range otherSplits {
			r.addSplit(axis, s)
		}
	}
	for rect := range r1.rects {
		for _, rect := range r.splitRect(rect) {
			r.rects[rect] = true
		}
	}
}

// Remove subtracts a rectangular volume from the set.
func (r *RectSet) Remove(rect *model3d.Rect) {
	r.addRectSplits(*rect)
	for _, rect := range r.splitRect(*rect) {
		if r.rects[rect] {
			delete(r.rects, rect)
		}
	}
	r.rebuildSplits()
}

// RemoveRectSet subtracts another RectSet's volume from
// the set.
func (r *RectSet) RemoveRectSet(r1 *RectSet) {
	for axis, otherSplits := range r1.splits {
		for _, s := range otherSplits {
			r.addSplit(axis, s)
		}
	}

	for rect := range r1.rects {
		for _, rect := range r.splitRect(rect) {
			if r.rects[rect] {
				delete(r.rects, rect)
			}
		}
	}

	r.rebuildSplits()
}

func (r *RectSet) rebuildSplits() {
	var axisValues [3]map[float64]bool
	for i := range axisValues {
		axisValues[i] = map[float64]bool{}
	}

	for rect := range r.rects {
		for _, c := range []model3d.Coord3D{rect.MinVal, rect.MaxVal} {
			for axis, value := range c.Array() {
				axisValues[axis][value] = true
			}
		}
	}

	for axis, values := range axisValues {
		valuesSlice := make([]float64, 0, len(values))
		for x := range values {
			valuesSlice = append(valuesSlice, x)
		}
		sort.Float64s(valuesSlice)
		r.splits[axis] = valuesSlice
	}
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
	if min[axis] >= value || max[axis] <= value {
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
