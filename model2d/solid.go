package model2d

// A Solid defines a two-dimensional shape.
//
// For any given coordinate, the solid can check if the
// shape contains that coordinate.
type Solid interface {
	Min() Coord
	Max() Coord
	Contains(c Coord) bool
}

// InSolidBounds returns true if c is contained within the
// bounding rectangular area of s.
func InSolidBounds(s Solid, c Coord) bool {
	min := s.Min()
	max := s.Max()
	return c.Min(min) == min && c.Max(max) == max
}

type bitmapSolid struct {
	B *Bitmap
}

// BitmapToSolid creates a Solid which is true at pixels
// where the bitmap is true, and false elsewhere.
func BitmapToSolid(b *Bitmap) Solid {
	return &bitmapSolid{B: b}
}

func (b *bitmapSolid) Min() Coord {
	return Coord{}
}

func (b *bitmapSolid) Max() Coord {
	return Coord{X: float64(b.B.Width), Y: float64(b.B.Height)}
}

func (b *bitmapSolid) Contains(c Coord) bool {
	if !InSolidBounds(b, c) {
		return false
	}
	return b.B.Get(int(c.X), int(c.Y))
}

type scaledSolid struct {
	Solid Solid
	Scale float64
}

// ScaleSolid creates a new Solid that scales incoming
// coordinates c by 1/s.
// Thus, the new solid is s times larger.
func ScaleSolid(solid Solid, s float64) Solid {
	return &scaledSolid{Solid: solid, Scale: s}
}

func (s *scaledSolid) Min() Coord {
	return s.Solid.Min().Scale(s.Scale)
}

func (s *scaledSolid) Max() Coord {
	return s.Solid.Max().Scale(s.Scale)
}

func (s *scaledSolid) Contains(c Coord) bool {
	return s.Solid.Contains(c.Scale(s.Scale))
}
