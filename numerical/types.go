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

// Vec is a vector of arbitrary dimension.
type Vec []float64

// Add returns v + v1.
func (v Vec) Add(v1 Vec) Vec {
	if len(v) != len(v1) {
		panic("vector lengths do not match")
	}
	res := make(Vec, len(v))
	for i, x := range v {
		y := v1[i]
		res[i] = x + y
	}
	return res
}

// Sub returns v - v1.
func (v Vec) Sub(v1 Vec) Vec {
	if len(v) != len(v1) {
		panic("vector lengths do not match")
	}
	res := make(Vec, len(v))
	for i, x := range v {
		y := v1[i]
		res[i] = x - y
	}
	return res
}

// Dot returns the dot product of v and v1.
func (v Vec) Dot(v1 Vec) float64 {
	if len(v) != len(v1) {
		panic("vector lengths do not match")
	}
	var res float64
	for i, x := range v {
		y := v1[i]
		res += x * y
	}
	return res
}

// Scale returns v * s.
func (v Vec) Scale(s float64) Vec {
	res := make(Vec, len(v))
	for i, x := range v {
		res[i] = x * s
	}
	return res
}

// NormSquared returns ||v||^2.
func (v Vec) NormSquared() float64 {
	var res float64
	for _, x := range v {
		res += x * x
	}
	return res
}

// Norm returns ||v||.
func (v Vec) Norm() float64 {
	return math.Sqrt(v.NormSquared())
}

// Normalize returns v/||v||.
func (v Vec) Normalize() Vec {
	return v.Scale(1 / v.Norm())
}

// ProjectOut projects the direction v1 out of v.
func (v Vec) ProjectOut(v1 Vec) Vec {
	normed := v1.Normalize()
	return v.Sub(normed.Scale(normed.Dot(v)))
}
