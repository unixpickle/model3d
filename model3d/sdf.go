package model3d

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
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

	SDF(c Coord3D) float64
}

// A PointSDF is an SDF that can additionally get the
// nearest point on a surface.
type PointSDF interface {
	SDF

	// PointSDF gets the SDF at c and also returns the
	// nearest point to c on the surface.
	PointSDF(c Coord3D) (Coord3D, float64)
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

func (c *colliderSDF) SDF(coord Coord3D) float64 {
	min, max := c.boundDistance(coord)
	for i := 0; i < c.Iterations; i++ {
		mid := (min + max) / 2
		if c.Collider.SphereCollision(coord, mid) {
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

func (c *colliderSDF) boundDistance(coord Coord3D) (min, max float64) {
	lastDist := 1.0
	newDist := 1.0
	initial := c.Collider.SphereCollision(coord, lastDist)
	for i := 0; i < c.Iterations; i++ {
		lastDist = newDist
		if initial {
			newDist = lastDist / 2.0
		} else {
			newDist = lastDist * 2.0
		}
		if c.Collider.SphereCollision(coord, newDist) != initial {
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
	tris := m.TriangleSlice()
	GroupTriangles(tris)
	return GroupedTrianglesToSDF(tris)
}

// GroupedTrianglesToSDF creates a PointSDF from a slice
// of triangles.
// If the triangles are not grouped by GroupTriangles(),
// the resulting PointSDF is inefficient.
func GroupedTrianglesToSDF(tris []*Triangle) PointSDF {
	if len(tris) == 0 {
		panic("cannot create empty SDF")
	}
	return &meshSDF{
		Solid: NewColliderSolid(GroupedTrianglesToCollider(tris)),
		MDF:   newMeshDistFunc(tris),
	}
}

func (m *meshSDF) SDF(c Coord3D) float64 {
	dist := m.MDF.Dist(c, math.Inf(1))
	if m.Solid.Contains(c) {
		return dist
	} else {
		return -dist
	}
}

func (m *meshSDF) PointSDF(c Coord3D) (Coord3D, float64) {
	point := Coord3D{}
	dist := math.Inf(1)
	m.MDF.PointDist(c, &point, &dist)
	if !m.Solid.Contains(c) {
		dist = -dist
	}
	return point, dist
}

type meshDistFunc struct {
	min Coord3D
	max Coord3D

	root     *Triangle
	children [2]*meshDistFunc
}

func newMeshDistFunc(tris []*Triangle) *meshDistFunc {
	if len(tris) == 1 {
		return &meshDistFunc{root: tris[0], min: tris[0].Min(), max: tris[0].Max()}
	}

	midIdx := len(tris) / 2
	t1 := newMeshDistFunc(tris[:midIdx])
	t2 := newMeshDistFunc(tris[midIdx:])
	return &meshDistFunc{
		min:      t1.Min().Min(t2.Min()),
		max:      t1.Max().Max(t2.Max()),
		children: [2]*meshDistFunc{t1, t2},
	}

}

func (m *meshDistFunc) Min() Coord3D {
	return m.min
}

func (m *meshDistFunc) Max() Coord3D {
	return m.max
}

func (m *meshDistFunc) Dist(c Coord3D, curMin float64) float64 {
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

func (m *meshDistFunc) PointDist(c Coord3D, curPoint *Coord3D, curDist *float64) {
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

type profileSDF struct {
	SDF2D  model2d.SDF
	MinVal Coord3D
	MaxVal Coord3D
}

// ProfileSDF turns a 2D SDF into a 3D SDF by elongating
// the 2D SDF along the Z axis.
func ProfileSDF(sdf2d model2d.SDF, minZ, maxZ float64) SDF {
	min, max := sdf2d.Min(), sdf2d.Max()
	return &profileSDF{
		SDF2D:  sdf2d,
		MinVal: XYZ(min.X, min.Y, minZ),
		MaxVal: XYZ(max.X, max.Y, maxZ),
	}
}

func (p *profileSDF) Min() Coord3D {
	return p.MinVal
}

func (p *profileSDF) Max() Coord3D {
	return p.MaxVal
}

func (p *profileSDF) SDF(c Coord3D) float64 {
	sdf2d := p.SDF2D.SDF(c.XY())
	zDist := math.Min(math.Abs(c.Z-p.MinVal.Z), math.Abs(c.Z-p.MaxVal.Z))
	insideZ := c.Z >= p.MinVal.Z && c.Z <= p.MaxVal.Z
	if !insideZ {
		if sdf2d > 0 {
			// We can go directly to the z-plane and hit the profile.
			return -zDist
		} else {
			// We must go to the z-plane, then to the side of the profile.
			return -math.Sqrt(zDist*zDist + sdf2d*sdf2d)
		}
	}
	if sdf2d > 0 {
		// We are inside the model, so the closest point is either at
		// the face or the side.
		return math.Min(sdf2d, zDist)
	} else {
		// We are outside the model, and the closest point is on the
		// side of the profile.
		return sdf2d
	}
}

type profilePointSDF struct {
	profileSDF
	PointSDF2D model2d.PointSDF
}

// ProfilePointSDF turns a 2D PointSDF into a 3D PointSDF
// by elongating the 2D SDF along the Z axis.
func ProfilePointSDF(sdf2d model2d.PointSDF, minZ, maxZ float64) PointSDF {
	min, max := sdf2d.Min(), sdf2d.Max()
	return &profilePointSDF{
		profileSDF: profileSDF{
			SDF2D:  sdf2d,
			MinVal: XYZ(min.X, min.Y, minZ),
			MaxVal: XYZ(max.X, max.Y, maxZ),
		},
		PointSDF2D: sdf2d,
	}
}

func (p *profilePointSDF) PointSDF(c Coord3D) (Coord3D, float64) {
	point2d, sdf2d := p.PointSDF2D.PointSDF(c.XY())

	minDist := math.Abs(c.Z - p.MinVal.Z)
	maxDist := math.Abs(c.Z - p.MaxVal.Z)
	zDist := math.Min(minDist, maxDist)
	hitZ := p.MinVal.Z
	if maxDist < minDist {
		hitZ = p.MaxVal.Z
	}
	insideZ := c.Z >= p.MinVal.Z && c.Z <= p.MaxVal.Z

	if !insideZ {
		if sdf2d > 0 {
			// We can go directly to the z-plane and hit the profile.
			return XYZ(c.X, c.Y, hitZ), -zDist
		} else {
			// We must go to the z-plane, then to the side of the profile.
			return XYZ(point2d.X, point2d.Y, hitZ),
				-math.Sqrt(zDist*zDist + sdf2d*sdf2d)
		}
	}
	if sdf2d > 0 {
		// We are inside the model, so the closest point is either at
		// the face or the side.
		if zDist < sdf2d {
			return XYZ(c.X, c.Y, hitZ), zDist
		} else {
			return XYZ(point2d.X, point2d.Y, c.Z), sdf2d
		}
	} else {
		// We are outside the model, and the closest point is on the
		// side of the profile.
		return XYZ(point2d.X, point2d.Y, c.Z), sdf2d
	}
}
