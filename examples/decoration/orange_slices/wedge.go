package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	WedgeThickness = PeelLongSide - 0.03
)

type Wedge struct {
	faces model3d.FaceSDF
}

func NewWedge(x1, x2 float64) *Wedge {
	p := PeelCentralCurve()
	p1, p2 := p(x1), p(x2)
	edge := model3d.NewSegment(p1, p2)

	tris := model3d.NewMesh()
	delta := 0.01
	for x := x1; x < x2-delta; x += delta {
		p1, p2 = p(x), p(x+delta)
		proj1, proj2 := edge.Closest(p1), edge.Closest(p2)
		tris.AddQuad(p1, p2, proj2, proj1)
	}

	// Remove singular triangles at the endpoints.
	tris.Iterate(func(t *model3d.Triangle) {
		if t.Area() < 1e-8 {
			tris.Remove(t)
		}
	})

	return &Wedge{faces: model3d.MeshToSDF(tris)}
}

func (w *Wedge) Min() model3d.Coord3D {
	min := w.faces.Min()
	return min.Sub(model3d.Ones(WedgeThickness))
}

func (w *Wedge) Max() model3d.Coord3D {
	min := w.faces.Max()
	return min.Add(model3d.Ones(WedgeThickness))
}

func (w *Wedge) Contains(c model3d.Coord3D) bool {
	tri, neighbor, dist := w.faces.FaceSDF(c)

	// Project c onto the triangle plane, and make sure it
	// is within the triangle.
	n := tri.Normal()
	proj := c.ProjectOut(n).Add(n.Scale(n.Dot(tri[0])))
	if neighbor.Dist(proj) > 1e-5 {
		return false
	}

	return math.Abs(dist) < WedgeThickness/2
}
