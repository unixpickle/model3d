package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateFrame() model3d.Solid {
	var solid model3d.JoinedSolid

	// Side walls.
	for _, x := range []float64{-FrameThickness, DrawerWidth} {
		solid = append(solid, &model3d.Rect{
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
		solid = append(solid, &model3d.Rect{
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
	solid = append(solid, model3d.Subtract(model3d.NewRect(wallMin, solid.Max()), BackWallHole{}))

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

	// Frame holes.
	var holes model3d.JoinedSolid
	for i := 0; i < DrawerCount-1; i++ {
		holes = append(holes, &FrameHole{Z: float64(i+1) * DrawerHeight})
	}
	holes = append(holes, BottomFrameHole{})

	rotate := &model3d.Matrix3Transform{
		Matrix: &model3d.Matrix3{
			1, 0, 0,
			0, 0, -1,
			0, -1, 0,
		},
	}
	return model3d.TransformSolid(rotate, model3d.Subtract(solid, holes))
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
	return model3d.XYZ(math.Min(r.X1, r.X2), 0, r.Z-RidgeDepth)
}

func (r *RidgeSolid) Max() model3d.Coord3D {
	return model3d.XYZ(math.Max(r.X1, r.X2), DrawerDepth+FrameThickness, r.Z+RidgeDepth)
}

func (r *RidgeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(r, c) {
		return false
	}
	return math.Abs(c.Z-r.Z) <= math.Abs(c.X-r.X2)
}

func CreateFoot(x, y float64) model3d.Solid {
	center := model3d.XYZ(x, y, -FrameThickness)
	halfSize := model3d.Coord3D{X: FrameFootWidth / 2, Y: FrameFootWidth / 2}
	return &toolbox3d.Ramp{
		Solid: &model3d.Rect{
			MinVal: center.Sub(halfSize).Sub(model3d.Z(FrameFootHeight)),
			MaxVal: center.Add(halfSize),
		},
		P1: center.Sub(model3d.Z(FrameFootRampHeight)),
		P2: center,
	}
}

type FrameHole struct {
	Z float64
}

func (f *FrameHole) Min() model3d.Coord3D {
	return model3d.Coord3D{
		X: -FrameThickness - 1e-5,
		Y: FrameHoleMargin,
		Z: f.Z - FrameHoleWidth/2,
	}
}

func (f *FrameHole) Max() model3d.Coord3D {
	return model3d.Coord3D{
		X: DrawerWidth + FrameThickness + 1e-5,
		Y: DrawerDepth + FrameThickness - FrameHoleMargin,
		Z: f.Z + FrameHoleWidth/2,
	}
}

func (f *FrameHole) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(f, c) {
		return false
	}
	edgeDist := math.Min(math.Min(c.Y-f.Min().Y, f.Max().Y-c.Y), FrameHoleWidth/2)
	return math.Abs(c.Z-f.Z) <= edgeDist
}

type BottomFrameHole struct{}

func (b BottomFrameHole) Min() model3d.Coord3D {
	return model3d.Coord3D{
		X: (DrawerWidth - FrameBottomHoleRadius) / 2,
		Y: (DrawerDepth - FrameBottomHoleRadius) / 2,
		Z: -(FrameThickness + 1e-5),
	}
}

func (b BottomFrameHole) Max() model3d.Coord3D {
	return model3d.Coord3D{
		X: (DrawerWidth + FrameBottomHoleRadius) / 2,
		Y: (DrawerDepth + FrameBottomHoleRadius) / 2,
		Z: 1e-5,
	}
}

func (b BottomFrameHole) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	mid := b.Min().Mid(b.Max())
	return math.Abs(c.X-mid.X) <= c.Y-b.Min().Y
}

type BackWallHole struct{}

func (b BackWallHole) Min() model3d.Coord3D {
	return model3d.XYZ(-FrameThickness, DrawerDepth-1e-5, -FrameThickness)
}

func (b BackWallHole) Max() model3d.Coord3D {
	return model3d.Coord3D{X: FrameThickness + DrawerWidth, Y: DrawerDepth + FrameThickness + 1e-5,
		Z: DrawerHeight*DrawerCount + FrameThickness}
}

func (b BackWallHole) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(b, c) {
		return false
	}
	minDiff := c.Sub(b.Min())
	maxDiff := b.Max().Sub(c)
	for _, f := range [4]float64{minDiff.X, minDiff.Z, maxDiff.X, maxDiff.Z} {
		if f < FrameBackHoleMargin {
			return false
		}
	}

	// Cool wavy pattern.
	zWave := (math.Sin(c.X*math.Pi*2) + 1) * 0.1
	if math.Mod(c.Z+zWave, 1.0) < 0.2 {
		return false
	}
	xWave := (math.Sin(c.Z*math.Pi*2) + 1) * 0.1
	if math.Mod(c.X+xWave, 1.0) < 0.2 {
		return false
	}

	return true
}
