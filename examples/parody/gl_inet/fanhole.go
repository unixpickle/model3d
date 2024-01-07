package main

import (
	"math"

	"github.com/unixpickle/model3d/model2d"
	"github.com/unixpickle/model3d/model3d"
)

func FanHole() model3d.Solid {
	oneHole := model3d.CheckedFuncSolid(
		model3d.XYZ(0.9, 0.7, 0.45),
		model3d.XYZ(BodySideLength/2+0.01, 0.9, 0.65),
		func(c model3d.Coord3D) bool {
			x := (c.Y - 0.8) / 0.1
			y := (c.Z - 0.55) / 0.1
			xy := model2d.XY(x, y)
			if xy.Norm() > 1.0 {
				return false
			}
			if c.X < BodySideLength/2-0.03 {
				r := xy.Norm()
				theta := math.Atan2(y, x) + r
				if theta < 0 {
					theta += math.Pi * 2
				}
				for i := 0; i < 3; i++ {
					if thetaDiff(theta, float64(i)*math.Pi*2/3+0.1) < 0.3/math.Max(0.03, r) {
						return false
					}
				}
			}
			return true
		},
	)
	transformed := model3d.TransformSolid(
		&model3d.Matrix3Transform{
			Matrix: &model3d.Matrix3{-1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0},
		},
		oneHole,
	)
	return model3d.JoinedSolid{oneHole, transformed}
}

func thetaDiff(t1, t2 float64) float64 {
	diff := t2 - t1
	for diff < 0 {
		diff += math.Pi * 2
	}
	for diff > math.Pi*2 {
		diff -= math.Pi * 2
	}
	return diff
}
