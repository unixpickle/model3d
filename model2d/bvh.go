// Generated from templates/bvh.template

package model2d

import (
	"math"
	"sort"
)

// GeneralBVH represents a (possibly unbalanced)
// axis-aligned bounding box hierarchy.
//
// A GeneralBVH can store arbitrary Bounders.
// For a mesh-specific version, see BVH.
type GeneralBVH struct {
	// Leaf, if non-nil, is the final bounder.
	Leaf Bounder

	// Branch, if Leaf is nil, points to two children.
	Branch []*GeneralBVH
}

// NewGeneralBVHAreaDensity creates a GeneralBVH by
// minimizing the product of each bounding box's perimeter with
// the number of objects contained in the bounding box at
// each branch.
//
// This is good for efficient ray collision detection.
func NewGeneralBVHAreaDensity(objects []Bounder) *GeneralBVH {
	return newGeneralBVH(sortBounders(objects), make([]float64, len(objects)),
		areaDensityBVHSplit)
}

func newGeneralBVH(sortedBounders [2][]*flaggedBounder, cache []float64,
	splitter func([]*flaggedBounder, []float64) (int, float64)) *GeneralBVH {
	numObjs := len(sortedBounders[0])
	if numObjs == 0 {
		panic("empty sorted objects")
	} else if numObjs == 1 {
		return &GeneralBVH{Leaf: sortedBounders[0][0].B}
	} else if numObjs == 2 {
		return &GeneralBVH{Branch: []*GeneralBVH{
			{Leaf: sortedBounders[0][0].B},
			{Leaf: sortedBounders[0][1].B},
		}}
	}

	xIndex, xScore := splitter(sortedBounders[0], cache)
	yIndex, yScore := splitter(sortedBounders[1], cache)

	var split [2][2][]*flaggedBounder
	if xScore < yScore {
		split = splitBounders(sortedBounders, 0, xIndex)
	} else {
		split = splitBounders(sortedBounders, 1, yIndex)
	}
	return &GeneralBVH{
		Branch: []*GeneralBVH{
			newGeneralBVH(split[0], cache, splitter),
			newGeneralBVH(split[1], cache, splitter),
		},
	}
}

// BVH represents a (possibly unbalanced) axis-aligned
// bounding box hierarchy of segments.
//
// A BVH can be used to accelerate collision detection.
// See BVHToCollider() for more details.
//
// A BVH node is either a leaf (a single segment), or a branch
// with two or more children.
//
// For a more generic BVH that supports any object rather
// than just segments, see GeneralBVH.
type BVH struct {
	// Leaf, if non-nil, is the sole object in this node.
	Leaf *Segment

	// Branch, if Leaf is nil, points to two children.
	Branch []*BVH
}

// NewBVHAreaDensity is like NewGeneralBVHAreaDensity but
// for segments.
func NewBVHAreaDensity(objs []*Segment) *BVH {
	return generalBVHToBVH(NewGeneralBVHAreaDensity(facesToBounders(objs)))
}

func generalBVHToBVH(g *GeneralBVH) *BVH {
	if g.Leaf != nil {
		return &BVH{Leaf: g.Leaf.(*Segment)}
	}
	res := &BVH{Branch: make([]*BVH, len(g.Branch))}
	for i, g1 := range g.Branch {
		res.Branch[i] = generalBVHToBVH(g1)
	}
	return res
}

// areaDensityBVHSplit chooses a split index that
// minimizes a goodness score, and returns the index and
// score.
//
// The score of a bbox is equal to the surface area times
// the number of segments.
//
// The cache must contain at least len(faces) entries.
func areaDensityBVHSplit(faces []*flaggedBounder, cache []float64) (int, float64) {
	// Fill the cache with scores going in the other
	// direction.
	min, max := faces[len(faces)-1].Min, faces[len(faces)-1].Max
	for i := len(faces) - 2; i >= 0; i-- {
		cache[i] = boundsArea(min, max) * float64(len(faces)-i-1)
		t := faces[i]
		min = min.Min(t.Min)
		max = max.Max(t.Max)
	}

	var bestScore float64
	var bestIndex int

	min, max = faces[0].Min, faces[0].Max
	for i := 1; i < len(faces)-1; i++ {
		t := faces[i]
		min = min.Min(t.Min)
		max = max.Max(t.Max)
		score := boundsArea(min, max)*float64(i) + cache[i]
		if score < bestScore || i == 1 {
			bestScore = score
			bestIndex = i + 1
		}
	}

	return bestIndex, bestScore
}

// GroupSegments is like GroupBounders, but for segments
// in particular.
//
// This can be used to prepare models for being turned
// into a collider efficiently, or for storing meshes in
// an order well-suited for file compression.
//
// The resulting hierarchy can be passed directly to
// GroupedSegmentsToCollider().
func GroupSegments(faces []*Segment) {
	bs := facesToBounders(faces)
	groupBounders(sortBounders(bs), bs)
	for i, b := range bs {
		faces[i] = b.(*Segment)
	}
}

// GroupBounders sorts a slice of objects into a balanced
// bounding box hierarchy.
//
// The sorted slice can be recursively cut in half, and
// each half will be spatially separated as well as
// possible along some axis.
// To cut a slice in half, divide the length by two, round
// down, and use the result as the start index for the
// second half.
func GroupBounders(objects []Bounder) {
	groupBounders(sortBounders(objects), objects)
}

