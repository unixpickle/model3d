// Generated from templates/bvh.template

package model2d

import (
	"math"
	"sort"
)

// BVH represents a (possibly unbalanced) axis-aligned
// bounding box hierarchy.
//
// A BVH can be used to accelerate collision detection.
// See BVHToCollider() for more details.
//
// A BVH node is either a leaf (a single Bounder), or a
// branch with two or more children.
type BVH[B Bounder] struct {
	// Leaf, if non-nil, is the final bounder.
	Leaf B

	// Branch, if Leaf is nil, points to two children.
	Branch []*BVH[B]
}

// NewBVHAreaDensity creates a BVH by minimizing
// the product of each bounding box's perimeter with
// the number of objects contained in the bounding box at
// each branch.
//
// This is good for efficient ray collision detection.
func NewBVHAreaDensity[B Bounder](objects []B) *BVH[B] {
	return newBVH(sortBounders(objects), make([]float64, len(objects)),
		areaDensityBVHSplit[B])
}

func newBVH[B Bounder](sortedBounders [2][]*flaggedBounder[B], cache []float64,
	splitter func([]*flaggedBounder[B], []float64) (int, float64)) *BVH[B] {
	numObjs := len(sortedBounders[0])
	if numObjs == 0 {
		panic("empty sorted objects")
	} else if numObjs == 1 {
		return &BVH[B]{Leaf: sortedBounders[0][0].B}
	} else if numObjs == 2 {
		return &BVH[B]{Branch: []*BVH[B]{
			{Leaf: sortedBounders[0][0].B},
			{Leaf: sortedBounders[0][1].B},
		}}
	}

	xIndex, xScore := splitter(sortedBounders[0], cache)
	yIndex, yScore := splitter(sortedBounders[1], cache)

	var split [2][2][]*flaggedBounder[B]
	if xScore < yScore {
		split = splitBounders(sortedBounders, 0, xIndex)
	} else {
		split = splitBounders(sortedBounders, 1, yIndex)
	}
	return &BVH[B]{
		Branch: []*BVH[B]{
			newBVH(split[0], cache, splitter),
			newBVH(split[1], cache, splitter),
		},
	}
}

// areaDensityBVHSplit chooses a split index that
// minimizes a goodness score, and returns the index and
// score.
//
// The score of a bbox is equal to the surface area times
// the number of segments.
//
// The cache must contain at least len(faces) entries.
func areaDensityBVHSplit[B Bounder](faces []*flaggedBounder[B], cache []float64) (int, float64) {
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
// This is now equivalent to GroupBounders(faces).
//
// This can be used to prepare models for being turned
// into a collider efficiently, or for storing meshes in
// an order well-suited for file compression.
//
// The resulting hierarchy can be passed directly to
// GroupedSegmentsToCollider().
func GroupSegments(faces []*Segment) {
	GroupBounders(faces)
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
func GroupBounders[B Bounder](objects []B) {
	groupBounders(sortBounders(objects), objects)
}

func groupBounders[B Bounder](sortedBounders [2][]*flaggedBounder[B], output []B) {
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

func splitBounders[B Bounder](sortedBounders [2][]*flaggedBounder[B],
	axis, midIdx int) [2][2][]*flaggedBounder[B] {
	for i, b := range sortedBounders[axis] {
		b.Flag = i < midIdx
	}

	separated := [2][]*flaggedBounder[B]{}
	separated[axis] = sortedBounders[axis]

	numObjs := len(sortedBounders[0])
	for newAxis := 0; newAxis < 2; newAxis++ {
		if newAxis == axis {
			continue
		}
		sep := make([]*flaggedBounder[B], numObjs)
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

	return [2][2][]*flaggedBounder[B]{
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

func bestSplitAxis[B Bounder](sortedBounders [2][]*flaggedBounder[B]) int {
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

func sortBounders[B Bounder](bs []B) [2][]*flaggedBounder[B] {
	// Allocate all of the flaggedBounders at once all
	// next to each other in memory.
	flagged := make([]flaggedBounder[B], len(bs))
	for i, b := range bs {
		min, max := b.Min(), b.Max()
		flagged[i] = flaggedBounder[B]{
			B:   b,
			Min: min,
			Max: max,
			Mid: min.Mid(max),
		}
	}

	var result [2][]*flaggedBounder[B]
	for axis := range result {
		bsCopy := make([]*flaggedBounder[B], len(flagged))
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

func multipleBoundsArea[B Bounder](bs []*flaggedBounder[B]) float64 {
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

type flaggedBounder[B Bounder] struct {
	B    B
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
