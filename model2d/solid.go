package model2d

// A Solid defines a two-dimensional shape.
//
// For any given coordinate, the solid can check if the
// shape contains that coordinate.
//
// All methods of a Solid are safe for concurrency.
type Solid interface {
	// Contains must always return false outside of the
	// boundaries of the solid.
	Bounder

	Contains(c Coord) bool
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
	if !InBounds(b, c) {
		return false
	}
	return b.B.Get(int(c.X), int(c.Y))
}

// JoinedSolid combines one or more other solids into a
// single union.
type JoinedSolid []Solid

func (j JoinedSolid) Min() Coord {
	res := j[0].Min()
	for _, s := range j[1:] {
		res = res.Min(s.Min())
	}
	return res
}

func (j JoinedSolid) Max() Coord {
	res := j[0].Max()
	for _, s := range j[1:] {
		res = res.Max(s.Max())
	}
	return res
}

func (j JoinedSolid) Contains(c Coord) bool {
	for _, s := range j {
		if s.Contains(c) {
			return true
		}
	}
	return false
}

// ColliderSolid is a Solid which uses the even-odd test
// for a Collider.
type ColliderSolid struct {
	collider Collider
	min      Coord
	max      Coord
	inset    float64
	radius   float64
}

// NewColliderSolid creates a basic ColliderSolid.
func NewColliderSolid(c Collider) *ColliderSolid {
	return &ColliderSolid{collider: c, min: c.Min(), max: c.Max()}
}

// NewColliderSolidInset creates a ColliderSolid that only
// reports containment at some distance from the surface.
//
// If inset is negative, then the solid is outset from the
// collider.
func NewColliderSolidInset(c Collider, inset float64) *ColliderSolid {
	min := c.Min().Add(Coord{X: inset, Y: inset})
	max := min.Max(c.Max().Sub(Coord{X: inset, Y: inset}))
	return &ColliderSolid{collider: c, min: min, max: max, inset: inset}
}

// NewColliderSolidHollow creates a ColliderSolid that
// only reports containment around the edges.
func NewColliderSolidHollow(c Collider, r float64) *ColliderSolid {
	min := c.Min().Sub(Coord{X: r, Y: r})
	max := c.Max().Add(Coord{X: r, Y: r})
	return &ColliderSolid{collider: c, min: min, max: max, radius: r}
}

// Min gets the minimum of the bounding box.
func (c *ColliderSolid) Min() Coord {
	return c.min
}

// Max gets the maximum of the bounding box.
func (c *ColliderSolid) Max() Coord {
	return c.max
}

// Contains checks if coord is in the solid.
func (c *ColliderSolid) Contains(coord Coord) bool {
	if !InBounds(c, coord) {
		return false
	}
	if c.radius != 0 {
		return c.collider.CircleCollision(coord, c.radius)
	}
	return ColliderContains(c.collider, coord, c.inset)
}

type scaledSolid struct {
	Solid Solid
	Scale float64
}

// ScaleSolid creates a new Solid that scales incoming
// coordinates c by 1/s.
// Thus, the new solid is s times larger.
func ScaleSolid(solid Solid, s float64) Solid {
	return &scaledSolid{Solid: solid, Scale: 1 / s}
}

func (s *scaledSolid) Min() Coord {
	return s.Solid.Min().Scale(1 / s.Scale)
}

func (s *scaledSolid) Max() Coord {
	return s.Solid.Max().Scale(1 / s.Scale)
}

func (s *scaledSolid) Contains(c Coord) bool {
	return s.Solid.Contains(c.Scale(s.Scale))
}
