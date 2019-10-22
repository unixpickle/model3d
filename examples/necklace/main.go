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

	ClaspRingRadius  = 0.4
	ClaspBarLength   = 1.2
	ClaspBarDistance = 0.2

	NumRows = 7
	NumCols = 5
)

func main() {
	mesh := model3d.NewMesh()
	center := model3d.Coord3D{}
	direction := 1.0
	for i := 0; i < NumRows; i++ {
		endCenter := center
		endCenter.Y += direction * (RingOuterRadius) * (NumCols*2 + 1)
		if i == NumRows/2 {
			AddMiddleRow(mesh, center, direction)
		} else {
			AddNormalRow(mesh, center, direction, i)
		}
		center = endCenter
		if i+1 < NumRows {
			center.X += RingOuterRadius + RingSpacing
			AddRing(mesh, center, 1)
			center.X += RingOuterRadius + RingSpacing
			direction *= -1
		}
	}

	ioutil.WriteFile("model.stl", mesh.EncodeSTL(), 0755)
}

func AddMiddleRow(m *model3d.Mesh, center model3d.Coord3D, direction float64) {
	// TODO: custom jewelry here.
	AddNormalRow(m, center, direction, 1)
}

func AddNormalRow(m *model3d.Mesh, center model3d.Coord3D, direction float64, i int) {
	for j := 0; j < NumCols; j++ {
		if i == 0 && j == 0 {
			offset := (ClaspRingRadius - RingOuterRadius) / math.Sqrt2
			AddClaspRing(m, center.Sub(model3d.Coord3D{X: offset, Y: offset}), 2)
		} else if i+1 == NumRows && j+1 == NumCols {
			AddBarRing(m, center)
		} else {
			AddRing(m, center, 2)
		}
		if j+1 < NumCols {
			center.Y += direction * (RingOuterRadius + RingSpacing)
			AddRing(m, center, 0)
			center.Y += direction * (RingOuterRadius + RingSpacing)
		}
	}
}

func AddBarRing(m *model3d.Mesh, center model3d.Coord3D) {
	barLeft := center
	barLeft.Y -= RingOuterRadius + ClaspBarDistance
	barLeft.X -= ClaspBarLength / 2
	solid := model3d.JoinedSolid{
		&model3d.TorusSolid{
			Axis:        model3d.Coord3D{Z: 1},
			Center:      center,
			InnerRadius: RingInnerRadius,
			OuterRadius: RingOuterRadius,
		},
		&model3d.CylinderSolid{
			P1:     center.Sub(model3d.Coord3D{Y: RingOuterRadius}),
			P2:     center.Sub(model3d.Coord3D{Y: RingOuterRadius + ClaspBarDistance}),
			Radius: RingInnerRadius,
		},
		&model3d.CylinderSolid{
			P1:     barLeft,
			P2:     barLeft.Add(model3d.Coord3D{X: ClaspBarLength}),
			Radius: RingInnerRadius,
		},
	}
	m.AddMesh(model3d.SolidToMesh(solid, 0.01, 1, 0.8, 5))
}

func AddRing(m *model3d.Mesh, center model3d.Coord3D, normalDim int) {
	addRing(m, center, normalDim, RingOuterRadius)
}

func AddClaspRing(m *model3d.Mesh, center model3d.Coord3D, normalDim int) {
	addRing(m, center, normalDim, ClaspRingRadius)
}

func addRing(m *model3d.Mesh, center model3d.Coord3D, normalDim int, outerRadius float64) {
	torusPoint := func(outer, inner float64) model3d.Coord3D {
		outerDirection := model3d.Coord3D{
			X: math.Cos(outer),
			Y: math.Sin(outer),
		}
		d1 := outerDirection
		d2 := model3d.Coord3D{Z: 1}

		p := outerDirection.Scale(outerRadius)
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
	innerAngles := circleAngles(20.0)
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
