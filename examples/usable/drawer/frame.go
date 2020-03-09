package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateFrame() model3d.Solid {
	var solid model3d.JoinedSolid

	// Side walls.
	for _, x := range []float64{-FrameThickness, DrawerWidth} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Z: -FrameThickness},
			MaxVal: model3d.Coord3D{
				X: x + FrameThickness,
				Y: DrawerDepth + FrameThickness,
				Z: DrawerCount*DrawerHeight + FrameThickness,
			},
		})
	}

	// Top/bottom walls.
	for _, z := range []float64{-FrameThickness, DrawerHeight * DrawerCount} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -FrameThickness, Z: z},
			MaxVal: model3d.Coord3D{
				X: DrawerWidth + FrameThickness,
				Y: DrawerDepth + FrameThickness,
				Z: z + FrameThickness,
			},
		})
	}

	// Back wall.
	wallMin := solid.Min()
	wallMin.Y = solid.Max().Y - FrameThickness
	solid = append(solid, &model3d.RectSolid{
		MinVal: wallMin,
		MaxVal: solid.Max(),
	})

	// Ridges for shelves.
	for i := 0; i < DrawerCount; i++ {
		for _, right := range []bool{false, true} {
			solid = append(solid, CreateRidge((float64(i)+0.5)*DrawerHeight, right))
		}
	}

	footXs := []float64{solid.Min().X + FrameFootWidth/2, solid.Max().X - FrameFootWidth/2}
	footYs := []float64{solid.Min().Y + FrameFootWidth/2, solid.Max().Y - FrameFootWidth/2}
	for _, x := range footXs {
		for _, y := range footYs {
			solid = append(solid, CreateFoot(x, y))
		}
	}

	return solid
}

func CreateRidge(z float64, onRight bool) model3d.Solid {
	if !onRight {
		return &RidgeSolid{X1: 0, X2: RidgeDepth, Z: z}
	} else {
		return &RidgeSolid{X1: DrawerWidth, X2: DrawerWidth - RidgeDepth, Z: z}
	}
}

type RidgeSolid struct {
	X1 float64
	X2 float64
	Z  float64
}

func (r *RidgeSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: math.Min(r.X1, r.X2), Y: 0, Z: r.Z - RidgeDepth}
}

func (r *RidgeSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: math.Max(r.X1, r.X2), Y: DrawerDepth + FrameThickness,
		Z: r.Z + RidgeDepth}
}

func (r *RidgeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(r, c) {
		return false
	}
	return math.Abs(c.Z-r.Z) <= math.Abs(c.X-r.X2)
}

func CreateFoot(x, y float64) model3d.Solid {
	center := model3d.Coord3D{X: x, Y: y, Z: -FrameThickness}
	halfSize := model3d.Coord3D{X: FrameFootWidth / 2, Y: FrameFootWidth / 2}
	return &toolbox3d.Ramp{
		Solid: &model3d.RectSolid{
			MinVal: center.Sub(halfSize).Sub(model3d.Coord3D{Z: FrameFootHeight}),
			MaxVal: center.Add(halfSize),
		},
		P1: center.Sub(model3d.Coord3D{Z: FrameFootRampHeight}),
		P2: center,
	}
}
