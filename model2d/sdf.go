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

// MeshToSDF turns a mesh into a PointSDF.
func MeshToSDF(m *Mesh) PointSDF {
	faces := m.SegmentSlice()
	GroupSegments(faces)
	return GroupedSegmentsToSDF(faces)
}

// GroupedSegmentsToSDF creates a PointSDF from a slice
// of segments.
// If the segments are not grouped by GroupSegments(),
// the resulting PointSDF is inefficient.
func GroupedSegmentsToSDF(faces []*Segment) PointSDF {
	if len(faces) == 0 {
		panic("cannot create empty SDF")
	}
	return &meshSDF{
		Solid: NewColliderSolid(GroupedSegmentsToCollider(faces)),
		MDF:   newMeshDistFunc(faces),
	}
}

func (m *meshSDF) SDF(c Coord) float64 {
	dist := m.MDF.Dist(c, math.Inf(1))
	if m.Solid.Contains(c) {
		return dist
	} else {
		return -dist
	}
}

func (m *meshSDF) PointSDF(c Coord) (Coord, float64) {
	point := Coord{}
	dist := math.Inf(1)
	m.MDF.PointDist(c, &point, &dist)
	if !m.Solid.Contains(c) {
		dist = -dist
	}
	return point, dist
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

func (m *meshDistFunc) Dist(c Coord, curMin float64) float64 {
	if m.root != nil {
		return math.Min(curMin, m.root.Dist(c))
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
		if boundDists[i] > curMin*curMin {
			continue
		}
		curMin = math.Min(curMin, child.Dist(c, curMin))
	}
	return curMin
}

func (m *meshDistFunc) PointDist(c Coord, curPoint *Coord, curDist *float64) {
	if m.root != nil {
		cp := m.root.Closest(c)
		dist := cp.Dist(c)
		if dist < *curDist {
			*curDist = dist
			*curPoint = cp
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
		child.PointDist(c, curPoint, curDist)
	}
}
