package numerical

import (
	"math"
	"math/rand"
)

// A Vec3 is a 3-dimensional tuple of floats.
type Vec3 [3]float64

// NewVec3RandomNormal creates a random normal vector.
func NewVec3RandomNormal() Vec3 {
	return Vec3{
		rand.NormFloat64(),
		rand.NormFloat64(),
		rand.NormFloat64(),
	}
}

// Add returns v + v1.
func (v Vec3) Add(v1 Vec3) Vec3 {
	return Vec3{v[0] + v1[0], v[1] + v1[1], v[2] + v1[2]}
}

// Sub returns v - v1.
func (v Vec3) Sub(v1 Vec3) Vec3 {
	return Vec3{v[0] - v1[0], v[1] - v1[1], v[2] - v1[2]}
}

// Scale returns v * f.
func (v Vec3) Scale(f float64) Vec3 {
	return Vec3{v[0] * f, v[1] * f, v[2] * f}
}

// Dist gets the Euclidean distance ||v - v1||.
func (v Vec3) Dist(v1 Vec3) float64 {
	dx := v[0] - v1[0]
	dy := v[1] - v1[1]
	dz := v[2] - v1[2]
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// Sum gets the sum of the components.
func (v Vec3) Sum() float64 {
	return v[0] + v[1] + v[2]
}
