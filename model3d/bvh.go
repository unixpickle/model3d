package model3d

import (
	"math"
	"sort"
)

// GeneralBVH represents a (possibly unbalanced)
// axis-aligned bounding volume hierarchy.
//
// A GeneralBVH can store arbitrary Bounders.
// For a triangle-specific version, see BNH.
type GeneralBVH struct {
	// Leaf, if non-nil, is the final bounder.
	Leaf Bounder

	// Branch, if Leaf is nil, points to two children.
	Branch []*GeneralBVH
}

// NewGeneralBVHAreaDensity creates a GeneralBVH by
// minimizing the product of each bounding box's area with
// the number of objects contained in the bounding box at
// each branch.
//
// This is good for efficient ray collision detection.
func NewGeneralBVHAreaDensity(objects []Bounder) *GeneralBVH {
	return newGeneralBVH(sortBounders(objects), make([]float64, len(objects)),
		areaDensityBVHSplit)
}

func newGeneralBVH(sortedBounders [3][]*flaggedBounder, cache []float64,
	splitter func([]*flaggedBounder, []float64) (int, float64)) *GeneralBVH {
	numTris := len(sortedBounders[0])
	if numTris == 0 {
		panic("empty sorted triangles")
	} else if numTris == 1 {
		return &GeneralBVH{Leaf: sortedBounders[0][0].B}
	} else if numTris == 2 {
		return &GeneralBVH{Branch: []*GeneralBVH{
			{Leaf: sortedBounders[0][0].B},
			{Leaf: sortedBounders[0][1].B},
		}}
	}

	xIndex, xScore := splitter(sortedBounders[0], cache)
	yIndex, yScore := splitter(sortedBounders[1], cache)
	zIndex, zScore := splitter(sortedBounders[2], cache)

	var split [2][3][]*flaggedBounder
	if xScore < yScore && xScore < zScore {
		split = splitBounders(sortedBounders, 0, xIndex)
	} else if yScore < xScore && yScore < zScore {
		split = splitBounders(sortedBounders, 1, yIndex)
	} else {
		split = splitBounders(sortedBounders, 2, zIndex)
	}
	return &GeneralBVH{
		Branch: []*GeneralBVH{
			newGeneralBVH(split[0], cache, splitter),
			newGeneralBVH(split[1], cache, splitter),
		},
	}
}

// BVH represents a (possibly unbalanced) axis-aligned
// bounding volume hierarchy of triangles.
//
// A BVH can be used to accelerate collision detection.
// See BVHToCollider() for more details.
//
// A BVH node is either a leaf (a triangle), or a branch
// with two or more children.
//
// For a more generic BVH that supports any object rather
// than just triangles, see GeneralBVH.
type BVH struct {
	// Leaf, if non-nil, is the final triangle.
	Leaf *Triangle

	// Branch, if Leaf is nil, points to two children.
	Branch []*BVH
}

// NewBVHAreaDensity is like NewGeneralBVHAreaDensity but
// for triangles.
func NewBVHAreaDensity(tris []*Triangle) *BVH {
	return generalBVHToBVH(NewGeneralBVHAreaDensity(trianglesToBounders(tris)))
}

