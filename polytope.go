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
	epsilon := c.spacialEpsilon()
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
				vertex, found := c.vertex(i1, i2, i3, epsilon)
				if found {
					vertices = append(vertices, vertex)
				}
			}
		}
		if len(vertices) > 2 {
			addConvexFace(m, vertices, c[i1].Normal)
		}
	}
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

func (c ConvexPolytope) vertex(i1, i2, i3 int, epsilon float64) (Coord3D, bool) {
	// Make sure the indices are sorted so that we yield
	// deterministic results for different first faces.
	if i2 < i1 {
		return c.vertex(i2, i1, i3, epsilon)
	} else if i3 < i1 {
		return c.vertex(i3, i2, i1, epsilon)
	} else if i3 < i2 {
		return c.vertex(i1, i3, i2, epsilon)
	}

	l1, l2, l3 := c[i1], c[i2], c[i3]
	matrix := NewMatrix3Columns(l1.Normal, l2.Normal, l3.Normal).Transpose()

	// Check for singular (or poorly conditioned) matrix.
	rawArea := l1.Normal.Norm() * l2.Normal.Norm() * l3.Normal.Norm()
	if math.Abs(matrix.Det()) < rawArea*1e-8 {
		return Coord3D{}, false
	}

	maxes := Coord3D{l1.Max, l2.Max, l3.Max}
	solution := matrix.Inverse().MulColumn(maxes)

	for i, l := range c {
		if i == i1 || i == i2 || i == i3 {
			continue
		}
		if l.Normal.Dot(solution) > l.Max+epsilon/l.Normal.Norm() {
			return solution, false
		}
	}

	return solution, true
}

func (c ConvexPolytope) spacialEpsilon() float64 {
	var min, max Coord3D
	var first bool
	for i := 0; i < len(c)-2; i++ {
		for j := i + 1; j < len(c)-1; j++ {
			for k := j + 1; k < len(c); k++ {
				coord, found := c.vertex(i, j, k, math.Inf(1))
				if found {
					if first {
						min = coord
						max = coord
						first = false
					} else {
						min = min.Min(coord)
						max = max.Max(coord)
					}
				}
			}
		}
	}
	return max.Sub(min).Norm() * 1e-8
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
		angles[i] = math.Atan2(y, x)
	}

	essentials.VoodooSort(angles, func(i, j int) bool {
		return angles[i] < angles[j]
	}, vertices)

	for i := 0; i < len(vertices); i++ {
		t := &Triangle{vertices[i], vertices[(i+1)%len(vertices)], center}
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
