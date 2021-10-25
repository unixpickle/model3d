package model2d

import (
	"math"
	"math/rand"
)

var Origin = Coord{}

// A Coord is a coordinate in 2-D Euclidean space.
type Coord struct {
	X float64
	Y float64
}

// NewCoordRandNorm creates a random Coord with normally
// distributed components.
func NewCoordRandNorm() Coord {
	return Coord{
		X: rand.NormFloat64(),
		Y: rand.NormFloat64(),
	}
}

// NewCoordRandUnit creates a random Coord with magnitude
// 1.
func NewCoordRandUnit() Coord {
	for {
		res := NewCoordRandNorm()
		norm := res.Norm()
		// Edge case to avoid numerical issues.
		if norm > 1e-8 {
			return res.Scale(1 / norm)
		}
	}
}

// NewCoordRandUniform creates a random Coord with
// uniformly random coordinates in [0, 1).
func NewCoordRandUniform() Coord {
	return Coord{
		X: rand.Float64(),
		Y: rand.Float64(),
	}
}

// NewCoordRandBounds creates a random Coord uniformly
// inside the given rectangular boundary.
func NewCoordRandBounds(min, max Coord) Coord {
	c := NewCoordRandUniform()
	return c.Mul(max.Sub(min)).Add(min)
}

// NewCoordArray creates a Coord from an array of x and y.
func NewCoordArray(a [2]float64) Coord {
	return Coord{a[0], a[1]}
}

// NewCoordPolar converts polar coordinates to a Coord.
func NewCoordPolar(theta, radius float64) Coord {
	return XY(math.Cos(theta), math.Sin(theta)).Scale(radius)
}

// Ones creates the unit vector scaled by a constant.
func Ones(a float64) Coord {
	return Coord{X: a, Y: a}
}

// XY constructs a coordinate.
func XY(x, y float64) Coord {
	return Coord{X: x, Y: y}
}

// X gets a coordinate in the X direction.
func X(x float64) Coord {
	return Coord{X: x}
}

// Y gets a coordinate in the Y direction.
func Y(y float64) Coord {
	return Coord{Y: y}
}

// Mid computes the midpoint between c and c1.
func (c Coord) Mid(c1 Coord) Coord {
	return c.Add(c1).Scale(0.5)
}

// Norm computes the vector L2 norm.
func (c Coord) Norm() float64 {
	return math.Sqrt(c.X*c.X + c.Y*c.Y)
}

// Dot computes the dot product of c and c1.
func (c Coord) Dot(c1 Coord) float64 {
	return c.X*c1.X + c.Y*c1.Y
}

// Mul computes the element-wise product of c and c1.
func (c Coord) Mul(c1 Coord) Coord {
	return Coord{X: c.X * c1.X, Y: c.Y * c1.Y}
}

// Div computes the element-wise quotient of c / c1.
func (c Coord) Div(c1 Coord) Coord {
	return Coord{X: c.X / c1.X, Y: c.Y / c1.Y}
}

// Recip gets a coordinate as 1 / c.
func (c Coord) Recip() Coord {
	return XY(1/c.X, 1/c.Y)
}

// Scale scales all the coordinates by s and returns the
// new coordinate.
func (c Coord) Scale(s float64) Coord {
	c.X *= s
	c.Y *= s
	return c
}

// AddScalar adds s to all of the coordinates and returns
// the new coordinate.
func (c Coord) AddScalar(s float64) Coord {
	c.X += s
	c.Y += s
	return c
}

// Add computes the sum of c and c1.
func (c Coord) Add(c1 Coord) Coord {
	return Coord{
		X: c.X + c1.X,
		Y: c.Y + c1.Y,
	}
}

// Sub computes c - c1.
func (c Coord) Sub(c1 Coord) Coord {
	return c.Add(c1.Scale(-1))
}

// Dist computes the Euclidean distance to c1.
func (c Coord) Dist(c1 Coord) float64 {
	d1 := c.X - c1.X
	d2 := c.Y - c1.Y
	return math.Sqrt(d1*d1 + d2*d2)
}

// SquaredDist gets the squared Euclidean distance to c1.
func (c Coord) SquaredDist(c1 Coord) float64 {
	d1 := c.X - c1.X
	d2 := c.Y - c1.Y
	return d1*d1 + d2*d2
}

// L1Dist computes the L1 distance to c1.
func (c Coord) L1Dist(c1 Coord) float64 {
	return math.Abs(c.X-c1.X) + math.Abs(c.Y-c1.Y)
}

// Min gets the element-wise minimum of c and c1.
func (c Coord) Min(c1 Coord) Coord {
	return Coord{math.Min(c.X, c1.X), math.Min(c.Y, c1.Y)}
}

// Max gets the element-wise maximum of c and c1.
func (c Coord) Max(c1 Coord) Coord {
	return Coord{math.Max(c.X, c1.X), math.Max(c.Y, c1.Y)}
}

// Sum sums the elements of c.
func (c Coord) Sum() float64 {
	return c.X + c.Y
}

// Normalize gets a unit vector from c.
func (c Coord) Normalize() Coord {
	return c.Scale(1 / c.Norm())
}

// ProjectOut projects the c1 direction out of c.
func (c Coord) ProjectOut(c1 Coord) Coord {
	normed := c1.Normalize()
	return c.Sub(normed.Scale(normed.Dot(c)))
}

// Reflect reflects c1 around c.
func (c Coord) Reflect(c1 Coord) Coord {
	n := c.Normalize()
	return c1.Add(n.Scale(-2 * n.Dot(c1))).Scale(-1)
}

// Array gets an array of x and y.
func (c Coord) Array() [2]float64 {
	return [2]float64{c.X, c.Y}
}

// fastHash generates a hash of the coordinate using a
// dot product with a random vector.
func (c Coord) fastHash() uint32 {
	x := c.fastHash64()
	return uint32(x&0xffffffff) ^ uint32(x>>32)
}

// fastHash64 is like fastHash, but uses a 64-bit hash
// space to help mitigate collisions.
func (c Coord) fastHash64() uint64 {
	// Coefficients are random (keyboard mashed).
	return math.Float64bits(0.78378384728594870293*c.X + 0.12938729312040294193*c.Y)
}
