package numerical

import (
	"math"
	"math/rand"
)

// A Vec2 is a 2-dimensional tuple of floats.
type Vec2 [2]float64

// NewVec2RandomNormal creates a random normal vector.
func NewVec2RandomNormal() Vec2 {
	return Vec2{
		rand.NormFloat64(),
		rand.NormFloat64(),
	}
}

// NewVec2RandomUnit creates a unit-length Vec2 in a random
// direction.
func NewVec2RandomUnit() Vec2 {
	for {
		res := NewVec2RandomNormal()
		norm := res.Norm()
		// Edge case to avoid numerical issues.
		if norm > 1e-8 {
			return res.Scale(1 / norm)
		}
	}
}

// Add returns v + v1.
func (v Vec2) Add(v1 Vec2) Vec2 {
	return Vec2{v[0] + v1[0], v[1] + v1[1]}
}

// Sub returns v - v1.
func (v Vec2) Sub(v1 Vec2) Vec2 {
	return Vec2{v[0] - v1[0], v[1] - v1[1]}
}

// Scale returns v * f.
func (v Vec2) Scale(f float64) Vec2 {
	return Vec2{v[0] * f, v[1] * f}
}

// Dist gets the Euclidean distance ||v - v1||.
func (v Vec2) Dist(v1 Vec2) float64 {
	dx := v[0] - v1[0]
	dy := v[1] - v1[1]
	return math.Sqrt(dx*dx + dy*dy)
}

// Dot computes (v dot v1).
func (v Vec2) Dot(v1 Vec2) float64 {
	return v[0]*v1[0] + v[1]*v1[1]
}

// Norm computes ||v||.
func (v Vec2) Norm() float64 {
	return math.Sqrt(v.Dot(v))
}

// Sum gets the sum of the components.
func (v Vec2) Sum() float64 {
	return v[0] + v[1]
}

// Normalize gets a unit vector from v.
func (v Vec2) Normalize() Vec2 {
	return v.Scale(1 / v.Norm())
}

// ProjectOut projects the v1 direction out of v.
func (v Vec2) ProjectOut(v1 Vec2) Vec2 {
	normed := v1.Normalize()
	return v.Add(normed.Scale(-normed.Dot(v)))
}

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

