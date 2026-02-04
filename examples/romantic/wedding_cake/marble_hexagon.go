package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
	. "github.com/unixpickle/model3d/shorthand"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	HexThickness      = 1.0
	GoldDripThickness = 0.035
)

var GoldDripColor = RGB(1.0, 0.73, 0.04)

func MarbleHexagonLayer() (Solid3, toolbox3d.CoordColorFunc) {
	hexMesh := model2d.NewMeshPolar(func(theta float64) float64 {
		return 1.0
	}, 6)
	solid2d := hexMesh.Solid()
	solid3d := model3d.ProfileSolid(solid2d, 0.0, HexThickness)

	streaks := CreateMarbleStreaks()

	dripColl := model3d.MeshToCollider(GoldDripMesh())
	drip := model3d.NewColliderSolidHollow(dripColl, GoldDripThickness)
	rotDrip := Rotate3(drip, Z(1), math.Pi)

	return Join3(solid3d, drip, rotDrip), func(c C3) Color {
		if drip.Contains(c) || rotDrip.Contains(c) {
			return GoldDripColor
		}
		maxStreak := 0.0
		for _, s := range streaks {
			maxStreak = math.Max(maxStreak, s.Evaluate(c))
		}
		return Gray(1.0).AddScalar(-maxStreak * 0.5)
	}
}

type MarbleStreak struct {
	Theta     float64
	Power     float64
	Darkness  float64
	Thickness model2d.BezierCurve
	Curve     model2d.BezierCurve
}

func CreateMarbleStreaks() []*MarbleStreak {
	return []*MarbleStreak{
		{
			Theta:    0.0,
			Power:    3.0,
			Darkness: 0.28,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.05),
				XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.5, 0.3),
				XY(1.0, -0.3),
			},
		},
		{
			Theta:    math.Pi * 0.7,
			Power:    2.0,
			Darkness: 0.34,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.02),
				XY(1.0, 0.08),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.25, 0.1),
				XY(0.6, -0.3),
				XY(1.0, -0.1),
			},
		},
		{
			Theta:    math.Pi * 1.5,
			Power:    2.0,
			Darkness: 0.25,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.05),
				XY(0.25, 0.04),
				XY(1.0, 0.045),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.2, 0.1),
				XY(0.5, -0.1),
				XY(0.7, -0.25),
				XY(1.0, -0.3),
			},
		},
		{
			Theta:    math.Pi * 0.3,
			Power:    3.0,
			Darkness: 0.15,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.08),
				XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.3, 0.1),
				XY(0.6, -0.1),
				XY(1.0, 0.2),
			},
		},
		{
			Theta:    math.Pi * 1.2,
			Power:    3.0,
			Darkness: 0.3,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.08),
				XY(0.8, 0.05),
				XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.3, -0.2),
				XY(0.6, -0.4),
				XY(1.0, -0.5),
			},
		},
		{
			Theta:    math.Pi * 1.8,
			Power:    2.0,
			Darkness: 0.4,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.08),
				XY(0.5, 0.09),
				XY(1.0, 0.05),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.3, 0.1),
				XY(0.6, -0.1),
				XY(1.0, 0.2),
			},
		},
		{
			Theta:    math.Pi * 0.5,
			Power:    2.0,
			Darkness: 0.2,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.1),
				XY(0.5, 0.11),
				XY(1.0, 0.08),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.2, 0.1),
				XY(0.5, -0.2),
				XY(1.0, -0.3),
			},
		},
		{
			Theta:    math.Pi * 1.2,
			Power:    2.0,
			Darkness: 0.2,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.1),
				XY(1.0, 0.08),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.5, 0.05),
				XY(1.0, 0.2),
			},
		},
		{
			Theta:    math.Pi * 0.6,
			Power:    2.5,
			Darkness: 0.2,
			Thickness: model2d.BezierCurve{
				XY(0.0, 0.05),
				XY(0.5, 0.06),
				XY(1.0, 0.06),
			},
			Curve: model2d.BezierCurve{
				XY(0.0, 0.0),
				XY(0.5, 0.05),
				XY(1.0, -0.2),
			},
		},
	}
}

