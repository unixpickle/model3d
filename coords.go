package model3d

import "math"

import "github.com/unixpickle/model3d/model2d"

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

// A Coord3D is a coordinate in 3-D Euclidean space.
type Coord3D struct {
	X float64
	Y float64
	Z float64
}

func newCoord3DArray(a [3]float64) Coord3D {
	return Coord3D{a[0], a[1], a[2]}
}

// Mid computes the midpoint between c and c1.
func (c Coord3D) Mid(c1 Coord3D) Coord3D {
	return c.Add(c1).Scale(0.5)
}

// Norm computes the vector L2 norm.
func (c Coord3D) Norm() float64 {
	return math.Sqrt(c.X*c.X + c.Y*c.Y + c.Z*c.Z)
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

// Scale scales all the coordinates by s and returns the
// new coordinate.
func (c Coord3D) Scale(s float64) Coord3D {
	c.X *= s
	c.Y *= s
	c.Z *= s
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
func (c Coord3D) Coord2D() Coord2D {
	return Coord2D{X: c.X, Y: c.Y}
}

// Min gets the element-wise minimum of c and c1.
func (c Coord3D) Min(c1 Coord3D) Coord3D {
	return Coord3D{math.Min(c.X, c1.X), math.Min(c.Y, c1.Y), math.Min(c.Z, c1.Z)}
}

// Max gets the element-wise maximum of c and c1.
func (c Coord3D) Max(c1 Coord3D) Coord3D {
	return Coord3D{math.Max(c.X, c1.X), math.Max(c.Y, c1.Y), math.Max(c.Z, c1.Z)}
}

// Normalize gets a unit vector from c.
func (c Coord3D) Normalize() Coord3D {
	return c.Scale(1 / c.Norm())
}

// OrthoBasis creates two unit vectors which are
// orthogonal to c and to each other.
func (c Coord3D) OrthoBasis() (Coord3D, Coord3D) {
	// Create the first basis vector by swapping two
	// coordinates and negating one of them.
	// For numerical stability, we involve the component
	// with the largest absolute value.
	var basis1 Coord3D
	if math.Abs(c.X) > math.Abs(c.Y) && math.Abs(c.X) > math.Abs(c.Z) {
		basis1.X = c.Y
		basis1.Y = -c.X
	} else {
		basis1.Y = c.Z
		basis1.Z = -c.Y
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

func (c Coord3D) array() [3]float64 {
	return [3]float64{c.X, c.Y, c.Z}
}
