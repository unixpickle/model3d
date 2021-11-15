package model3d

import (
	"math"
	"math/rand"

	"github.com/unixpickle/model3d/model2d"
)

var Origin = Coord3D{}

// A GeoCoord specifies a location on a sphere with a unit
// radius.
//
// The latitude is an angle from -math.Pi/2 to math.pi/2
// representing the North-South direction.
// The longitude is an angle from -math.Pi to math.Pi
// representing the West-East direction.
type GeoCoord struct {
	Lat float64
	Lon float64
}

// Distance gets the Euclidean distance between g and g1
// when traveling on the surface of the sphere.
func (g GeoCoord) Distance(g1 GeoCoord) float64 {
	return math.Acos(math.Min(1, math.Max(-1, g.Coord3D().Dot(g1.Coord3D()))))
}

// Coord3D converts g to Euclidean coordinates on a unit
// sphere centered at the origin.
func (g GeoCoord) Coord3D() Coord3D {
	return Coord3D{
		X: math.Sin(g.Lon) * math.Cos(g.Lat),
		Y: math.Sin(g.Lat),
		Z: math.Cos(g.Lon) * math.Cos(g.Lat),
	}
}

// Normalize brings the latitude and longitude into the
// standard range while (approximately) preserving the
// absolute position.
func (g GeoCoord) Normalize() GeoCoord {
	p := g.Coord3D()
	return p.Geo()
}

type Coord2D = model2d.Coord

// NewCoord2DRandNorm creates a random Coord2D with
// normally distributed components.
func NewCoord2DRandNorm() Coord2D {
	return model2d.NewCoordRandNorm()
}

// NewCoord2DRandUnit creates a random Coord2D with
// magnitude 1.
func NewCoord2DRandUnit() Coord2D {
	return model2d.NewCoordRandUnit()
}

// NewCoord2DRandUniform creates a random Coord2D with
// uniformly random coordinates in [0, 1).
func NewCoord2DRandUniform() Coord2D {
	return model2d.NewCoordRandUniform()
}

// NewCoord2DRandBounds creates a random Coord2D uniformly
// inside the given rectangular boundary.
func NewCoordRandBounds(min, max Coord2D) Coord2D {
	return model2d.NewCoordRandBounds(min, max)
}

// NewCoord2DPolar converts polar coordinates to a
// Coord2D.
func NewCoord2DPolar(theta, radius float64) Coord2D {
	return model2d.NewCoordPolar(theta, radius)
}

// A Coord3D is a coordinate in 3-D Euclidean space.
type Coord3D struct {
	X float64
	Y float64
	Z float64
}

// NewCoord3DArray creates a Coord3D from an array of x,
// y, and z.
func NewCoord3DArray(a [3]float64) Coord3D {
	return Coord3D{a[0], a[1], a[2]}
}

// NewCoord3DRandNorm creates a random Coord3D with
// normally distributed components.
func NewCoord3DRandNorm() Coord3D {
	return Coord3D{
		X: rand.NormFloat64(),
		Y: rand.NormFloat64(),
		Z: rand.NormFloat64(),
	}
}

// NewCoord3DRandUnit creates a random Coord3D with
// magnitude 1.
func NewCoord3DRandUnit() Coord3D {
	for {
		res := NewCoord3DRandNorm()
		norm := res.Norm()
		// Edge case to avoid numerical issues.
		if norm > 1e-8 {
			return res.Scale(1 / norm)
		}
	}
}

// NewCoord3DRandUniform creates a random Coord3D with
// uniformly random coordinates in [0, 1).
func NewCoord3DRandUniform() Coord3D {
	return Coord3D{
		X: rand.Float64(),
		Y: rand.Float64(),
		Z: rand.Float64(),
	}
}

// NewCoord3DRandBounds creates a random Coord3D uniformly
// inside the given rectangular boundary.
func NewCoord3DRandBounds(min, max Coord3D) Coord3D {
	c := NewCoord3DRandUniform()
	return c.Mul(max.Sub(min)).Add(min)
}

// Ones creates the unit vector scaled by a constant.
func Ones(a float64) Coord3D {
	return Coord3D{X: a, Y: a, Z: a}
}

// X gets a coordinate in the X direction.
func X(x float64) Coord3D {
	return Coord3D{X: x}
}

// XY gets a coordinate in the X and Y direction.
func XY(x, y float64) Coord3D {
	return Coord3D{X: x, Y: y}
}

