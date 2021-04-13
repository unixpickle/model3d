// Generated from templates/polytope.template

package model3d

import (
	"math"

	"github.com/unixpickle/essentials"
)

// A LinearConstraint defines a half-space of all points c
// such that c.Dot(Normal) <= Max.
type LinearConstraint struct {
	Normal Coord3D
	Max    float64
}

// Contains checks if the half-space contains c.
func (l *LinearConstraint) Contains(c Coord3D) bool {
	return c.Dot(l.Normal) <= l.Max
}

// A ConvexPolytope is the intersection of some linear
// constraints.
type ConvexPolytope []*LinearConstraint

// NewConvexPolytopeRect creates a rectangular convex
// polytope.
func NewConvexPolytopeRect(min, max Coord3D) ConvexPolytope {
	return ConvexPolytope{
		&LinearConstraint{
			Normal: X(1),
			Max:    max.X,
		},
		&LinearConstraint{
			Normal: X(-1),
			Max:    -min.X,
		},
		&LinearConstraint{
			Normal: Y(1),
			Max:    max.Y,
		},
		&LinearConstraint{
			Normal: Y(-1),
			Max:    -min.Y,
		},
		&LinearConstraint{
			Normal: Z(1),
			Max:    max.Z,
		},
		&LinearConstraint{
			Normal: Z(-1),
			Max:    -min.Z,
		},
	}
}

// Contains checks that c satisfies the constraints.
func (c ConvexPolytope) Contains(coord Coord3D) bool {
	for _, l := range c {
		if !l.Contains(coord) {
			return false
		}
	}
	return true
}

// Mesh creates a mesh containing all of the finite faces
// of the polytope.
//
// For complicated polytopes, this may take a long time to
// run, since it is O(n^3) in the constraints.
func (c ConvexPolytope) Mesh() *Mesh {
	m := NewMesh()
	epsilon := c.spatialEpsilon()
	norms := make([]float64, len(c))
	for i, l := range c {
		norms[i] = l.Normal.Norm()
	}
	for i1 := 0; i1 < len(c); i1++ {
		vertices := []Coord3D{}
		for i2 := 0; i2 < len(c)-1; i2++ {
			if i2 == i1 {
				continue
			}
			for i3 := i2 + 1; i3 < len(c); i3++ {
				if i3 == i1 {
					continue
				}
				vertex, found := c.vertex(i1, i2, i3, norms, epsilon)
				if found {
					vertices = append(vertices, vertex)
				}
			}
		}
		if len(vertices) > 2 {
			addConvexFace(m, vertices, c[i1].Normal)
		}
	}

	// Sometimes more than three planes intersect, in
	// which case a bunch of nearly duplicate faces are
	// created.
	m = m.Repair(epsilon)

	// Remove zero-area triangles.
	m.Iterate(func(t *Triangle) {
		if t[0] == t[1] || t[1] == t[2] || t[0] == t[2] {
			m.Remove(t)
		}
	})

	return m
}

// Solid creates a solid out of the polytope.
//
// This runs in O(n^3) in the constraints, so it may be
// unacceptable for large polytopes.
func (c ConvexPolytope) Solid() Solid {
	m := c.Mesh()
	return &polytopeSolid{
		P:      c,
		MinVal: m.Min(),
		MaxVal: m.Max(),
	}
}

func (c ConvexPolytope) vertex(i1, i2, i3 int, norms []float64, epsilon float64) (Coord3D, bool) {
	// Make sure the indices are sorted so that we yield
	// deterministic results for different first faces.
	if i2 < i1 {
		return c.vertex(i2, i1, i3, norms, epsilon)
	} else if i3 < i1 {
		return c.vertex(i3, i2, i1, norms, epsilon)
	} else if i3 < i2 {
		return c.vertex(i1, i3, i2, norms, epsilon)
	}

	l1, l2, l3 := c[i1], c[i2], c[i3]
	matrix := Matrix3{
		l1.Normal.X, l1.Normal.Y, l1.Normal.Z,
		l2.Normal.X, l2.Normal.Y, l2.Normal.Z,
		l3.Normal.X, l3.Normal.Y, l3.Normal.Z,
	}

	// Check for singular (or poorly conditioned) matrix.
	rawArea := norms[i1] * norms[i2] * norms[i3]
	det := matrix.Det()
	if math.Abs(det) < rawArea*1e-8 {
		return Coord3D{}, false
	}

	maxes := Coord3D{l1.Max, l2.Max, l3.Max}
	matrix.InvertInPlaceDet(det)
	solution := matrix.MulColumn(maxes)

	for i, l := range c {
		if i == i1 || i == i2 || i == i3 {
			continue
		}
		if l.Normal.Dot(solution) > l.Max+epsilon*norms[i] {
			return solution, false
		}
	}

	return solution, true
}

func (c ConvexPolytope) spatialEpsilon() float64 {
	// Use an epsilon measure that scales as the planes
	// move further from the origin, since intersections
	// will likely happen further out and result in larger
	// floating point errors.
	var maxMagnitude float64
	for _, eq := range c {
		maxMagnitude = math.Max(maxMagnitude, math.Abs(eq.Max)/eq.Normal.Norm())
	}
	return maxMagnitude * 1e-8
}

func addConvexFace(m *Mesh, vertices []Coord3D, normal Coord3D) {
	center := Coord3D{}
	for _, v := range vertices {
		center = center.Add(v)
	}
	center = center.Scale(1 / float64(len(vertices)))

	basis1, basis2 := normal.OrthoBasis()
	angles := make([]float64, len(vertices))
	for i, v := range vertices {
		diff := v.Sub(center)
		x := basis1.Dot(diff)
		y := basis2.Dot(diff)
		a := math.Atan2(y, x)
		angles[i] = a
	}

	essentials.VoodooSort(angles, func(i, j int) bool {
		return angles[i] < angles[j]
	}, vertices)

	for i := 2; i < len(vertices); i++ {
		t := &Triangle{vertices[0], vertices[i-1], vertices[i]}
		if t.Normal().Dot(normal) < 0 {
			t[0], t[1] = t[1], t[0]
		}
		m.Add(t)
	}
}

type polytopeSolid struct {
	P      ConvexPolytope
	MinVal Coord3D
	MaxVal Coord3D
}

func (p *polytopeSolid) Min() Coord3D {
	return p.MinVal
}

func (p *polytopeSolid) Max() Coord3D {
	return p.MaxVal
}

func (p *polytopeSolid) Contains(c Coord3D) bool {
	return InBounds(p, c) && p.P.Contains(c)
}