func groupBounders(sortedBounders [2][]*flaggedBounder, output []Bounder) {
	numObjs := len(sortedBounders[0])
	if numObjs == 2 {
		// The area-based splitting criterion doesn't
		// distinguish between axes, now.
		output[0] = sortedBounders[0][0].B
		output[1] = sortedBounders[0][1].B
		return
	} else if numObjs == 1 {
		output[0] = sortedBounders[0][0].B
		return
	} else if numObjs == 0 {
		return
	}

	midIdx := numObjs / 2
	axis := bestSplitAxis(sortedBounders)

	separated := splitBounders(sortedBounders, axis, midIdx)
	groupBounders(separated[0], output[:midIdx])
	groupBounders(separated[1], output[midIdx:])
}

func splitBounders(sortedBounders [2][]*flaggedBounder, axis, midIdx int) [2][2][]*flaggedBounder {
	for i, b := range sortedBounders[axis] {
		b.Flag = i < midIdx
	}

	separated := [2][]*flaggedBounder{}
	separated[axis] = sortedBounders[axis]

	numObjs := len(sortedBounders[0])
	for newAxis := 0; newAxis < 2; newAxis++ {
		if newAxis == axis {
			continue
		}
		sep := make([]*flaggedBounder, numObjs)
		idx0 := 0
		idx1 := midIdx
		for _, b := range sortedBounders[newAxis] {
			if b.Flag {
				sep[idx0] = b
				idx0++
			} else {
				sep[idx1] = b
				idx1++
			}
		}
		separated[newAxis] = sep
	}

	return [2][2][]*flaggedBounder{
		{
			separated[0][:midIdx],
			separated[1][:midIdx],
		},
		{
			separated[0][midIdx:],
			separated[1][midIdx:],
		},
	}
}

func bestSplitAxis(sortedBounders [2][]*flaggedBounder) int {
	midIdx := len(sortedBounders[0]) / 2

	areaForAxis := func(axis int) float64 {
		return multipleBoundsArea(sortedBounders[axis][:midIdx]) +
			multipleBoundsArea(sortedBounders[axis][midIdx:])
	}

	axis := 0
	minArea := areaForAxis(0)
	for i := 1; i < 2; i++ {
		if a := areaForAxis(i); a < minArea {
			minArea = a
			axis = i
		}
	}

	return axis
}

func sortBounders(bs []Bounder) [2][]*flaggedBounder {
	// Allocate all of the flaggedBounders at once all
	// next to each other in memory.
	flagged := make([]flaggedBounder, len(bs))
	for i, b := range bs {
		min, max := b.Min(), b.Max()
		flagged[i] = flaggedBounder{
			B:   b,
			Min: min,
			Max: max,
			Mid: min.Mid(max),
		}
	}

	var result [2][]*flaggedBounder
	for axis := range result {
		bsCopy := make([]*flaggedBounder, len(flagged))
		for i := range flagged {
			bsCopy[i] = &flagged[i]
		}
		if axis == 0 {
			sort.Slice(bsCopy, func(i, j int) bool {
				return bsCopy[i].Mid.X < bsCopy[j].Mid.X
			})
		} else if axis == 1 {
			sort.Slice(bsCopy, func(i, j int) bool {
				return bsCopy[i].Mid.Y < bsCopy[j].Mid.Y
			})

		}
		result[axis] = bsCopy
	}
	return result
}

func multipleBoundsArea(bs []*flaggedBounder) float64 {
	min, max := bs[0].Min, bs[0].Max
	for i := 1; i < len(bs); i++ {
		b := bs[i]

		// This is very expanded (unwrapped) vs. using
		// Min() and Max(), but it is faster and this is
		// surprisingly a large bottleneck.
		min1 := b.Min
		if min1.X < min.X {
			min.X = min1.X
		}
		if min1.Y < min.Y {
			min.Y = min1.Y
		}
		max1 := b.Max
		if max1.X > max.X {
			max.X = max1.X
		}
		if max1.Y > max.Y {
			max.Y = max1.Y
		}
	}
	return boundsArea(min, max)
}

func boundsArea(min, max Coord) float64 {
	diff := max.Sub(min)
	return 2 * (diff.X + diff.Y)
}

type flaggedBounder struct {
	B    Bounder
	Min  Coord
	Max  Coord
	Mid  Coord
	Flag bool
}

func circleTouchesBounds(center Coord, r float64, min, max Coord) bool {
	return pointToBoundsDistSquared(center, min, max) <= r*r
}

func pointToBoundsDistSquared(center Coord, min, max Coord) float64 {
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 2; axis++ {
		min := min.Array()[axis]
		max := max.Array()[axis]
		value := center.Array()[axis]
		if value < min {
			distSquared += (min - value) * (min - value)
		} else if value > max {
			distSquared += (max - value) * (max - value)
		}
	}
	return distSquared
}

func rayCollisionWithBounds(r *Ray, min, max Coord) (minFrac, maxFrac float64) {
	minFrac = math.Inf(-1)
	maxFrac = math.Inf(1)
	for axis := 0; axis < 2; axis++ {
		origin := r.Origin.Array()[axis]
		rate := r.Direction.Array()[axis]
		if rate == 0 {
			if origin < min.Array()[axis] || origin > max.Array()[axis] {
				return 0, -1
			}
			continue
		}
		t1 := (min.Array()[axis] - origin) / rate
		t2 := (max.Array()[axis] - origin) / rate
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t2 < 0 {
			// Short-circuit optimization.
			return 0, -1
		}
		if t1 > minFrac {
			minFrac = t1
		}
		if t2 < maxFrac {
			maxFrac = t2
		}
	}
	return
}

func facesToBounders(faces []*Segment) []Bounder {
	bs := make([]Bounder, len(faces))
	for i, t := range faces {
		bs[i] = t
	}
	return bs
}