func (m *MarbleStreak) Evaluate(c C3) float64 {
	xAxis := XYZ(-math.Sin(m.Theta), math.Cos(m.Theta), 0)
	xDist := c.Dot(xAxis)
	zFrac := math.Max(0.0, math.Min(1.0, c.Z/HexThickness))
	offset := m.Curve.EvalX(zFrac)
	thickness := m.Thickness.EvalX(zFrac)
	distFrac := math.Abs(offset-xDist) / thickness
	if distFrac > 1 {
		return 0.0
	}
	return m.Darkness * (1 - math.Pow(distFrac, m.Power))
}

func GoldDripMesh() *model3d.Mesh {
	topHexMesh := model2d.NewMeshPolar(func(theta float64) float64 {
		return 1.0
	}, 6)
	topHexCollider := model2d.MeshToCollider(topHexMesh)
	radiusForTheta := func(theta float64) float64 {
		d := XY(math.Cos(theta), math.Sin(theta))
		rc, _ := topHexCollider.FirstRayCollision(&model2d.Ray{
			Origin:    model2d.Origin,
			Direction: d,
		})
		return rc.Scale
	}
	upperPoint := func(theta, radius float64) C3 {
		return XYZ(math.Cos(theta)*radius, math.Sin(theta)*radius, HexThickness)
	}

	theta0 := 0.7
	theta1 := 1.9

	topShape := model2d.BezierCurve{
		XY(theta0, radiusForTheta(theta0)),
		XY(theta0, radiusForTheta(theta0)-0.2),
		XY(theta0*0.7+theta1*0.3, radiusForTheta(theta0)-0.5),
		XY(theta0*0.2+theta1*0.8, radiusForTheta(theta0)-0.3),
		XY(theta1, radiusForTheta(theta1)-0.2),
		XY(theta1, radiusForTheta(theta1)),
	}
	bottomShape := model2d.BezierCurve{
		XY(theta0, 0.0),
		XY(theta0, 0.4),
		XY(theta1*0.6+theta0*0.4, -0.2),
		XY(theta1*0.2+theta0*0.8, -0.3),
		XY(theta1, 0.6),
		XY(theta1, 0.0),
	}
	_ = bottomShape

	delta := 0.01
	mesh := model3d.NewMesh()
	for theta := theta0; theta < theta1; theta += delta {
		next := math.Min(theta+delta, theta1)
		r0 := radiusForTheta(theta)
		r1 := topShape.EvalX(theta)
		r2 := topShape.EvalX(next)
		r3 := radiusForTheta(next)
		mesh.AddQuad(
			upperPoint(theta, r0),
			upperPoint(theta, r1),
			upperPoint(next, r2),
			upperPoint(next, r3),
		)

		p0 := upperPoint(theta, r0)
		p1 := p0.Add(Z(-bottomShape.EvalX(theta)))
		p3 := upperPoint(next, r3)
		p2 := p3.Add(Z(-bottomShape.EvalX(next)))
		mesh.AddQuad(p0, p1, p2, p3)
	}

	dropXAndLen := [][2]float64{
		{0.05, 0.32},
		{0.2, 0.3},
		{0.31, 0.25},
		{0.4, 0.35},
		{0.52, 0.25},
		{0.61, 0.3},
		{0.69, 0.33},
		{0.8, 0.4},
		{0.93, 0.3},
	}
	for _, dropInfo := range dropXAndLen {
		theta := dropInfo[0]*(theta1-theta0) + theta0
		length := dropInfo[1]
		p1 := upperPoint(theta, radiusForTheta(theta))
		p2 := p1.Add(Z(-length))
		mesh.Add(&model3d.Triangle{p1, p2, p2})
	}

	return mesh
}
