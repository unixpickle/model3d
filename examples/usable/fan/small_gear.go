package main

import (
	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func SmallGearMesh() *model3d.Mesh {
	solid := SmallGearSolid()
	mesh := model3d.MarchingCubesSearch(solid, 0.005, 8)
	return mesh
}

func SmallGearSolid() model3d.Solid {
	return model3d.StackedSolid{
		&toolbox3d.HelicalGear{
			P2: model3d.Z(GearThickness),
			Profile: toolbox3d.InvoluteGearProfileSizes(GearPressureAngle, GearModule,
				GearAddendum, GearDedendum, SmallGearTeeth),
			Angle: -GearHelicalAngle,
		},
		&model3d.Cylinder{
			P2:     model3d.Coord3D{Z: SpineThickness + SpineWasherSize + PoleExtraLength},
			Radius: PoleRadius,
		},
		&toolbox3d.ScrewSolid{
			P2:         model3d.Z(BladeDepth),
			Radius:     ScrewRadius - ScrewSlack,
			GrooveSize: ScrewGrooveSize,
		},
	}
}
