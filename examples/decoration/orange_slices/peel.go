package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

const (
	PeelLongSide  = 0.12
	PeelSmallSide = 0.05
	PeelStops     = 800
	PeelRounding  = 0.01
)

func NewPeel() model3d.Solid {
	mesh := PeelMesh(PeelStops)
	return model3d.NewColliderSolidInset(model3d.MeshToCollider(mesh), -PeelRounding)
}

// PeelMesh creates a mesh for the peel.
func PeelMesh(stops int) *model3d.Mesh {
	curve := PeelCentralCurve()
	twist := PeelTwist()
	centralDir := func(x float64) model3d.Coord3D {
		delta := 0.001
		if x > 0 {
			x -= delta
		}
		return curve(x + delta).Sub(curve(x)).Normalize()
	}
	corners := func(t int) [4]model3d.Coord3D {
		x := float64(t)/float64(stops)*2 - 1
		c := curve(x)
		dir := centralDir(x)

		theta := model2d.CurveEvalX(twist, x)
		rotation := model3d.Rotation(dir, theta)

		longDir := model3d.Z(1.0)
		shortDir := longDir.Cross(dir).Normalize()
		longDir = rotation.Apply(longDir).Scale(PeelLongSide / 2.0)
		shortDir = rotation.Apply(shortDir).Scale(PeelSmallSide / 2.0)

		scales := [4][2]float64{{1, 1}, {1, -1}, {-1, -1}, {-1, 1}}
		res := [4]model3d.Coord3D{}
		for i, scale := range scales {
			res[i] = c.Add(shortDir.Scale(scale[0])).Add(longDir.Scale(scale[1]))
		}
		return res
	}

	res := model3d.NewMesh()
	for t := 0; t < stops; t++ {
		corners1 := corners(t)
		corners2 := corners(t + 1)
		for i := 0; i < 4; i++ {
			res.AddQuad(corners1[(i+1)%4], corners1[i], corners2[i], corners2[(i+1)%4])
		}
	}
	for _, t := range [2]int{0, stops} {
		corners := corners(t)
		if t == 0 {
			res.AddQuad(corners[0], corners[1], corners[2], corners[3])
		} else {
			res.AddQuad(corners[1], corners[0], corners[3], corners[2])
		}
	}

	return res
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
			model2d.XY(-0.05, 0.0),
			model2d.XY(0.05, math.Pi),
			model2d.XY(0.2, math.Pi),
		},
		model2d.BezierCurve{
			model2d.XY(0.2, math.Pi),
			model2d.XY(1.0, math.Pi),
		},
	}
}
