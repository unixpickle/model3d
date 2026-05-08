// Generated from templates/bounder.template

package model2d

import (
	"fmt"
	"math"
)

// A Bounder is an object contained in an axis-aligned
// bounding box.
type Bounder interface {
	// Get the corners of a bounding box.
	//
	// A point p satisfies p >= Min and p <= Max if it is
	// within the bounds.
	Min() Coord
	Max() Coord
}

// InBounds returns true if c is contained within the
// bounding box of b.
func InBounds(b Bounder, c Coord) bool {
	min := b.Min()
	max := b.Max()
	return c.Min(min) == min && c.Max(max) == max
}

// BoundsValid checks for numerical issues with the bounds.
func BoundsValid(b Bounder) bool {
	min, max := b.Min(), b.Max()
	if math.IsNaN(min.Sum()) || math.IsNaN(max.Sum()) ||
		math.IsInf(min.Sum(), 0) || math.IsInf(max.Sum(), 0) {
		return false
	}
	if max.X < min.X {
		return false
	}
	if max.Y < min.Y {
		return false
	}
	return true
}

func validateBounds(b Bounder) {
	if !BoundsValid(b) {
		panic(fmt.Sprintf("invalid bounds: min=%#v max=%#v", b.Min(), b.Max()))
	}
}

// BoundsUnion computes the bounds of one or more bounder.
func BoundsUnion[B Bounder](bs []B) (min Coord, max Coord) {
	min = bs[0].Min()
	max = bs[0].Max()
	for _, b := range bs[1:] {
		min = min.Min(b.Min())
		max = max.Max(b.Max())
	}
	return
}

// BoundsIntersection computes the intersection of bounders.
func BoundsIntersection[B Bounder](bs []B) (min, max Coord) {
	min = bs[0].Min()
	max = bs[0].Max()
	for _, b := range bs[1:] {
		min = min.Max(b.Min())
		max = max.Min(b.Max())
	}
	return min.Min(max), max
}

// InsetBounds adds a delta to the min and subtracts it from the max,
// handling collapsed volumes by falling back to a midpoint.
func InsetBounds(min, max Coord, delta float64) (Coord, Coord) {
	min = min.AddScalar(delta)
	max = max.AddScalar(-delta)
	return averageDegenerateBounds(min, max)
}

func averageDegenerateBounds(min, max Coord) (Coord, Coord) {
	if min.X > max.X {
		mid := (min.X + max.X) / 2
		min.X, max.X = mid, mid
	}
	if min.Y > max.Y {
		mid := (min.Y + max.Y) / 2
		min.Y, max.Y = mid, mid
	}
	return min, max
}
