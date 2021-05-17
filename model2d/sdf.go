// Generated from templates/sdf.template

package model2d

import (
	"math"
)

// An SDF is a signed distance function.
//
// An SDF returns 0 on the boundary of some surface,
// positive values inside the surface, and negative values
// outside the surface.
// The magnitude is the distance to the surface.
//
// All methods of an SDF are safe for concurrency.
type SDF interface {
	Bounder

	SDF(c Coord) float64
}

// A PointSDF is an SDF that can additionally get the
// nearest point on a surface.
type PointSDF interface {
	SDF

	// PointSDF gets the SDF at c and also returns the
	// nearest point to c on the surface.
	PointSDF(c Coord) (Coord, float64)
}

// A FaceSDF is a PointSDF that can additionally get the
// segment containing the closest point.
type FaceSDF interface {
	PointSDF

	// FaceSDF gets the SDF at c and also returns the
	// nearest point and face to c on the surface.
	FaceSDF(c Coord) (*Segment, Coord, float64)
}

type funcSDF struct {
	min Coord
	max Coord
	f   func(c Coord) float64
}

// FuncSDF creates an SDF from a function.
//
// If the bounds are invalid, FuncSDF() will panic().
// In particular, max must be no less than min, and all
// floating-point values must be finite numbers.
func FuncSDF(min, max Coord, f func(Coord) float64) SDF {
	if !BoundsValid(NewRect(min, max)) {
		panic("invalid bounds")
	}
	return &funcSDF{
		min: min,
		max: max,
		f:   f,
	}
}

func (f *funcSDF) Min() Coord {
	return f.min
}

func (f *funcSDF) Max() Coord {
	return f.max
}

func (f *funcSDF) SDF(c Coord) float64 {
	return f.f(c)
}

type funcPointSDF struct {
	min Coord
	max Coord
	f   func(c Coord) (Coord, float64)
}

// FuncPointSDF creates a PointSDF from a function.
//
// If the bounds are invalid, FuncPointSDF() will panic().
// In particular, max must be no less than min, and all
// floating-point values must be finite numbers.
func FuncPointSDF(min, max Coord, f func(Coord) (Coord, float64)) PointSDF {
	if !BoundsValid(NewRect(min, max)) {
		panic("invalid bounds")
	}
	return &funcPointSDF{
		min: min,
		max: max,
		f:   f,
	}
}

func (f *funcPointSDF) Min() Coord {
	return f.min
}

func (f *funcPointSDF) Max() Coord {
	return f.max
}

func (f *funcPointSDF) SDF(c Coord) float64 {
	_, d := f.f(c)
	return d
}

func (f *funcPointSDF) PointSDF(c Coord) (Coord, float64) {
	return f.f(c)
}

type colliderSDF struct {
	Collider
	Solid      Solid
	Iterations int
}

// ColliderToSDF generates an SDF that uses bisection
// search to approximate the SDF for any Collider.
//
// The iterations argument controls the precision.
// If set to 0, a default of 32 is used.
func ColliderToSDF(c Collider, iterations int) SDF {
	if iterations == 0 {
		iterations = 32
	}
	return &colliderSDF{
		Collider:   c,
		Solid:      NewColliderSolid(c),
		Iterations: iterations,
	}
}

func (c *colliderSDF) SDF(coord Coord) float64 {
	min, max := c.boundDistance(coord)
	for i := 0; i < c.Iterations; i++ {
		mid := (min + max) / 2
		if c.Collider.CircleCollision(coord, mid) {
			max = mid
		} else {
			min = mid
		}
	}
	res := (min + max) / 2
	if !c.Solid.Contains(coord) {
		res *= -1
	}
	return res
}

