package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	PeelLongSide  = 0.12
	PeelSmallSide = 0.05
)

type Peel struct {
	twist        model2d.Curve
	centralCurve model3d.PointSDF
	centralDir   func(x float64) model3d.Coord3D
}

func NewPeel() *Peel {
	p := PeelCentralCurve()
	centralTris := model3d.NewMesh()
	delta := 0.01
	for x := -1.0; x+delta < 1.0; x += delta {
		p1, p2 := p(x), p(x+delta)
		centralTris.Add(&model3d.Triangle{p1, p1.Mid(p2), p2})
	}
	return &Peel{
		twist:        PeelTwist(),
		centralCurve: model3d.MeshToSDF(centralTris),
		centralDir: func(x float64) model3d.Coord3D {
			delta := 0.001
			if x > 0 {
				x -= delta
			}
			return p(x + delta).Sub(p(x)).Normalize()

		},
	}
}

func (p *Peel) Min() model3d.Coord3D {
	return model3d.XYZ(-1.1, -0.5, -0.5)
}

func (p *Peel) Max() model3d.Coord3D {
	return model3d.XYZ(1.1, 0.5, 0.2)
}

func (p *Peel) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	nearest, _ := p.centralCurve.PointSDF(c)
	centralDir := p.centralDir(nearest.X)
	diff := c.Sub(nearest)

	// Check if we are past an endpoint.
	if math.Abs(nearest.Add(centralDir.Scale(centralDir.Dot(diff))).X) > 1 {
		return false
	}

	theta := model2d.CurveEvalX(p.twist, nearest.X)
	rotation := model3d.Rotation(centralDir, theta)

	longDir := model3d.Z(1.0)
	shortDir := longDir.Cross(centralDir).Normalize()
	longDir = rotation.Apply(longDir)
	shortDir = rotation.Apply(shortDir)

	return math.Abs(diff.Dot(longDir)) < PeelLongSide/2 && math.Abs(diff.Dot(shortDir)) < PeelSmallSide/2
}

// PeelCentralCurve gets the curve of the peel's center,
// as a parameter of x.
func PeelCentralCurve() func(x float64) model3d.Coord3D {
	planeCurve := PeelCurve()
	zCurve := PeelHeight()
	return func(x float64) model3d.Coord3D {
		return model3d.XYZ(x, model2d.CurveEvalX(planeCurve, x), model2d.CurveEvalX(zCurve, x))
	}
}

// PeelCurve creates a 2D projection of the overall peel
// on the plane.
func PeelCurve() model2d.Curve {
	return model2d.JoinedCurve{
		model2d.BezierCurve{
			model2d.XY(-1.0, 0.0),
			model2d.XY(-0.6, 0.6),
			model2d.XY(0.0, 0.0),
		},
		model2d.BezierCurve{
			model2d.XY(0.0, 0.0),
			model2d.XY(0.6, -0.6),
			model2d.XY(1.0, 0.0),
		},
	}
}

// PeelHeight creates the z-component of the peel on the
// plane, where x maps to z.
func PeelHeight() model2d.Curve {
	return model2d.JoinedCurve{
		model2d.BezierCurve{
			model2d.XY(-1.0, -0.3),
			model2d.XY(-0.2, -0.06),
		},
		model2d.BezierCurve{
			model2d.XY(-0.2, -0.06),
			model2d.XY(0.0, 0.0),
			model2d.XY(0.2, -0.06),
		},
		model2d.BezierCurve{
			model2d.XY(0.2, -0.06),
			model2d.XY(1.0, -0.3),
		},
	}
}

// PeelTwist creates a function of theta with respect to x.
func PeelTwist() model2d.Curve {
	return model2d.JoinedCurve{
		model2d.BezierCurve{
			model2d.XY(-1.0, 0.0),
			model2d.XY(-0.2, 0.0),
		},
		model2d.BezierCurve{
			model2d.XY(-0.2, 0.0),
			model2d.XY(-0.19, 0.0),
			model2d.XY(0.0, math.Pi/2),
			model2d.XY(0.19, math.Pi),
			model2d.XY(0.2, math.Pi),
		},
		model2d.BezierCurve{
			model2d.XY(0.2, math.Pi),
			model2d.XY(1.0, math.Pi),
		},
	}
}
