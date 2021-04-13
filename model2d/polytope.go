// Generated from templates/polytope.template

package model2d

import (
	"math"

	"github.com/unixpickle/essentials"
)

// A LinearConstraint defines a half-space of all points c
// such that c.Dot(Normal) <= Max.
type LinearConstraint struct {
	Normal Coord
	Max    float64
}

// Contains checks if the half-space contains c.
func (l *LinearConstraint) Contains(c Coord) bool {
	return c.Dot(l.Normal) <= l.Max
}

// A ConvexPolytope is the intersection of some linear
// constraints.
type ConvexPolytope []*LinearConstraint

// NewConvexPolytopeRect creates a rectangular convex
// polytope.
func NewConvexPolytopeRect(min, max Coord) ConvexPolytope {
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
	}
}

// Contains checks that c satisfies the constraints.
func (c ConvexPolytope) Contains(coord Coord) bool {
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
// run, since it is O(n^2) in the constraints.
func (c ConvexPolytope) Mesh() *Mesh {
	m := NewMesh()
	epsilon := c.spatialEpsilon()
	norms := make([]float64, len(c))
	for i, l := range c {
		norms[i] = l.Normal.Norm()
	}
	for i1 := 0; i1 < len(c); i1++ {
		vertices := []Coord{}
		for i2 := 0; i2 < len(c); i2++ {
			if i2 == i1 {
				continue
			}
			vertex, found := c.vertex(i1, i2, norms, epsilon)
			if found {
				vertices = append(vertices, vertex)
			}
		}
		if len(vertices) > 1 {
			addConvexFace(m, vertices, c[i1].Normal)
		}
	}

	// Sometimes more than three planes intersect, in
	// which case a bunch of nearly duplicate faces are
	// created.
	m = m.Repair(epsilon)

	// Remove zero-length segments.
	m.Iterate(func(s *Segment) {
		if s[0] == s[1] {
			m.Remove(s)
		}
	})

	return m
}

// Solid creates a solid out of the polytope.
//
// This runs in O(n^2) in the constraints, so it may be
// unacceptable for large polytopes.
func (c ConvexPolytope) Solid() Solid {
	m := c.Mesh()
	return &polytopeSolid{
		P:      c,
		MinVal: m.Min(),
		MaxVal: m.Max(),
	}
}

func (c ConvexPolytope) vertex(i1, i2 int, norms []float64, epsilon float64) (Coord, bool) {
	// Make sure the indices are sorted so that we yield
	// deterministic results for different first faces.
	if i2 < i1 {
		return c.vertex(i2, i1, norms, epsilon)
	}

	l1, l2 := c[i1], c[i2]
	matrix := Matrix2{l1.Normal.X, l1.Normal.Y, l2.Normal.X, l2.Normal.Y}

	// Check for singular (or poorly conditioned) matrix.
	rawArea := l1.Normal.Norm() * l2.Normal.Norm()
	det := matrix.Det()
	if math.Abs(det) < rawArea*1e-8 {
		return Coord{}, false
	}

	maxes := Coord{l1.Max, l2.Max}
	solution := matrix.MulColumnInv(maxes, det)

	for i, l := range c {
		if i == i1 || i == i2 {
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

func addConvexFace(m *Mesh, vertices []Coord, normal Coord) {
	direction := XY(-normal.Y, normal.X)
	dots := make([]float64, len(vertices))
	for i, x := range vertices {
		dots[i] = x.Dot(direction)
	}
	essentials.VoodooSort(dots, func(i, j int) bool {
		return dots[i] < dots[j]
	}, vertices)
	for i := 1; i < len(vertices); i++ {
		seg := &Segment{vertices[i-1], vertices[i]}
		if seg.Normal().Dot(normal) < 0 {
			seg[0], seg[1] = seg[1], seg[0]
		}
		m.Add(seg)
	}
}

type polytopeSolid struct {
	P      ConvexPolytope
	MinVal Coord
	MaxVal Coord
}

func (p *polytopeSolid) Min() Coord {
	return p.MinVal
}

func (p *polytopeSolid) Max() Coord {
	return p.MaxVal
}

func (p *polytopeSolid) Contains(c Coord) bool {
	return InBounds(p, c) && p.P.Contains(c)
}