// XYZ creates a 3D coordinate.
func XYZ(x, y, z float64) Coord3D {
	return Coord3D{X: x, Y: y, Z: z}
}

// XZ gets a coordinate in the X and Z direction.
func XZ(x, z float64) Coord3D {
	return Coord3D{X: x, Z: z}
}

// Y gets a coordinate in the Y direction.
func Y(y float64) Coord3D {
	return Coord3D{Y: y}
}

// YZ gets a coordinate in the Y and Z direction.
func YZ(y, z float64) Coord3D {
	return Coord3D{Y: y, Z: z}
}

// Z gets a coordinate in the Z direction.
func Z(z float64) Coord3D {
	return Coord3D{Z: z}
}

// Mid computes the midpoint between c and c1.
func (c Coord3D) Mid(c1 Coord3D) Coord3D {
	return c.Add(c1).Scale(0.5)
}

// Norm computes the vector L2 norm.
func (c Coord3D) Norm() float64 {
	return math.Sqrt(c.X*c.X + c.Y*c.Y + c.Z*c.Z)
}

// NormSquared computes the squared vector L2 norm.
func (c Coord3D) NormSquared() float64 {
	return c.X*c.X + c.Y*c.Y + c.Z*c.Z
}

// Dot computes the dot product of c and c1.
func (c Coord3D) Dot(c1 Coord3D) float64 {
	return c.X*c1.X + c.Y*c1.Y + c.Z*c1.Z
}

// Cross computes the cross product of c and c1.
func (c Coord3D) Cross(c1 Coord3D) Coord3D {
	return Coord3D{
		X: c.Y*c1.Z - c.Z*c1.Y,
		Y: c.Z*c1.X - c.X*c1.Z,
		Z: c.X*c1.Y - c.Y*c1.X,
	}
}

// Mul computes the element-wise product of c and c1.
func (c Coord3D) Mul(c1 Coord3D) Coord3D {
	return XYZ(c.X*c1.X, c.Y*c1.Y, c.Z*c1.Z)
}

// Div computes the element-wise quotient of c / c1.
func (c Coord3D) Div(c1 Coord3D) Coord3D {
	return XYZ(c.X/c1.X, c.Y/c1.Y, c.Z/c1.Z)
}

// Recip gets a coordinate as 1 / c.
func (c Coord3D) Recip() Coord3D {
	return XYZ(1/c.X, 1/c.Y, 1/c.Z)
}

// Scale scales all the coordinates by s and returns the
// new coordinate.
func (c Coord3D) Scale(s float64) Coord3D {
	c.X *= s
	c.Y *= s
	c.Z *= s
	return c
}

// AddScalar adds s to all of the coordinates and returns
// the new coordinate.
func (c Coord3D) AddScalar(s float64) Coord3D {
	c.X += s
	c.Y += s
	c.Z += s
	return c
}

// Add computes the sum of c and c1.
func (c Coord3D) Add(c1 Coord3D) Coord3D {
	return Coord3D{
		X: c.X + c1.X,
		Y: c.Y + c1.Y,
		Z: c.Z + c1.Z,
	}
}

// Sub computes c - c1.
func (c Coord3D) Sub(c1 Coord3D) Coord3D {
	return c.Add(c1.Scale(-1))
}

// Dist computes the Euclidean distance to c1.
func (c Coord3D) Dist(c1 Coord3D) float64 {
	d1 := c.X - c1.X
	d2 := c.Y - c1.Y
	d3 := c.Z - c1.Z
	return math.Sqrt(d1*d1 + d2*d2 + d3*d3)
}

// SquaredDist gets the squared Euclidean distance to c1.
func (c Coord3D) SquaredDist(c1 Coord3D) float64 {
	d1 := c.X - c1.X
	d2 := c.Y - c1.Y
	d3 := c.Z - c1.Z
	return d1*d1 + d2*d2 + d3*d3
}

// L1Dist computes the L1 distance to c1.
func (c Coord3D) L1Dist(c1 Coord3D) float64 {
	return math.Abs(c.X-c1.X) + math.Abs(c.Y-c1.Y) + math.Abs(c.Z-c1.Z)
}

// Geo computes a normalized geo coordinate.
func (c Coord3D) Geo() GeoCoord {
	if c.Norm() == 0 {
		return GeoCoord{}
	}
	p := c.Scale(1 / c.Norm())
	g := GeoCoord{}
	g.Lat = math.Asin(p.Y)
	g.Lon = math.Atan2(p.X, p.Z)
	return g
}

