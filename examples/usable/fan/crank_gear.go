package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CrankGearMesh() *model3d.Mesh {
	solid := CrankGearSolid()
	mesh := model3d.SolidToMesh(solid, 0.01, 0, -1, 5)
	return mesh
}

func CrankGearSolid() model3d.Solid {
	handlePoint := model3d.Coord3D{X: LargeGearRadius - CrankHandleRadius*1.5, Z: GearThickness}
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&toolbox3d.HelicalGear{
				P2: model3d.Coord3D{Z: GearThickness},
				Profile: toolbox3d.InvoluteGearProfileSizes(GearPressureAngle, GearModule,
					GearAddendum, GearDedendum, LargeGearTeeth),
				Angle: GearHelicalAngle,
			},
			&model3d.CylinderSolid{
				P1:     handlePoint,
				P2:     handlePoint.Add(model3d.Coord3D{Z: CrankHandleLength}),
				Radius: CrankHandleRadius,
			},
		},
		Negative: model3d.JoinedSolid{
			CrankGearHollow{},
			&toolbox3d.ScrewSolid{
				P2:         model3d.Coord3D{Z: GearThickness},
				Radius:     ScrewRadius,
				GrooveSize: ScrewGrooveSize,
			},
		},
	}
}

type CrankGearHollow struct{}

func (c CrankGearHollow) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -LargeGearRadius, Y: -LargeGearRadius}
}

func (c CrankGearHollow) Max() model3d.Coord3D {
	return c.Min().Scale(-1).Add(model3d.Coord3D{Z: GearThickness})
}

func (c CrankGearHollow) Contains(coord model3d.Coord3D) bool {
	if !model3d.InBounds(c, coord) {
		return false
	}
	c2 := coord.Coord2D()
	rad := c2.Norm()
	if rad < CrankGearCenterRadius || rad > LargeGearRadius-CrankGearRimSize {
		return false
	}

	theta := math.Atan2(c2.Y, c2.X) + math.Pi*2
	bound := CrankGearPoleSize / rad
	_, modulo := math.Modf(1.5 + theta*CrankGearSections/(math.Pi*2))
	return math.Abs(modulo-0.5) > bound/2
}
