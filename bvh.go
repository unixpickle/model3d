package model3d

import (
	"math"
	"sort"
)

// BVH represents a (possibly unbalanced) axis-aligned
// bounding volume hierarchy of triangles.
//
// A BVH can be used to accelerate collision detection.
// See BVHToCollider() for more details.
//
// A BVH node is either a leaf (a triangle), or a branch
// with exactly two children.
type BVH struct {
	// Leaf, if non-nil, is the final triangle.
	Leaf *Triangle

	// Branch, if Leaf is nil, points to two children.
	Branch [2]*BVH
}

// NewBVHAreaDensity creates a BVH by minimizing the
// product of each bounding box's area with the number of
// triangles contained in the bounding box.
//
// This is good for efficient ray collision detection.
func NewBVHAreaDensity(triangles []*Triangle) *BVH {
	return newBVH(sortTriangles(triangles), make([]float64, len(triangles)),
		areaDensityBVHSplit)
}

func newBVH(sortedTris [3][]*flaggedTriangle, cache []float64,
	splitter func([]*flaggedTriangle, []float64) (int, float64)) *BVH {
	numTris := len(sortedTris[0])
	if numTris == 0 {
		panic("empty sorted triangles")
	} else if numTris == 1 {
		return &BVH{Leaf: sortedTris[0][0].T}
	} else if numTris == 2 {
		return &BVH{Branch: [2]*BVH{
			{Leaf: sortedTris[0][0].T},
			{Leaf: sortedTris[0][1].T},
		}}
	}

	xIndex, xScore := splitter(sortedTris[0], cache)
	yIndex, yScore := splitter(sortedTris[1], cache)
	zIndex, zScore := splitter(sortedTris[2], cache)

	var split [2][3][]*flaggedTriangle
	if xScore < yScore && xScore < zScore {
		split = splitTriangles(sortedTris, 0, xIndex)
	} else if yScore < xScore && yScore < zScore {
		split = splitTriangles(sortedTris, 1, yIndex)
	} else {
		split = splitTriangles(sortedTris, 2, zIndex)
	}
	return &BVH{
		Branch: [2]*BVH{
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
// the number of triangles.
//
// The cache must contain at least len(tris) entries.
func areaDensityBVHSplit(tris []*flaggedTriangle, cache []float64) (int, float64) {
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

// GroupTriangles sorts the triangle slice into a balanced
// bounding volume hierarchy.
// In particular, the sorted slice can be recursively cut
// in half, and each half will be spatially separated as
// well as possible along some axis.
//
// This can be used to prepare models for being turned
// into a collider efficiently, or for storing meshes in
// an order well-suited for file compression.
//
// The resulting hierarchy can be passed directly to
// GroupedTrianglesToCollider().
func GroupTriangles(tris []*Triangle) {
	groupTriangles(sortTriangles(tris), tris)
}

func groupTriangles(sortedTris [3][]*flaggedTriangle, output []*Triangle) {
	numTris := len(sortedTris[0])
	if numTris == 2 {
		// The area-based splitting criterion doesn't
		// distinguish between axes, now.
		output[0] = sortedTris[0][0].T
		output[1] = sortedTris[0][1].T
		return
	} else if numTris == 1 {
		output[0] = sortedTris[0][0].T
		return
	} else if numTris == 0 {
		return
	}

	midIdx := numTris / 2
	axis := bestSplitAxis(sortedTris)

	separated := splitTriangles(sortedTris, axis, midIdx)
	groupTriangles(separated[0], output[:midIdx])
	groupTriangles(separated[1], output[midIdx:])
}

func splitTriangles(sortedTris [3][]*flaggedTriangle, axis, midIdx int) [2][3][]*flaggedTriangle {
	for i, t := range sortedTris[axis] {
		t.Flag = i < midIdx
	}

	separated := [3][]*flaggedTriangle{}
	separated[axis] = sortedTris[axis]

	numTris := len(sortedTris[0])
	for newAxis := 0; newAxis < 3; newAxis++ {
		if newAxis == axis {
			continue
		}
		sep := make([]*flaggedTriangle, numTris)
		idx0 := 0
		idx1 := midIdx
		for _, t := range sortedTris[newAxis] {
			if t.Flag {
				sep[idx0] = t
				idx0++
			} else {
				sep[idx1] = t
				idx1++
			}
		}
		separated[newAxis] = sep
	}

	return [2][3][]*flaggedTriangle{
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

func bestSplitAxis(sortedTris [3][]*flaggedTriangle) int {
	midIdx := len(sortedTris[0]) / 2

	areaForAxis := func(axis int) float64 {
		return triangleBoundArea(sortedTris[axis][:midIdx]) +
			triangleBoundArea(sortedTris[axis][midIdx:])
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

func sortTriangles(tris []*Triangle) [3][]*flaggedTriangle {
	// Allocate all of the flaggedTriangles at once all
	// next to each other in memory.
	ts := make([]flaggedTriangle, len(tris))
	for i, t := range tris {
		min, max := t.Min(), t.Max()
		ts[i] = flaggedTriangle{
			T:   t,
			Min: min,
			Max: max,
			Mid: min.Mid(max),
		}
	}

	var result [3][]*flaggedTriangle
	for axis := range result {
		tsCopy := make([]*flaggedTriangle, len(ts))
		for i := range ts {
			tsCopy[i] = &ts[i]
		}
		if axis == 0 {
			sort.Slice(tsCopy, func(i, j int) bool {
				return tsCopy[i].Mid.X < tsCopy[j].Mid.X
			})
		} else if axis == 1 {
			sort.Slice(tsCopy, func(i, j int) bool {
				return tsCopy[i].Mid.Y < tsCopy[j].Mid.Y
			})
		} else {
			sort.Slice(tsCopy, func(i, j int) bool {
				return tsCopy[i].Mid.Z < tsCopy[j].Mid.Z
			})
		}
		result[axis] = tsCopy
	}
	return result
}

func triangleBoundArea(tris []*flaggedTriangle) float64 {
	min, max := tris[0].Min, tris[0].Max
	for i := 1; i < len(tris); i++ {
		t := tris[i]

		// This is very expanded (unwrapped) vs. using
		// Min() and Max(), but it is faster and this is
		// surprisingly a large bottleneck.
		min1 := t.Min
		if min1.X < min.X {
			min.X = min1.X
		}
		if min1.Y < min.Y {
			min.Y = min1.Y
		}
		if min1.Z < min.Z {
			min.Z = min1.Z
		}
		max1 := t.Max
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

type flaggedTriangle struct {
	T    *Triangle
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