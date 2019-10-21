package main

import (
	"io/ioutil"
	"math"

	"github.com/unixpickle/model3d"
)

const (
	RingOuterRadius = 0.3
	RingInnerRadius = 0.05
	RingSpacing     = 0.12

	NumRows  = 10
	RowCount = 5
)

func main() {
	mesh := model3d.NewMesh()
	center := model3d.Coord3D{}
	direction := 1.0
	for i := 0; i < NumRows; i++ {
		for j := 0; j < RowCount; j++ {
			AddRing(mesh, center, 2)
			if j+1 < RowCount {
				center.Y += direction * (RingOuterRadius + RingSpacing)
				AddRing(mesh, center, 0)
				center.Y += direction * (RingOuterRadius + RingSpacing)
			}
		}
		if i+1 < NumRows {
			center.X += RingOuterRadius + RingSpacing
			AddRing(mesh, center, 1)
			center.X += RingOuterRadius + RingSpacing
			direction *= -1
		}
	}

	ioutil.WriteFile("model.stl", mesh.EncodeSTL(), 0755)
}

func AddRing(m *model3d.Mesh, center model3d.Coord3D, normalDim int) {
	torusPoint := func(outer, inner float64) model3d.Coord3D {
		outerDirection := model3d.Coord3D{
			X: math.Cos(outer),
			Y: math.Sin(outer),
		}
		d1 := outerDirection
		d2 := model3d.Coord3D{Z: 1}

		p := outerDirection.Scale(RingOuterRadius)
		p = p.Add(d1.Scale(RingInnerRadius * math.Cos(inner)))
		p = p.Add(d2.Scale(RingInnerRadius * math.Sin(inner)))

		if normalDim == 0 {
			p.X, p.Z = p.Z, p.X
		} else if normalDim == 1 {
			p.Y, p.Z = p.Z, p.Y
		} else {
			// Fix normal.
			p.X, p.Y = p.Y, p.X
		}
		return p.Add(center)
	}

	outerAngles := circleAngles(50.0)
	innerAngles := circleAngles(10.0)
	for i, outer := range outerAngles {
		outer1 := outerAngles[(i+1)%len(outerAngles)]
		for j, inner := range innerAngles {
			inner1 := innerAngles[(j+1)%len(innerAngles)]
			p1 := torusPoint(outer, inner)
			p2 := torusPoint(outer, inner1)
			p3 := torusPoint(outer1, inner1)
			p4 := torusPoint(outer1, inner)
			m.Add(&model3d.Triangle{p1, p2, p3})
			m.Add(&model3d.Triangle{p1, p3, p4})
		}
	}
}

func circleAngles(stops float64) []float64 {
	var res []float64
	for i := 0.0; i < stops; i++ {
		res = append(res, math.Pi*2*i/stops)
	}
	return res
}
