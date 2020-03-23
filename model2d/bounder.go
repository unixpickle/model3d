package model2d

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
