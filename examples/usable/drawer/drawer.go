package main

import "github.com/unixpickle/model3d"

func CreateDrawer() model3d.Solid {
	min := model3d.Coord3D{
		X: DrawerSlack,
		Y: 0,
		Z: FrameThickness + DrawerSlack,
	}
	max := model3d.Coord3D{
		X: DrawerWidth - DrawerSlack,
		Y: DrawerDepth - DrawerSlack,
		Z: FrameThickness + DrawerHeight - DrawerSlack,
	}

	result := model3d.JoinedSolid{
		// Bottom face.
		&model3d.RectSolid{
			MinVal: min,
			MaxVal: model3d.Coord3D{X: max.X, Y: max.Y, Z: min.Z + DrawerBottom},
		},
	}

	// Side faces.
	for _, x := range []float64{min.X, max.X - DrawerThickness} {
		result = append(result, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Y: min.Y, Z: min.Z},
			MaxVal: model3d.Coord3D{X: x + DrawerThickness, Y: max.Y, Z: max.Z},
		})
	}

	// Front/back faces.
	for _, y := range []float64{min.Y, max.Y - DrawerThickness} {
		result = append(result, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: min.X, Y: y, Z: min.Z},
			MaxVal: model3d.Coord3D{X: max.X, Y: y + DrawerThickness, Z: max.Z},
		})
	}

	mid := min.Mid(max)

	return &model3d.SubtractedSolid{
		Positive: result,
		Negative: model3d.JoinedSolid{
			&RidgeSolid{X1: min.X, X2: min.X + RidgeDepth, Z: mid.Z},
			&RidgeSolid{X1: max.X, X2: max.X - RidgeDepth, Z: mid.Z},
			&HoleCutout{
				X:      mid.X,
				Z:      mid.Z,
				Y1:     min.Y - 1e-5,
				Y2:     min.Y + DrawerThickness + 1e-5,
				Radius: DrawerHoleRadius,
			},
		},
	}
}

type HoleCutout struct {
	X float64
	Z float64

	Y1     float64
	Y2     float64
	Radius float64
}

func (h *HoleCutout) Min() model3d.Coord3D {
	return model3d.Coord3D{X: h.X - h.Radius, Y: h.Y1, Z: h.Z - h.Radius}
}

func (h *HoleCutout) Max() model3d.Coord3D {
	return model3d.Coord3D{X: h.X + h.Radius, Y: h.Y2, Z: h.Z + h.Radius*2}
}

func (h *HoleCutout) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(h, c) {
		return false
	}
	c2d := model3d.Coord2D{X: c.X - h.X, Y: c.Z - h.Z}
	if c2d.Norm() <= h.Radius {
		return true
	}
	if c2d.Y < 0 {
		return false
	}
	// Pointed tip to avoid support.
	vec := model3d.Coord2D{X: 1, Y: 1}.Normalize()
	vec1 := vec.Mul(model3d.Coord2D{X: -1, Y: 1})
	return c2d.Dot(vec) <= h.Radius && c2d.Dot(vec1) <= h.Radius
}