// NewVec3RandomUnit creates a unit-length Vec3 in a random
// direction.
func NewVec3RandomUnit() Vec3 {
	for {
		res := NewVec3RandomNormal()
		norm := res.Norm()
		// Edge case to avoid numerical issues.
		if norm > 1e-8 {
			return res.Scale(1 / norm)
		}
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

// Dot computes (v dot v1).
func (v Vec3) Dot(v1 Vec3) float64 {
	return v[0]*v1[0] + v[1]*v1[1] + v[2]*v1[2]
}

// Cross computes the cross product of v and v1.
func (v Vec3) Cross(v1 Vec3) Vec3 {
	return Vec3{
		v[1]*v1[2] - v[2]*v1[1],
		v[2]*v1[0] - v[0]*v1[2],
		v[0]*v1[1] - v[1]*v1[0],
	}
}

// Norm computes ||v||.
func (v Vec3) Norm() float64 {
	return math.Sqrt(v.Dot(v))
}

// Sum gets the sum of the components.
func (v Vec3) Sum() float64 {
	return v[0] + v[1] + v[2]
}

// Normalize gets a unit vector from v.
func (v Vec3) Normalize() Vec3 {
	return v.Scale(1 / v.Norm())
}

// ProjectOut projects the v1 direction out of v.
func (v Vec3) ProjectOut(v1 Vec3) Vec3 {
	normed := v1.Normalize()
	return v.Add(normed.Scale(-normed.Dot(v)))
}

// OrthoBasis creates two unit vectors which are
// orthogonal to v and to each other.
//
// If v is axis-aligned, the other vectors will be as
// well.
func (v Vec3) OrthoBasis() (Vec3, Vec3) {
	absX := math.Abs(v[0])
	absY := math.Abs(v[1])
	absZ := math.Abs(v[2])

	// Create the first basis vector by swapping two
	// coordinates and negating one of them.
	// For numerical stability, we involve the component
	// with the largest absolute value.
	var basis1 Vec3
	if absX > absY && absX > absZ {
		basis1[0] = v[1] / absX
		basis1[1] = -v[0] / absX
	} else {
		basis1[1] = v[2]
		basis1[2] = -v[1]
		if absY > absZ {
			basis1[1] /= absY
			basis1[2] /= absY
		} else {
			basis1[1] /= absZ
			basis1[2] /= absZ
		}
	}

	basis2 := basis1.Cross(v)
	return basis1.Normalize(), basis2.Normalize()
}

// A Vec4 is a 4-dimensional tuple of floats.
type Vec4 [4]float64

// NewVec4RandomNormal creates a random normal vector.
func NewVec4RandomNormal() Vec4 {
	return Vec4{
		rand.NormFloat64(),
		rand.NormFloat64(),
		rand.NormFloat64(),
		rand.NormFloat64(),
	}
}

// NewVec4RandomUnit creates a unit-length Vec3 in a random
// direction.
func NewVec4RandomUnit() Vec4 {
	for {
		res := NewVec4RandomNormal()
		norm := res.Norm()
		// Edge case to avoid numerical issues.
		if norm > 1e-8 {
			return res.Scale(1 / norm)
		}
	}
}

// Add returns v + v1.
func (v Vec4) Add(v1 Vec4) Vec4 {
	return Vec4{v[0] + v1[0], v[1] + v1[1], v[2] + v1[2], v[3] + v1[3]}
}

// Sub returns v - v1.
func (v Vec4) Sub(v1 Vec4) Vec4 {
	return Vec4{v[0] - v1[0], v[1] - v1[1], v[2] - v1[2], v[3] - v1[3]}
}

// Scale returns v * f.
func (v Vec4) Scale(f float64) Vec4 {
	return Vec4{v[0] * f, v[1] * f, v[2] * f, v[3] * f}
}

// Dist gets the Euclidean distance ||v - v1||.
func (v Vec4) Dist(v1 Vec4) float64 {
	dx := v[0] - v1[0]
	dy := v[1] - v1[1]
	dz := v[2] - v1[2]
	dt := v[3] - v1[3]
	return math.Sqrt(dx*dx + dy*dy + dz*dz + dt*dt)
}

// Dot computes (v dot v1).
func (v Vec4) Dot(v1 Vec4) float64 {
	return v[0]*v1[0] + v[1]*v1[1] + v[2]*v1[2] + v[3]*v1[3]
}

// Norm computes ||v||.
func (v Vec4) Norm() float64 {
	return math.Sqrt(v.Dot(v))
}

// Sum gets the sum of the components.
func (v Vec4) Sum() float64 {
	return v[0] + v[1] + v[2] + v[3]
}

// Normalize gets a unit vector from v.
func (v Vec4) Normalize() Vec4 {
	return v.Scale(1 / v.Norm())
}

// ProjectOut projects the v1 direction out of v.
func (v Vec4) ProjectOut(v1 Vec4) Vec4 {
	normed := v1.Normalize()
	return v.Add(normed.Scale(-normed.Dot(v)))
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

// DistSquared computes ||v-v1||^2.
func (v Vec) DistSquared(v1 Vec) float64 {
	var sum float64
	for i, x := range v {
		diff := (x - v1[i])
		sum += diff * diff
	}
	return sum
}

// Dist computes ||v-v1||.
func (v Vec) Dist(v1 Vec) float64 {
	return math.Sqrt(v.DistSquared(v1))
}

// ProjectOut projects the direction v1 out of v.
func (v Vec) ProjectOut(v1 Vec) Vec {
	normed := v1.Normalize()
	return v.Sub(normed.Scale(normed.Dot(v)))
}
