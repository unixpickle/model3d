package model3d

import "math"

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

// A Coord2D is a coordinate in 2-D Euclidean space.
type Coord2D struct {
	X float64
	Y float64
}

// A Coord3D is a coordinate in 3-D Euclidean space.
type Coord3D struct {
	X float64
	Y float64
	Z float64
}

// Norm computes the vector L2 norm.
func (c Coord3D) Norm() float64 {
	return math.Sqrt(c.X*c.X + c.Y*c.Y + c.Z*c.Z)
}

// Dot computes the dot product of c and c1.
func (c Coord3D) Dot(c1 Coord3D) float64 {
	return c.X*c1.X + c.Y*c1.Y + c.Z*c1.Z
}

// Scale scales all the coordinates by s and returns the
// new coordinate.
func (c Coord3D) Scale(s float64) Coord3D {
	c.X *= s
	c.Y *= s
	c.Z *= s
	return c
}

// Dist computes the Euclidean distance to c1.
func (c Coord3D) Dist(c1 Coord3D) float64 {
	return math.Sqrt(math.Pow(c.X-c1.X, 2) + math.Pow(c.Y-c1.Y, 2) + math.Pow(c.Z-c1.Z, 2))
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
