package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateContainer() model3d.Solid {
	var solid model3d.JoinedSolid

	// Side walls.
	for _, x := range []float64{-ContainerThickness, ShelfWidth} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: x, Z: -ContainerThickness},
			MaxVal: model3d.Coord3D{
				X: x + ContainerThickness,
				Y: ShelfDepth + ContainerThickness,
				Z: ShelfCount*ShelfHeight + ContainerThickness,
			},
		})
	}

	// Top/bottom walls.
	for _, z := range []float64{-ContainerThickness, ShelfHeight * ShelfCount} {
		solid = append(solid, &model3d.RectSolid{
			MinVal: model3d.Coord3D{X: -ContainerThickness, Z: z},
			MaxVal: model3d.Coord3D{
				X: ShelfWidth + ContainerThickness,
				Y: ShelfDepth + ContainerThickness,
				Z: z + ContainerThickness,
			},
		})
	}

	// Back wall.
	wallMin := solid.Min()
	wallMin.Y = solid.Max().Y - ContainerThickness
	solid = append(solid, &model3d.RectSolid{
		MinVal: wallMin,
		MaxVal: solid.Max(),
	})

	// Ridges for shelves.
	for i := 0; i < ShelfCount; i++ {
		for _, right := range []bool{false, true} {
			solid = append(solid, CreateRidge((float64(i)+0.5)*ShelfHeight, right))
		}
	}

	footXs := []float64{solid.Min().X + ContainerFootWidth/2, solid.Max().X - ContainerFootWidth/2}
	footYs := []float64{solid.Min().Y + ContainerFootWidth/2, solid.Max().Y - ContainerFootWidth/2}
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
		return &RidgeSolid{X1: ShelfWidth, X2: ShelfWidth - RidgeDepth, Z: z}
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
	return model3d.Coord3D{X: math.Max(r.X1, r.X2), Y: ShelfDepth + ContainerThickness,
		Z: r.Z + RidgeDepth}
}

func (r *RidgeSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(r, c) {
		return false
	}
	return math.Abs(c.Z-r.Z) <= math.Abs(c.X-r.X2)
}

func CreateFoot(x, y float64) model3d.Solid {
	center := model3d.Coord3D{X: x, Y: y, Z: -ContainerThickness}
	halfSize := model3d.Coord3D{X: ContainerFootWidth / 2, Y: ContainerFootWidth / 2}
	return &toolbox3d.Ramp{
		Solid: &model3d.RectSolid{
			MinVal: center.Sub(halfSize).Sub(model3d.Coord3D{Z: ContainerFootHeight}),
			MaxVal: center.Add(halfSize),
		},
		P1: center.Sub(model3d.Coord3D{Z: ContainerFootRampHeight}),
		P2: center,
	}
}
