package model3d

// A Bounder is a shape contained within a rectangular
// volume of space.
type Bounder interface {
	// Get the corners of a bounding rectangular volume.
	//
	// A point p satisfies: p >= Min and p <= Max if it is
	// within the bounds.
	Min() Coord3D
	Max() Coord3D
}

// InBounds returns true if c is contained within the
// bounding rectangular volume of b.
func InBounds(b Bounder, c Coord3D) bool {
	min := b.Min()
	max := b.Max()
	return c.Min(min) == min && c.Max(max) == max
}
