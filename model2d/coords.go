package model2d

import (
	"math"
	"math/rand"
)

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

// NewCoordArray creates a Coord from an array of x and y.
func NewCoordArray(a [2]float64) Coord {
	return Coord{a[0], a[1]}
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

// Scale scales all the coordinates by s and returns the
// new coordinate.
func (c Coord) Scale(s float64) Coord {
	c.X *= s
	c.Y *= s
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
	return math.Sqrt(math.Pow(c.X-c1.X, 2) + math.Pow(c.Y-c1.Y, 2))
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