func generalBVHToBVH(g *GeneralBVH) *BVH {
	if g.Leaf != nil {
		return &BVH{Leaf: g.Leaf.(*Triangle)}
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
// the number of triangles.
//
// The cache must contain at least len(tris) entries.
func areaDensityBVHSplit(tris []*flaggedBounder, cache []float64) (int, float64) {
	// Fill the cache with scores going in the other
	// direction.
	min, max := tris[len(tris)-1].Min, tris[len(tris)-1].Max
	for i := len(tris) - 2; i >= 0; i-- {
		cache[i] = boundsArea(min, max) * float64(len(tris)-i-1)
		t := tris[i]
		min = min.Min(t.Min)
		max = max.Max(t.Max)
	}

	var bestScore float64
	var bestIndex int

	min, max = tris[0].Min, tris[0].Max
	for i := 1; i < len(tris)-1; i++ {
		t := tris[i]
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

// GroupTriangles is like GroupBounders, but for triangles
// in particular.
//
// This can be used to prepare models for being turned
// into a collider efficiently, or for storing meshes in
// an order well-suited for file compression.
//
// The resulting hierarchy can be passed directly to
// GroupedTrianglesToCollider().
func GroupTriangles(tris []*Triangle) {
	bs := trianglesToBounders(tris)
	groupBounders(sortBounders(bs), bs)
	for i, b := range bs {
		tris[i] = b.(*Triangle)
	}
}

// GroupBounders sorts a slice of objects into a balanced
// bounding volume hierarchy.
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

func groupBounders(sortedBounders [3][]*flaggedBounder, output []Bounder) {
	numTris := len(sortedBounders[0])
	if numTris == 2 {
		// The area-based splitting criterion doesn't
		// distinguish between axes, now.
		output[0] = sortedBounders[0][0].B
		output[1] = sortedBounders[0][1].B
		return
	} else if numTris == 1 {
		output[0] = sortedBounders[0][0].B
		return
	} else if numTris == 0 {
		return
	}

	midIdx := numTris / 2
	axis := bestSplitAxis(sortedBounders)

	separated := splitBounders(sortedBounders, axis, midIdx)
	groupBounders(separated[0], output[:midIdx])
	groupBounders(separated[1], output[midIdx:])
}

func splitBounders(sortedBounders [3][]*flaggedBounder, axis, midIdx int) [2][3][]*flaggedBounder {
	for i, b := range sortedBounders[axis] {
		b.Flag = i < midIdx
	}

	separated := [3][]*flaggedBounder{}
	separated[axis] = sortedBounders[axis]

	numTris := len(sortedBounders[0])
	for newAxis := 0; newAxis < 3; newAxis++ {
		if newAxis == axis {
			continue
		}
		sep := make([]*flaggedBounder, numTris)
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

	return [2][3][]*flaggedBounder{
		{
			separated[0][:midIdx],
			separated[1][:midIdx],
			separated[2][:midIdx],
		},
		{
			separated[0][midIdx:],
			separated[1][midIdx:],
			separated[2][midIdx:],
		},
	}
}

func bestSplitAxis(sortedBounders [3][]*flaggedBounder) int {
	midIdx := len(sortedBounders[0]) / 2

	areaForAxis := func(axis int) float64 {
		return multipleBoundsArea(sortedBounders[axis][:midIdx]) +
			multipleBoundsArea(sortedBounders[axis][midIdx:])
	}

	axis := 0
	minArea := areaForAxis(0)
	for i := 1; i < 3; i++ {
		if a := areaForAxis(i); a < minArea {
			minArea = a
			axis = i
		}
	}

	return axis
}

func sortBounders(bs []Bounder) [3][]*flaggedBounder {
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

	var result [3][]*flaggedBounder
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
		} else {
			sort.Slice(bsCopy, func(i, j int) bool {
				return bsCopy[i].Mid.Z < bsCopy[j].Mid.Z
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
		if min1.Z < min.Z {
			min.Z = min1.Z
		}
		max1 := b.Max
		if max1.X > max.X {
			max.X = max1.X
		}
		if max1.Y > max.Y {
			max.Y = max1.Y
		}
		if max1.Z > max.Z {
			max.Z = max1.Z
		}
	}
	return boundsArea(min, max)
}

func boundsArea(min, max Coord3D) float64 {
	diff := max.Sub(min)
	return 2 * (diff.X*(diff.Y+diff.Z) + diff.Y*diff.Z)
}

type flaggedBounder struct {
	B    Bounder
	Min  Coord3D
	Max  Coord3D
	Mid  Coord3D
	Flag bool
}

func sphereTouchesBounds(center Coord3D, r float64, min, max Coord3D) bool {
	return pointToBoundsDistSquared(center, min, max) <= r*r
}

func pointToBoundsDistSquared(center Coord3D, min, max Coord3D) float64 {
	// https://stackoverflow.com/questions/4578967/cube-sphere-intersection-test
	distSquared := 0.0
	for axis := 0; axis < 3; axis++ {
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

func rayCollisionWithBounds(r *Ray, min, max Coord3D) (minFrac, maxFrac float64) {
	minFrac = math.Inf(-1)
	maxFrac = math.Inf(1)
	for axis := 0; axis < 3; axis++ {
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

func trianglesToBounders(tris []*Triangle) []Bounder {
	bs := make([]Bounder, len(tris))
	for i, t := range tris {
		bs[i] = t
	}
	return bs
}