func (c *colliderSDF) boundDistance(coord Coord) (min, max float64) {
	lastDist := 1.0
	newDist := 1.0
	initial := c.Collider.CircleCollision(coord, lastDist)
	for i := 0; i < c.Iterations; i++ {
		lastDist = newDist
		if initial {
			newDist = lastDist / 2.0
		} else {
			newDist = lastDist * 2.0
		}
		if c.Collider.CircleCollision(coord, newDist) != initial {
			break
		}
	}
	if newDist > lastDist {
		return lastDist, newDist
	} else {
		return newDist, lastDist
	}
}

type meshSDF struct {
	Solid
	MDF *meshDistFunc
}

// MeshToSDF turns a mesh into a FaceSDF.
func MeshToSDF(m *Mesh) FaceSDF {
	faces := m.SegmentSlice()
	GroupSegments(faces)
	return GroupedSegmentsToSDF(faces)
}

// GroupedSegmentsToSDF creates a FaceSDF from a slice
// of segments.
// If the segments are not grouped by GroupSegments(),
// the resulting PointSDF is inefficient.
func GroupedSegmentsToSDF(faces []*Segment) FaceSDF {
	if len(faces) == 0 {
		panic("cannot create empty SDF")
	}
	return &meshSDF{
		Solid: NewColliderSolid(GroupedSegmentsToCollider(faces)),
		MDF:   newMeshDistFunc(faces),
	}
}

func (m *meshSDF) SDF(c Coord) float64 {
	dist := math.Inf(1)
	m.MDF.Dist(c, &dist, nil, nil)
	if m.Solid.Contains(c) {
		return dist
	} else {
		return -dist
	}
}

func (m *meshSDF) PointSDF(c Coord) (Coord, float64) {
	dist := math.Inf(1)
	point := Coord{}
	m.MDF.Dist(c, &dist, &point, nil)
	if !m.Solid.Contains(c) {
		dist = -dist
	}
	return point, dist
}

func (m *meshSDF) FaceSDF(c Coord) (*Segment, Coord, float64) {
	dist := math.Inf(1)
	point := Coord{}
	var face *Segment
	m.MDF.Dist(c, &dist, &point, &face)
	if !m.Solid.Contains(c) {
		dist = -dist
	}
	return face, point, dist
}

type meshDistFunc struct {
	min Coord
	max Coord

	root     *Segment
	children [2]*meshDistFunc
}

func newMeshDistFunc(faces []*Segment) *meshDistFunc {
	if len(faces) == 1 {
		return &meshDistFunc{root: faces[0], min: faces[0].Min(), max: faces[0].Max()}
	}

	midIdx := len(faces) / 2
	t1 := newMeshDistFunc(faces[:midIdx])
	t2 := newMeshDistFunc(faces[midIdx:])
	return &meshDistFunc{
		min:      t1.Min().Min(t2.Min()),
		max:      t1.Max().Max(t2.Max()),
		children: [2]*meshDistFunc{t1, t2},
	}

}

func (m *meshDistFunc) Min() Coord {
	return m.min
}

func (m *meshDistFunc) Max() Coord {
	return m.max
}

func (m *meshDistFunc) Dist(c Coord, curDist *float64, curPoint *Coord,
	curFace **Segment) {
	if m.root != nil {
		cp := m.root.Closest(c)
		dist := cp.Dist(c)
		if dist < *curDist {
			*curDist = dist
			if curPoint != nil {
				*curPoint = cp
			}
			if curFace != nil {
				*curFace = m.root
			}
		}
		return
	}

	boundDists := [2]float64{
		pointToBoundsDistSquared(c, m.children[0].min, m.children[0].max),
		pointToBoundsDistSquared(c, m.children[1].min, m.children[1].max),
	}
	iterates := m.children
	if boundDists[0] > boundDists[1] {
		iterates[0], iterates[1] = iterates[1], iterates[0]
		boundDists[0], boundDists[1] = boundDists[1], boundDists[0]
	}
	for i, child := range iterates {
		if boundDists[i] > (*curDist)*(*curDist) {
			continue
		}
		child.Dist(c, curDist, curPoint, curFace)
	}
}
