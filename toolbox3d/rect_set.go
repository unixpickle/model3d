package toolbox3d

import (
	"math"
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

// Min gets the minimum of the bounding box containing the
// set.
func (r *RectSet) Min() model3d.Coord3D {
	if len(r.rects) == 0 {
		return model3d.Coord3D{}
	}
	return model3d.XYZ(r.splits[0][0], r.splits[1][0], r.splits[2][0])
}

// Min gets the maximum of the bounding box containing the
// set.
func (r *RectSet) Max() model3d.Coord3D {
	if len(r.rects) == 0 {
		return model3d.Coord3D{}
	}
	var res [3]float64
	for i, splits := range r.splits {
		res[i] = splits[len(splits)-1]
	}
	return model3d.NewCoord3DArray(res)
}

// Solid creates a 3D solid from the set.
func (r *RectSet) Solid() model3d.Solid {
	return newRectSetSolid(r)
}

// Mesh creates a manifold 3D mesh from the set of rects.
func (r *RectSet) Mesh() *model3d.Mesh {
	m := r.ExactMesh()

	epsilon := r.Max().Dist(r.Min()) * 1e-8
	for _, splits := range r.splits {
		for i := 1; i < len(splits); i++ {
			epsilon = math.Min(epsilon, 0.1*(splits[i]-splits[i-1]))
		}
	}

	fixSingularEdges(m, epsilon)
	fixSingularVertices(m, epsilon*0.01)

	return m
}

// ExactMesh creates a 3D mesh from the set of rects.
//
// The returned mesh is not guaranteed to be manifold.
// For example, it is possible to create two rects that
// touch at a single vertex or edge, in which case there
// will be a singularity in the resulting mesh.
//
// To create a manifold mesh, use Mesh().
func (r *RectSet) ExactMesh() *model3d.Mesh {
	// Map (min, max) pairs to ordered quads to track
	// which faces are on the inside of the solid.
	uniqueQuads := map[[2]model3d.Coord3D][4]model3d.Coord3D{}

	for rect := range r.rects {
		min := rect.MinVal
		max := rect.MaxVal

		point := func(x, y, z int) model3d.Coord3D {
			res := min
			if x == 1 {
				res.X = max.X
			}
			if y == 1 {
				res.Y = max.Y
			}
			if z == 1 {
				res.Z = max.Z
			}
			return res
		}

		quads := [6][4]model3d.Coord3D{
			// Front and back faces.
			{min, point(1, 0, 0), point(1, 0, 1), point(0, 0, 1)},
			{max, point(1, 1, 0), point(0, 1, 0), point(0, 1, 1)},
			// Left and right faces.
			{min, point(0, 0, 1), point(0, 1, 1), point(0, 1, 0)},
			{max, point(1, 0, 1), point(1, 0, 0), point(1, 1, 0)},
			// Top and bottom faces.
			{min, point(0, 1, 0), point(1, 1, 0), point(1, 0, 0)},
			{max, point(0, 1, 1), point(0, 0, 1), point(1, 0, 1)},
		}

		for _, q := range quads {
			key := quadMinMax(q[0], q[1], q[2], q[3])
			if _, ok := uniqueQuads[key]; ok {
				delete(uniqueQuads, key)
			} else {
				uniqueQuads[key] = q
			}
		}
	}

	res := model3d.NewMesh()
	for _, q := range uniqueQuads {
		res.AddQuad(q[0], q[1], q[2], q[3])
	}
	return res
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

type rectSetSolid struct {
	axis   int
	cutoff float64

	singleRect *model3d.Rect

	below *rectSetSolid
	above *rectSetSolid

	min model3d.Coord3D
	max model3d.Coord3D
}

func newRectSetSolid(rs *RectSet) *rectSetSolid {
	if len(rs.rects) == 0 {
		return &rectSetSolid{
			singleRect: &model3d.Rect{},
		}
	} else if len(rs.rects) == 1 {
		var rect model3d.Rect
		for r := range rs.rects {
			rect = r
		}
		return &rectSetSolid{
			singleRect: &rect,
		}
	}

	rs1, rs2, axis, cutoff := splitRectSet(rs)
	return &rectSetSolid{
		axis:   axis,
		cutoff: cutoff,
		below:  newRectSetSolid(rs1),
		above:  newRectSetSolid(rs2),
		min:    rs.Min(),
		max:    rs.Max(),
	}
}

func (r *rectSetSolid) Min() model3d.Coord3D {
	if r.singleRect != nil {
		return r.singleRect.MinVal
	}
	return r.min
}

func (r *rectSetSolid) Max() model3d.Coord3D {
	if r.singleRect != nil {
		return r.singleRect.MaxVal
	}
	return r.max
}

func (r *rectSetSolid) Contains(c model3d.Coord3D) bool {
	if r.singleRect != nil {
		return r.singleRect.Contains(c)
	}
	if !model3d.InBounds(r, c) {
		return false
	}
	arr := c.Array()
	if arr[r.axis] < r.cutoff {
		return r.below.Contains(c)
	} else if arr[r.axis] > r.cutoff {
		return r.above.Contains(c)
	} else {
		return r.below.Contains(c) || r.above.Contains(c)
	}
}

func splitRectSet(rs *RectSet) (*RectSet, *RectSet, int, float64) {
	if len(rs.rects) < 2 {
		panic("cannot split singleton RectSet")
	}
	splitAxis := 0
	splitLen := 0
	for i, splits := range rs.splits {
		if len(splits) >= splitLen {
			splitLen = len(splits)
			splitAxis = i
		}
	}

	cutoff := rs.splits[splitAxis][splitLen/2]
	rs1 := NewRectSet()
	rs2 := NewRectSet()
	for rect := range rs.rects {
		if rect.MinVal.Array()[splitAxis] < cutoff {
			rs1.rects[rect] = true
		} else {
			rs2.rects[rect] = true
		}
	}
	rs1.rebuildSplits()
	rs2.rebuildSplits()
	return rs1, rs2, splitAxis, cutoff
}

func quadMinMax(p1, p2, p3, p4 model3d.Coord3D) [2]model3d.Coord3D {
	min := p1.Min(p2.Min(p3.Min(p4)))
	max := p1.Max(p2.Max(p3.Max(p4)))
	return [2]model3d.Coord3D{min, max}
}

// fixSingularEdges fixes edges of two touching diagonal
// edge boxes, since these edges belong to four faces at
// once (which is not allowed).
// The fix is done by splitting the edge apart and pulling
// the two middle vertices apart, producing singular
// points but no singular edges. Singular edges really
// ought not to be touching, since there is only a
// singularity because the touching vertices are not in
// the solid.
func fixSingularEdges(m *model3d.Mesh, epsilon float64) {
	changed := true
	for changed {
		changed = false
		sideToTriangle := map[model3d.Segment][]*model3d.Triangle{}
		m.Iterate(func(t *model3d.Triangle) {
			for _, seg := range t.Segments() {
				sideToTriangle[seg] = append(sideToTriangle[seg], t)
			}
		})
		for seg, triangles := range sideToTriangle {
			if len(triangles) == 2 {
				continue
			} else if len(triangles) == 4 {
				fixSingularEdge(m, seg, triangles, epsilon)
				changed = true
			} else {
				panic("unexpected edge situation")
			}
		}
	}
}

func fixSingularEdge(m *model3d.Mesh, seg model3d.Segment, tris []*model3d.Triangle, epsilon float64) {
	for _, t := range tris {
		if !m.Contains(t) {
			return
		}
	}
	t1 := tris[0]
	var minDot float64
	var t2 *model3d.Triangle
	for _, t := range tris[1:] {
		dir := seg.Other(t).Sub(seg.Other(t1))
		dot := dir.Dot(t1.Normal())
		if dot < minDot {
			minDot = dot
			t2 = t
		}
	}

	var t3, t4 *model3d.Triangle
	for _, t := range tris[1:] {
		if t != t2 {
			if t3 == nil {
				t3 = t
			} else {
				t4 = t
			}
		}
	}

	fixSingularEdgePair(m, seg, t1, t2, epsilon)
	fixSingularEdgePair(m, seg, t3, t4, epsilon)
}

func fixSingularEdgePair(m *model3d.Mesh, seg model3d.Segment, t1, t2 *model3d.Triangle, epsilon float64) {
	p1 := seg.Other(t1)
	p2 := seg.Other(t2)

	// Move the segment's midpoint away from the singular
	// edge to make the edges not touch.
	target := p1.Mid(p2)
	source := seg.Mid()
	direction := target.Sub(seg.Mid()).Normalize().Scale(epsilon)
	mp := source.Add(direction)

	fixSingularEdgeTriangle(m, seg, mp, t1)
	fixSingularEdgeTriangle(m, seg, mp, t2)
}

func fixSingularEdgeTriangle(m *model3d.Mesh, seg model3d.Segment, mid model3d.Coord3D, t *model3d.Triangle) {
	m.Remove(t)
	other := seg.Other(t)
	t1 := &model3d.Triangle{other, seg[0], seg.Mid()}
	t2 := &model3d.Triangle{other, seg[1], seg.Mid()}
	if t1.Normal().Dot(t.Normal()) < 0 {
		t1[0], t1[1] = t1[1], t1[0]
	}
	t1[2] = mid
	if t2.Normal().Dot(t.Normal()) < 0 {
		t2[0], t2[1] = t2[1], t2[0]
	}
	t2[2] = mid
	m.Add(t1)
	m.Add(t2)
}

// fixSingularVertices fixes singular vertices by
// duplicating them and then moving the duplicates
// slightly away from each other.
func fixSingularVertices(m *model3d.Mesh, epsilon float64) {
	for _, v := range m.SingularVertices() {
		for _, family := range singularVertexFamilies(m, v) {
			// Move the vertex closer to the mean of all points
			// in the triangle family. This is not guaranteed to
			// work in general, but seems effective in this case.
			mean := model3d.Coord3D{}
			count := 0.0
			for _, t := range family {
				for _, p := range t {
					count++
					mean = mean.Add(p)
				}
			}
			mean = mean.Scale(1 / count)
			offset := mean.Sub(v).Normalize().Scale(epsilon)
			v1 := v.Add(offset)
			for _, t := range family {
				m.Remove(t)
				for i, p := range t {
					if p == v {
						t[i] = v1
					}
				}
				m.Add(t)
			}
		}
	}
}

func singularVertexFamilies(m *model3d.Mesh, v model3d.Coord3D) [][]*model3d.Triangle {
	var families [][]*model3d.Triangle
	tris := m.Find(v)
	for len(tris) > 0 {
		var family []*model3d.Triangle
		family, tris = singularVertexNextFamily(m, tris)
		families = append(families, family)
	}
	return families
}

func singularVertexNextFamily(m *model3d.Mesh, tris []*model3d.Triangle) (family, leftover []*model3d.Triangle) {
	// See mesh.SingularVertices() for an explanation of
	// this algorithm.

	queue := make([]int, len(tris))
	queue[0] = 1
	changed := true
	numVisited := 1
	for changed {
		changed = false
		for i, status := range queue {
			if status != 1 {
				continue
			}
			t := tris[i]
			for j, t1 := range tris {
				if queue[j] == 0 && t.SharesEdge(t1) {
					queue[j] = 1
					numVisited++
					changed = true
				}
			}
			queue[i] = 2
		}
	}
	if numVisited == len(tris) {
		return tris, nil
	} else {
		for i, status := range queue {
			if status == 0 {
				leftover = append(leftover, tris[i])
			} else {
				family = append(family, tris[i])
			}
		}
		return
	}
}