// Coord2D projects c onto the x,y plane and drops the Z
// value.
// It is equivalent to c.XY().
func (c Coord3D) Coord2D() Coord2D {
	return c.XY()
}

// XY gets (x, y) as a 2D coordinate.
func (c Coord3D) XY() Coord2D {
	return Coord2D{X: c.X, Y: c.Y}
}

// XZ gets (x, z) as a 2D coordinate.
func (c Coord3D) XZ() Coord2D {
	return Coord2D{X: c.X, Y: c.Z}
}

// YX gets (y, x) as a 2D coordinate.
func (c Coord3D) YX() Coord2D {
	return Coord2D{X: c.Y, Y: c.X}
}

// YZ gets (y, z) as a 2D coordinate.
func (c Coord3D) YZ() Coord2D {
	return Coord2D{X: c.Y, Y: c.Z}
}

// ZX gets (z, x) as a 2D coordinate.
func (c Coord3D) ZX() Coord2D {
	return Coord2D{X: c.Z, Y: c.X}
}

// ZY gets (z, y) as a 2D coordinate.
func (c Coord3D) ZY() Coord2D {
	return Coord2D{X: c.Z, Y: c.Y}
}

// Min gets the element-wise minimum of c and c1.
func (c Coord3D) Min(c1 Coord3D) Coord3D {
	return Coord3D{math.Min(c.X, c1.X), math.Min(c.Y, c1.Y), math.Min(c.Z, c1.Z)}
}

// Max gets the element-wise maximum of c and c1.
func (c Coord3D) Max(c1 Coord3D) Coord3D {
	return Coord3D{math.Max(c.X, c1.X), math.Max(c.Y, c1.Y), math.Max(c.Z, c1.Z)}
}

// Sum sums the elements of c.
func (c Coord3D) Sum() float64 {
	return c.X + c.Y + c.Z
}

// Normalize gets a unit vector from c.
func (c Coord3D) Normalize() Coord3D {
	return c.Scale(1 / c.Norm())
}

// OrthoBasis creates two unit vectors which are
// orthogonal to c and to each other.
//
// If c is axis-aligned, the other vectors will be as
// well.
func (c Coord3D) OrthoBasis() (Coord3D, Coord3D) {
	absX := math.Abs(c.X)
	absY := math.Abs(c.Y)
	absZ := math.Abs(c.Z)

	// Create the first basis vector by swapping two
	// coordinates and negating one of them.
	// For numerical stability, we involve the component
	// with the largest absolute value.
	var basis1 Coord3D
	if absX > absY && absX > absZ {
		basis1.X = c.Y / absX
		basis1.Y = -c.X / absX
	} else {
		basis1.Y = c.Z
		basis1.Z = -c.Y
		if absY > absZ {
			basis1.Y /= absY
			basis1.Z /= absY
		} else {
			basis1.Y /= absZ
			basis1.Z /= absZ
		}
	}

	// Create the second basis vector using a cross product.
	basis2 := Coord3D{
		basis1.Y*c.Z - basis1.Z*c.Y,
		basis1.Z*c.X - basis1.X*c.Z,
		basis1.X*c.Y - basis1.Y*c.X,
	}

	return basis1.Normalize(), basis2.Normalize()
}

// ProjectOut projects the c1 direction out of c.
func (c Coord3D) ProjectOut(c1 Coord3D) Coord3D {
	normed := c1.Normalize()
	return c.Sub(normed.Scale(normed.Dot(c)))
}

// Reflect reflects c1 around c on the plane spanned by
// both vectors.
func (c Coord3D) Reflect(c1 Coord3D) Coord3D {
	n := c.Normalize()
	return c1.Add(n.Scale(-2 * n.Dot(c1))).Scale(-1)
}

// Array creates an array with the x, y, and z.
// This can be useful for some vectorized code.
func (c Coord3D) Array() [3]float64 {
	return [3]float64{c.X, c.Y, c.Z}
}

// fastHash generates a hash of the coordinate using a
// dot product with a random vector.
func (c Coord3D) fastHash() uint32 {
	x := c.fastHash64()
	return uint32(x&0xffffffff) ^ uint32(x>>32)
}

// fastHash64 is like fastHash, but uses a 64-bit hash
// space to help mitigate collisions.
func (c Coord3D) fastHash64() uint64 {
	// Coefficients are random (keyboard mashed).
	return math.Float64bits(0.78378384728594870293*c.X + 0.12938729312040294193*c.Y +
		0.98439472938948227499*c.Z)
}
