package model2d

import "math"

// A Bounder is an object contained in a rectangle.
type Bounder interface {
	// Get the corners of a bounding rectangle.
	//
	// A point p satisfies p >= Min and p <= Max if it is
	// within the bounds.
	Min() Coord
	Max() Coord
}

// InBounds returns true if c is contained within the
// bounding rectangle of b.
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
	if max.Y < min.X {
		return false
	}
	return true
}
